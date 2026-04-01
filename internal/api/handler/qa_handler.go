package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/api/dto"
	"github.com/new-energy-monitoring/internal/application/service"
)

// QAHandler 问答处理器
type QAHandler struct {
	qaService *service.QAService
}

// NewQAHandler 创建问答处理器
func NewQAHandler(qaService *service.QAService) *QAHandler {
	return &QAHandler{
		qaService: qaService,
	}
}

// CreateSession 创建会话
// @Summary 创建问答会话
// @Description 创建新的问答会话，用于与AI助手进行对话交互
// @Tags 智能问答
// @Accept json
// @Produce json
// @Param session body service.CreateSessionRequest true "会话信息"
// @Example request {"user_id":"user-001","title":"设备故障诊断"}
// @Success 201 {object} dto.Response{data=service.CreateSessionResponse} "创建成功"
// @Failure 400 {object} dto.ErrorResponse "请求参数错误"
// @Failure 500 {object} dto.ErrorResponse "服务器内部错误"
// @Router /qa/sessions [post]
func (h *QAHandler) CreateSession(c *gin.Context) {
	var req service.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:      400,
			Message:   "Invalid request parameters: " + err.Error(),
			Timestamp: 0,
		})
		return
	}

	resp, err := h.qaService.CreateSession(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:      500,
			Message:   err.Error(),
			Timestamp: 0,
		})
		return
	}

	c.JSON(http.StatusCreated, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      resp,
		Timestamp: 0,
	})
}

// GetSession 获取会话详情
// @Summary 获取会话详情
// @Description 根据ID获取会话详细信息，包括历史消息记录
// @Tags 智能问答
// @Accept json
// @Produce json
// @Param id path string true "会话ID" example(session-001)
// @Success 200 {object} dto.Response{data=service.SessionDetailResponse} "获取成功"
// @Failure 403 {object} dto.ErrorResponse "无权访问该会话"
// @Failure 404 {object} dto.ErrorResponse "会话不存在"
// @Failure 500 {object} dto.ErrorResponse "服务器内部错误"
// @Router /qa/sessions/{id} [get]
func (h *QAHandler) GetSession(c *gin.Context) {
	sessionID := c.Param("id")
	
	// TODO: 从上下文中获取用户ID
	userID := c.GetString("user_id")
	if userID == "" {
		userID = "default-user"
	}

	resp, err := h.qaService.GetSession(c.Request.Context(), sessionID, userID)
	if err != nil {
		if err == service.ErrSessionNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Code:      404,
				Message:   "Session not found",
				Timestamp: 0,
			})
			return
		}
		if err == service.ErrUnauthorizedAccess {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Code:      403,
				Message:   "Unauthorized access to session",
				Timestamp: 0,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:      500,
			Message:   err.Error(),
			Timestamp: 0,
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      resp,
		Timestamp: 0,
	})
}

// ListSessions 获取会话列表
// @Summary 获取用户会话列表
// @Description 获取当前用户的所有会话列表，按更新时间倒序排列
// @Tags 智能问答
// @Accept json
// @Produce json
// @Param page query int false "页码" minimum(1) default(1) example(1)
// @Param page_size query int false "每页数量" minimum(1) maximum(100) default(20) example(20)
// @Success 200 {object} dto.Response{data=service.SessionListResponse} "获取成功"
// @Failure 500 {object} dto.ErrorResponse "服务器内部错误"
// @Router /qa/sessions [get]
func (h *QAHandler) ListSessions(c *gin.Context) {
	// TODO: 从上下文中获取用户ID
	userID := c.GetString("user_id")
	if userID == "" {
		userID = "default-user"
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	resp, err := h.qaService.ListSessions(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:      500,
			Message:   err.Error(),
			Timestamp: 0,
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      resp,
		Timestamp: 0,
	})
}

// DeleteSession 删除会话
// @Summary 删除会话
// @Description 删除指定的问答会话（软删除），删除后不可恢复
// @Tags 智能问答
// @Accept json
// @Produce json
// @Param id path string true "会话ID" example(session-001)
// @Success 200 {object} dto.Response "删除成功"
// @Failure 403 {object} dto.ErrorResponse "无权访问该会话"
// @Failure 404 {object} dto.ErrorResponse "会话不存在"
// @Failure 500 {object} dto.ErrorResponse "服务器内部错误"
// @Router /qa/sessions/{id} [delete]
func (h *QAHandler) DeleteSession(c *gin.Context) {
	sessionID := c.Param("id")
	
	// TODO: 从上下文中获取用户ID
	userID := c.GetString("user_id")
	if userID == "" {
		userID = "default-user"
	}

	err := h.qaService.DeleteSession(c.Request.Context(), sessionID, userID)
	if err != nil {
		if err == service.ErrSessionNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Code:      404,
				Message:   "Session not found",
				Timestamp: 0,
			})
			return
		}
		if err == service.ErrUnauthorizedAccess {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Code:      403,
				Message:   "Unauthorized access to session",
				Timestamp: 0,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:      500,
			Message:   err.Error(),
			Timestamp: 0,
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Timestamp: 0,
	})
}

// Ask 提问
// @Summary 智能问答
// @Description 向AI助手提问并获取回答，支持上下文对话和多轮交互。如果不提供session_id，将自动创建新会话
// @Tags 智能问答
// @Accept json
// @Produce json
// @Param request body service.AskRequest true "提问请求"
// @Example request {"session_id":"session-001","user_id":"user-001","question":"逆变器温度过高怎么处理？"}
// @Success 200 {object} dto.Response{data=service.AskResponse} "获取成功"
// @Failure 400 {object} dto.ErrorResponse "请求参数错误或问题无效"
// @Failure 404 {object} dto.ErrorResponse "会话不存在"
// @Failure 500 {object} dto.ErrorResponse "服务器内部错误"
// @Router /qa/ask [post]
func (h *QAHandler) Ask(c *gin.Context) {
	var req service.AskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:      400,
			Message:   "Invalid request parameters: " + err.Error(),
			Timestamp: 0,
		})
		return
	}

	// TODO: 从上下文中获取用户ID
	if req.UserID == "" {
		req.UserID = c.GetString("user_id")
		if req.UserID == "" {
			req.UserID = "default-user"
		}
	}

	resp, err := h.qaService.Ask(c.Request.Context(), &req)
	if err != nil {
		if err == service.ErrSessionNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Code:      404,
				Message:   "Session not found",
				Timestamp: 0,
			})
			return
		}
		if err == service.ErrInvalidQuestion {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Code:      400,
				Message:   "Invalid question",
				Timestamp: 0,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:      500,
			Message:   err.Error(),
			Timestamp: 0,
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      resp,
		Timestamp: 0,
	})
}

// ArchiveSession 归档会话
// @Summary 归档会话
// @Description 将指定的问答会话归档，归档后会话将不再显示在活跃列表中
// @Tags 智能问答
// @Accept json
// @Produce json
// @Param id path string true "会话ID" example(session-001)
// @Success 200 {object} dto.Response "归档成功"
// @Failure 403 {object} dto.ErrorResponse "无权访问该会话"
// @Failure 404 {object} dto.ErrorResponse "会话不存在"
// @Failure 500 {object} dto.ErrorResponse "服务器内部错误"
// @Router /qa/sessions/{id}/archive [put]
func (h *QAHandler) ArchiveSession(c *gin.Context) {
	sessionID := c.Param("id")
	
	// TODO: 从上下文中获取用户ID
	userID := c.GetString("user_id")
	if userID == "" {
		userID = "default-user"
	}

	err := h.qaService.ArchiveSession(c.Request.Context(), sessionID, userID)
	if err != nil {
		if err == service.ErrSessionNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Code:      404,
				Message:   "Session not found",
				Timestamp: 0,
			})
			return
		}
		if err == service.ErrUnauthorizedAccess {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Code:      403,
				Message:   "Unauthorized access to session",
				Timestamp: 0,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:      500,
			Message:   err.Error(),
			Timestamp: 0,
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Timestamp: 0,
	})
}
