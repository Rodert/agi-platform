package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	TokenTypeUser  = "user"
	TokenTypeAdmin = "admin"
	APIKeyPrefix   = "agi"
)

var ErrInvalidToken = errors.New("invalid token")

type Manager interface {
	HashPassword(password string) (string, error)
	CheckPassword(hash string, password string) bool
	IssueUserToken(userID uint64) (*Token, error)
	IssueAdminToken(adminID uint64) (*Token, error)
	ParseUserToken(tokenText string) (*Claims, error)
	ParseAdminToken(tokenText string) (*Claims, error)
	NewAPIKey() (plain string, prefix string, hash string, err error)
	HashAPIKey(plain string) string
}

type Config struct {
	JWTSecret     string
	TokenLifetime time.Duration
}

type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

type Claims struct {
	UserID uint64
	Type   string
}

type manager struct {
	secret        []byte
	tokenLifetime time.Duration
}

func NewManager(cfg Config) Manager {
	return &manager{
		secret:        []byte(cfg.JWTSecret),
		tokenLifetime: cfg.TokenLifetime,
	}
}

func (m *manager) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (m *manager) CheckPassword(hash string, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func (m *manager) IssueUserToken(userID uint64) (*Token, error) {
	return m.issueToken(userID, TokenTypeUser)
}

func (m *manager) IssueAdminToken(adminID uint64) (*Token, error) {
	return m.issueToken(adminID, TokenTypeAdmin)
}

func (m *manager) issueToken(subjectID uint64, tokenType string) (*Token, error) {
	now := time.Now()
	expiresAt := now.Add(m.tokenLifetime)

	claims := jwt.MapClaims{
		"sub": strconv.FormatUint(subjectID, 10),
		"typ": tokenType,
		"iat": now.Unix(),
		"exp": expiresAt.Unix(),
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return nil, err
	}

	return &Token{
		AccessToken: signed,
		TokenType:   "Bearer",
		ExpiresIn:   int64(m.tokenLifetime.Seconds()),
	}, nil
}

func (m *manager) ParseUserToken(tokenText string) (*Claims, error) {
	return m.parseToken(tokenText, TokenTypeUser)
}

func (m *manager) ParseAdminToken(tokenText string) (*Claims, error) {
	return m.parseToken(tokenText, TokenTypeAdmin)
}

func (m *manager) parseToken(tokenText string, expectedType string) (*Claims, error) {
	parsed, err := jwt.Parse(tokenText, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})
	if err != nil || !parsed.Valid {
		return nil, ErrInvalidToken
	}

	mapClaims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	tokenType, _ := mapClaims["typ"].(string)
	if tokenType != expectedType {
		return nil, ErrInvalidToken
	}

	sub, _ := mapClaims["sub"].(string)
	userID, err := strconv.ParseUint(sub, 10, 64)
	if err != nil || userID == 0 {
		return nil, ErrInvalidToken
	}

	return &Claims{
		UserID: userID,
		Type:   tokenType,
	}, nil
}

func (m *manager) NewAPIKey() (plain string, prefix string, hash string, err error) {
	token, err := randomToken(32)
	if err != nil {
		return "", "", "", err
	}
	plain = fmt.Sprintf("%s_%s", APIKeyPrefix, token)
	prefix = keyPrefix(plain)
	hash = m.HashAPIKey(plain)
	return plain, prefix, hash, nil
}

func (m *manager) HashAPIKey(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}

func randomToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func keyPrefix(key string) string {
	if len(key) <= 12 {
		return key
	}
	return strings.TrimSpace(key[:12])
}
