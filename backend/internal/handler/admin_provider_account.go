package handler

import (
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/javapub/agi-platform-backend/internal/model"
	"github.com/javapub/agi-platform-backend/pkg/errors"
	"github.com/javapub/agi-platform-backend/pkg/response"
)

type providerAccountRequest struct {
	Name string `json:"name" binding:"required"`
	Provider string `json:"provider" binding:"required"`
	APIURL string `json:"api_url" binding:"required"`
	APIKey string `json:"api_key"`
	ExtraConfig map[string]interface{} `json:"extra_config"`
	IsActive bool `json:"is_active"`
}

func (h *AdminConfigHandler) ListProviderAccounts(c *gin.Context) { rows,err:=h.providerRepo.List();if err!=nil{response.Error(c,err);return};for _,row:=range rows{if row.APIKey!=""{row.APIKey="******"}};response.Success(c,rows) }
func (h *AdminConfigHandler) CreateProviderAccount(c *gin.Context) { var req providerAccountRequest;if c.ShouldBindJSON(&req)!=nil||req.APIKey==""{response.Error(c,errors.ErrBadRequest);return};extra,_:=json.Marshal(req.ExtraConfig);row:=&model.AIProviderAccount{Name:req.Name,Provider:req.Provider,APIURL:req.APIURL,APIKey:req.APIKey,ExtraConfig:extra,IsActive:req.IsActive};if err:=h.providerRepo.Create(row);err!=nil{response.Error(c,err);return};row.APIKey="******";response.Success(c,row) }
func (h *AdminConfigHandler) UpdateProviderAccount(c *gin.Context) { id,err:=strconv.ParseInt(c.Param("id"),10,64);if err!=nil{response.Error(c,errors.ErrBadRequest);return};var req providerAccountRequest;if c.ShouldBindJSON(&req)!=nil{response.Error(c,errors.ErrBadRequest);return};row,err:=h.providerRepo.Find(id);if err!=nil{response.Error(c,err);return};extra,_:=json.Marshal(req.ExtraConfig);row.Name=req.Name;row.Provider=req.Provider;row.APIURL=req.APIURL;row.ExtraConfig=extra;row.IsActive=req.IsActive;if req.APIKey!=""&&req.APIKey!="******"{row.APIKey=req.APIKey};if err=h.providerRepo.Update(row);err!=nil{response.Error(c,err);return};row.APIKey="******";response.Success(c,row) }
func (h *AdminConfigHandler) DeleteProviderAccount(c *gin.Context) { id,err:=strconv.ParseInt(c.Param("id"),10,64);if err==nil{err=h.providerRepo.Delete(id)};if err!=nil{response.Error(c,err);return};response.Success(c,nil) }
