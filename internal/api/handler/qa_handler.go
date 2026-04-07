package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/api/dto"
	"github.com/new-energy-monitoring/internal/application/service"
)

type QAHandler struct {
	qaService *service.QAService
}

func NewQAHandler(qaService *service.QAService) *QAHandler {
	return &QAHandler{
		qaService: qaService,
	}
}

func (h *QAHandler) CreateSession(c *gin.Context) {
	var req service.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    400,
			Message: "Invalid request parameters: " + err.Error(),
		})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		userID = "default-user"
	}

	resp, err := h.qaService.CreateSession(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.Response{
		Code:    0,
		Message: "success",
		Data:    resp,
	})
}

func (h *QAHandler) GetSession(c *gin.Context) {
	sessionID := c.Param("id")

	resp, err := h.qaService.GetSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    0,
		Message: "success",
		Data:    resp,
	})
}

func (h *QAHandler) ListSessions(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		userID = "default-user"
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	sessions, total, err := h.qaService.ListUserSessions(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    0,
		Message: "success",
		Data: map[string]interface{}{
			"list":     sessions,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

func (h *QAHandler) DeleteSession(c *gin.Context) {
	sessionID := c.Param("id")

	err := h.qaService.DeleteSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    0,
		Message: "success",
	})
}

func (h *QAHandler) Ask(c *gin.Context) {
	var req service.AskQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    400,
			Message: "Invalid request parameters: " + err.Error(),
		})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		userID = "default-user"
	}

	resp, err := h.qaService.AskQuestion(c.Request.Context(), &req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    0,
		Message: "success",
		Data:    resp,
	})
}

func (h *QAHandler) ArchiveSession(c *gin.Context) {
	sessionID := c.Param("id")

	err := h.qaService.ArchiveSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    0,
		Message: "success",
	})
}

func (h *QAHandler) GetHistory(c *gin.Context) {
	sessionID := c.Param("id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	messages, total, err := h.qaService.GetSessionHistory(c.Request.Context(), sessionID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    0,
		Message: "success",
		Data: map[string]interface{}{
			"list":     messages,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}
