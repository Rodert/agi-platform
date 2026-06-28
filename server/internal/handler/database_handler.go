package handler

import (
	"agi-platform/server/internal/response"
	"agi-platform/server/internal/service"

	"github.com/gin-gonic/gin"
)

type DatabaseHandler interface {
	ListTables(c *gin.Context)
	GetTable(c *gin.Context)
}

type databaseHandler struct {
	service service.DatabaseExplorerService
}

func NewDatabaseHandler(service service.DatabaseExplorerService) DatabaseHandler {
	return &databaseHandler{service: service}
}

func (h *databaseHandler) ListTables(c *gin.Context) {
	tables, err := h.service.ListTables(c.Request.Context())
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, tables)
}

func (h *databaseHandler) GetTable(c *gin.Context) {
	result, err := h.service.GetTable(c.Request.Context(), c.Param("table"), queryInt(c, "limit", 20), queryInt(c, "offset", 0))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, result)
}
