package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/javapub/agi-platform-backend/internal/dto"
	"github.com/javapub/agi-platform-backend/internal/model"
	"github.com/javapub/agi-platform-backend/internal/repository"
	"github.com/javapub/agi-platform-backend/internal/textadapter"
	apperrors "github.com/javapub/agi-platform-backend/pkg/errors"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const defaultPromptOptimizationSystemPrompt = `You improve prompts for AI image and video generation. Preserve the user's intent, named entities, language, and explicit constraints. Return only the improved prompt, without commentary, labels, quotation marks, or Markdown. For images, make composition, visual details, lighting, and style concrete. For videos, make subjects, action, camera movement, scene progression, and timing concrete. Do not invent brands, people, or requirements the user did not request.`

type PromptOptimizationService struct {
	configRepo       *repository.ConfigRepository
	modelRepo        *repository.AIModelRepository
	channelModelRepo *repository.ChannelModelRepository
	creditRepo       *repository.CreditRepository
	logRepo          *repository.PromptOptimizationRepository
	db               *gorm.DB
}

func NewPromptOptimizationService(configRepo *repository.ConfigRepository, modelRepo *repository.AIModelRepository, channelModelRepo *repository.ChannelModelRepository, creditRepo *repository.CreditRepository, logRepo *repository.PromptOptimizationRepository, db *gorm.DB) *PromptOptimizationService {
	return &PromptOptimizationService{configRepo: configRepo, modelRepo: modelRepo, channelModelRepo: channelModelRepo, creditRepo: creditRepo, logRepo: logRepo, db: db}
}

func (s *PromptOptimizationService) Optimize(userID int64, req *dto.PromptOptimizationRequest) (*dto.PromptOptimizationResponse, error) {
	config, err := s.configRepo.GetPromptOptimizationConfig()
	if err != nil {
		return nil, err
	}
	if !config.IsActive || strings.TrimSpace(config.ModelName) == "" {
		return nil, apperrors.New(apperrors.ErrCodeBadRequest, "提示词优化暂未启用")
	}
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" || utf8.RuneCountInString(prompt) > config.MaxInputLength {
		return nil, apperrors.New(apperrors.ErrCodeBadRequest, "提示词长度不符合优化配置")
	}
	count, err := s.logRepo.CountUserSince(userID, time.Now().Add(-time.Minute))
	if err != nil {
		return nil, err
	}
	if count >= int64(config.RateLimitPerMinute) {
		return nil, apperrors.New(apperrors.ErrCodeBadRequest, "提示词优化过于频繁，请稍后再试")
	}

	aiModel, err := s.modelRepo.FindByName(config.ModelName)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrCodeBadRequest, "未找到可用的提示词优化模型")
	}
	if aiModel.Type != "text" {
		return nil, apperrors.New(apperrors.ErrCodeBadRequest, "提示词优化模型必须是文本模型")
	}
	binding, err := s.channelModelRepo.SelectActiveChannel(aiModel.ID)
	if err != nil || binding.Channel == nil {
		return nil, apperrors.New(apperrors.ErrCodeBadRequest, "提示词优化模型没有可用渠道")
	}
	if config.CreditCost > 0 {
		account, accountErr := s.creditRepo.GetOrCreateAccount(userID)
		if accountErr != nil {
			return nil, accountErr
		}
		if account.Balance < config.CreditCost {
			return nil, apperrors.ErrInsufficientCredit
		}
	}

	params, _ := json.Marshal(req.Params)
	log := &model.PromptOptimizationLog{UserID: userID, ModelName: aiModel.Name, ChannelID: binding.ChannelID, TargetType: req.TargetType, TargetModelName: req.TargetModelName, Params: datatypes.JSON(params), OriginalPrompt: prompt, CreditCost: config.CreditCost, Status: "processing", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := s.logRepo.Create(log); err != nil {
		return nil, err
	}
	adapter, err := textadapter.GetAdapter(aiModel, binding.Channel)
	if err != nil {
		return nil, s.fail(log, err)
	}
	systemPrompt := strings.TrimSpace(config.SystemPrompt)
	if systemPrompt == "" {
		systemPrompt = defaultPromptOptimizationSystemPrompt
	}
	start := time.Now()
	optimized, err := adapter.Complete(context.Background(), &textadapter.Request{ModelName: aiModel.Name, SystemPrompt: systemPrompt, Prompt: buildOptimizationPrompt(prompt, req)})
	log.LatencyMS = int(time.Since(start).Milliseconds())
	if err != nil {
		return nil, s.fail(log, err)
	}
	optimized = strings.TrimSpace(optimized)
	if optimized == "" {
		return nil, s.fail(log, fmt.Errorf("文本渠道未返回优化结果"))
	}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if config.CreditCost > 0 {
			if err := s.deductCredit(tx, userID, config.CreditCost, log.ID); err != nil {
				return err
			}
		}
		log.OptimizedPrompt, log.Status, log.UpdatedAt = optimized, "success", time.Now()
		return s.logRepo.UpdateTx(tx, log)
	}); err != nil {
		return nil, s.fail(log, err)
	}
	return &dto.PromptOptimizationResponse{Prompt: optimized, ModelName: aiModel.Name, CreditCost: config.CreditCost}, nil
}

func (s *PromptOptimizationService) fail(log *model.PromptOptimizationLog, cause error) error {
	log.Status, log.ErrorMsg, log.UpdatedAt = "failed", truncateRunes(cause.Error(), 1000), time.Now()
	if err := s.logRepo.Update(log); err != nil {
		return err
	}
	return apperrors.New(apperrors.ErrCodeBadRequest, "提示词优化失败，请稍后重试")
}

func (s *PromptOptimizationService) deductCredit(tx *gorm.DB, userID int64, amount int, logID int64) error {
	account, err := s.creditRepo.GetAccountForUpdate(tx, userID)
	if err != nil || account.Balance < amount {
		return apperrors.ErrInsufficientCredit
	}
	account.Balance -= amount
	account.TotalExpense += amount
	account.UpdatedAt = time.Now()
	if err := s.creditRepo.UpdateAccount(tx, account); err != nil {
		return err
	}
	return s.creditRepo.CreateLedger(tx, &model.CreditLedger{UserID: userID, Type: "expense", Amount: amount, Title: "AI 优化提示词", SourceType: "prompt_optimization", SourceID: logID, BalanceAfter: account.Balance, IdempotencyKey: fmt.Sprintf("prompt_optimization_%d", logID), CreatedAt: time.Now()})
}

func (s *PromptOptimizationService) ListAdmin(page, pageSize int) ([]*model.PromptOptimizationLog, int64, error) {
	return s.logRepo.ListAdmin(page, pageSize)
}

func buildOptimizationPrompt(prompt string, req *dto.PromptOptimizationRequest) string {
	params, _ := json.Marshal(req.Params)
	return fmt.Sprintf("Target type: %s\nTarget generation model: %s\nSelected parameters: %s\n\nUser prompt:\n%s", req.TargetType, req.TargetModelName, string(params), prompt)
}

func truncateRunes(value string, max int) string {
	chars := []rune(value)
	if len(chars) <= max {
		return value
	}
	return string(chars[:max])
}
