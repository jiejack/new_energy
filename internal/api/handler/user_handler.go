package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/api/dto"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// CreateUser 创建用户
// @Summary 创建用户
// @Description 创建新用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user body dto.CreateUserRequest true "用户信息"
// @Success 201 {object} dto.Response{data=dto.UserResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req service.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:      400,
			Message:   "Invalid request parameters",
			Timestamp: 0,
		})
		return
	}

	// TODO: 从上下文中获取操作者ID
	operatorID := ""

	user, err := h.userService.CreateUser(c.Request.Context(), &req, operatorID)
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
		Data:      user,
		Timestamp: 0,
	})
}

// GetUser 获取用户详情
// @Summary 获取用户详情
// @Description 根据ID获取用户详细信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Success 200 {object} dto.Response{data=dto.UserResponse}
// @Failure 404 {object} dto.ErrorResponse
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")

	user, err := h.userService.GetUser(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Code:      404,
			Message:   "User not found",
			Timestamp: 0,
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      user,
		Timestamp: 0,
	})
}

// ListUsers 获取用户列表
// @Summary 获取用户列表
// @Description 获取所有用户的列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Success 200 {object} dto.Response{data=service.UserListResponse}
// @Failure 500 {object} dto.ErrorResponse
// @Router /users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	page := 1
	pageSize := 20

	resp, err := h.userService.ListUsers(c.Request.Context(), nil, page, pageSize)
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

// UpdateUser 更新用户
// @Summary 更新用户
// @Description 更新用户信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Param user body dto.UpdateUserRequest true "用户信息"
// @Success 200 {object} dto.Response{data=dto.UserResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")

	var req service.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:      400,
			Message:   "Invalid request parameters",
			Timestamp: 0,
		})
		return
	}

	// TODO: 从上下文中获取操作者ID
	operatorID := ""

	user, err := h.userService.UpdateUser(c.Request.Context(), id, &req, operatorID)
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
		Data:      user,
		Timestamp: 0,
	})
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Description 删除指定用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Success 204 "No Content"
// @Failure 404 {object} dto.ErrorResponse
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	// TODO: 从上下文中获取操作者ID
	operatorID := ""

	if err := h.userService.DeleteUser(c.Request.Context(), id, operatorID); err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Code:      404,
			Message:   "User not found",
			Timestamp: 0,
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 修改用户密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Param request body dto.ChangePasswordRequest true "密码信息"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /users/{id}/password [put]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	id := c.Param("id")

	var req service.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:      400,
			Message:   "Invalid request parameters",
			Timestamp: 0,
		})
		return
	}

	// TODO: 从上下文中获取操作者ID
	operatorID := ""

	if err := h.userService.ChangePassword(c.Request.Context(), id, &req, operatorID); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:      400,
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
