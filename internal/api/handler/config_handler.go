package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/api/dto"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/domain/entity"
)

// ConfigHandler 系统配置处理器
type ConfigHandler struct {
	configService *service.ConfigService
}

// NewConfigHandler 创建系统配置处理器
func NewConfigHandler(configService *service.ConfigService) *ConfigHandler {
	return &ConfigHandler{
		configService: configService,
	}
}

// GetAllConfigs 获取所有系统配置
// @Summary 获取所有系统配置
// @Description 获取系统中所有的配置项，返回配置列表
// @Tags 系统配置
// @Accept json
// @Produce json
// @Success 200 {object} dto.Response{data=[]entity.SystemConfig} "获取成功"
// @Failure 500 {object} dto.ErrorResponse "服务器内部错误"
// @Router /configs [get]
func (h *ConfigHandler) GetAllConfigs(c *gin.Context) {
	configs, err := h.configService.GetAllConfigs(c.Request.Context())
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
		Data:      configs,
		Timestamp: 0,
	})
}

// GetConfigsByCategory 获取指定分类的配置
// @Summary 获取指定分类的配置
// @Description 根据分类获取系统配置列表，如basic、alarm、notification等分类
// @Tags 系统配置
// @Accept json
// @Produce json
// @Param category path string true "配置分类" example(basic)
// @Success 200 {object} dto.Response{data=service.ConfigCategoryResponse} "获取成功"
// @Failure 500 {object} dto.ErrorResponse "服务器内部错误"
// @Router /configs/{category} [get]
func (h *ConfigHandler) GetConfigsByCategory(c *gin.Context) {
	category := c.Param("category")

	resp, err := h.configService.GetConfigsByCategory(c.Request.Context(), category)
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

// GetConfig 获取单个配置项
// @Summary 获取单个配置项
// @Description 根据分类和键获取配置项详细信息
// @Tags 系统配置
// @Accept json
// @Produce json
// @Param category path string true "配置分类" example(basic)
// @Param key path string true "配置键" example(system_name)
// @Success 200 {object} dto.Response{data=entity.SystemConfig} "获取成功"
// @Failure 404 {object} dto.ErrorResponse "配置项不存在"
// @Failure 500 {object} dto.ErrorResponse "服务器内部错误"
// @Router /configs/{category}/{key} [get]
func (h *ConfigHandler) GetConfig(c *gin.Context) {
	category := c.Param("category")
	key := c.Param("key")

	config, err := h.configService.GetConfig(c.Request.Context(), category, key)
	if err != nil {
		if err == service.ErrConfigNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Code:      404,
				Message:   "Config not found",
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
		Data:      config,
		Timestamp: 0,
	})
}

// UpdateConfig 更新配置项
// @Summary 更新配置项
// @Description 更新指定分类和键的配置项，支持更新值、值类型和描述
// @Tags 系统配置
// @Accept json
// @Produce json
// @Param category path string true "配置分类" example(basic)
// @Param key path string true "配置键" example(system_name)
// @Param request body service.UpdateConfigRequest true "配置信息"
// @Example request {"value":"新能源监控系统","value_type":"string","description":"系统名称"}
// @Success 200 {object} dto.Response{data=entity.SystemConfig} "更新成功"
// @Failure 400 {object} dto.ErrorResponse "请求参数错误或值类型转换失败"
// @Failure 404 {object} dto.ErrorResponse "配置项不存在"
// @Failure 500 {object} dto.ErrorResponse "服务器内部错误"
// @Router /configs/{category}/{key} [put]
func (h *ConfigHandler) UpdateConfig(c *gin.Context) {
	category := c.Param("category")
	key := c.Param("key")

	var req service.UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:      400,
			Message:   "Invalid request parameters: " + err.Error(),
			Timestamp: 0,
		})
		return
	}

	// TODO: 从上下文中获取操作者ID
	operatorID := ""

	config, err := h.configService.UpdateConfig(c.Request.Context(), category, key, &req, operatorID)
	if err != nil {
		if err == service.ErrConfigNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Code:      404,
				Message:   "Config not found",
				Timestamp: 0,
			})
			return
		}
		if err == service.ErrInvalidValueType || err == service.ErrValueConversion {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Code:      400,
				Message:   err.Error(),
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
		Data:      config,
		Timestamp: 0,
	})
}

// BatchUpdateConfigs 批量更新配置
// @Summary 批量更新配置
// @Description 批量更新多个配置项，支持一次性更新多个配置
// @Tags 系统配置
// @Accept json
// @Produce json
// @Param request body service.BatchUpdateConfigRequest true "批量更新请求"
// @Example request {"configs":[{"category":"basic","key":"system_name","value":"新能源监控系统","value_type":"string"},{"category":"basic","key":"system_version","value":"1.0.0","value_type":"string"}]}
// @Success 200 {object} dto.Response "更新成功"
// @Failure 400 {object} dto.ErrorResponse "请求参数错误或值类型转换失败"
// @Failure 404 {object} dto.ErrorResponse "配置项不存在"
// @Failure 500 {object} dto.ErrorResponse "服务器内部错误"
// @Router /configs/batch [post]
func (h *ConfigHandler) BatchUpdateConfigs(c *gin.Context) {
	var req service.BatchUpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:      400,
			Message:   "Invalid request parameters: " + err.Error(),
			Timestamp: 0,
		})
		return
	}

	// TODO: 从上下文中获取操作者ID
	operatorID := ""

	if err := h.configService.BatchUpdateConfigs(c.Request.Context(), &req, operatorID); err != nil {
		if err == service.ErrConfigNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Code:      404,
				Message:   err.Error(),
				Timestamp: 0,
			})
			return
		}
		if err == service.ErrInvalidValueType || err == service.ErrValueConversion {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Code:      400,
				Message:   err.Error(),
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

// CreateConfig 创建配置项
// @Summary 创建配置项
// @Description 创建新的系统配置项，配置键在分类内必须唯一
// @Tags 系统配置
// @Accept json
// @Produce json
// @Param request body service.CreateConfigRequest true "配置信息"
// @Example request {"category":"basic","key":"system_name","value":"新能源监控系统","value_type":"string","description":"系统名称"}
// @Success 201 {object} dto.Response{data=entity.SystemConfig} "创建成功"
// @Failure 400 {object} dto.ErrorResponse "请求参数错误、配置键已存在或值类型转换失败"
// @Failure 500 {object} dto.ErrorResponse "服务器内部错误"
// @Router /configs [post]
func (h *ConfigHandler) CreateConfig(c *gin.Context) {
	var req service.CreateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:      400,
			Message:   "Invalid request parameters: " + err.Error(),
			Timestamp: 0,
		})
		return
	}

	// TODO: 从上下文中获取操作者ID
	operatorID := ""

	config, err := h.configService.CreateConfig(c.Request.Context(), &req, operatorID)
	if err != nil {
		if err == service.ErrConfigKeyExists {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Code:      400,
				Message:   "Config key already exists",
				Timestamp: 0,
			})
			return
		}
		if err == service.ErrInvalidValueType || err == service.ErrValueConversion {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Code:      400,
				Message:   err.Error(),
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

	c.JSON(http.StatusCreated, dto.Response{
		Code:      0,
		Message:   "success",
		Data:      config,
		Timestamp: 0,
	})
}

// DeleteConfig 删除配置项
// @Summary 删除配置项
// @Description 删除指定的系统配置项，删除后不可恢复
// @Tags 系统配置
// @Accept json
// @Produce json
// @Param category path string true "配置分类" example(basic)
// @Param key path string true "配置键" example(system_name)
// @Success 204 "删除成功"
// @Failure 404 {object} dto.ErrorResponse "配置项不存在"
// @Failure 500 {object} dto.ErrorResponse "服务器内部错误"
// @Router /configs/{category}/{key} [delete]
func (h *ConfigHandler) DeleteConfig(c *gin.Context) {
	category := c.Param("category")
	key := c.Param("key")

	// TODO: 从上下文中获取操作者ID
	operatorID := ""

	if err := h.configService.DeleteConfig(c.Request.Context(), category, key, operatorID); err != nil {
		if err == service.ErrConfigNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Code:      404,
				Message:   "Config not found",
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

	c.JSON(http.StatusNoContent, nil)
}

// ListConfigs 分页查询配置列表
// @Summary 分页查询配置列表
// @Description 分页查询系统配置列表，支持按分类和键过滤
// @Tags 系统配置
// @Accept json
// @Produce json
// @Param category query string false "配置分类" example(basic)
// @Param key query string false "配置键" example(system)
// @Param page query int false "页码" minimum(1) default(1) example(1)
// @Param page_size query int false "每页数量" minimum(1) maximum(100) default(20) example(20)
// @Success 200 {object} dto.PagedResponse{data=[]entity.SystemConfig} "获取成功"
// @Failure 500 {object} dto.ErrorResponse "服务器内部错误"
// @Router /configs/list [get]
func (h *ConfigHandler) ListConfigs(c *gin.Context) {
	// 构建过滤器
	filter := &entity.SystemConfigFilter{
		Page:     1,
		PageSize: 20,
	}

	// 解析查询参数
	if category := c.Query("category"); category != "" {
		filter.Category = &category
	}
	if key := c.Query("key"); key != "" {
		filter.Key = &key
	}
	if page, err := strconv.Atoi(c.Query("page")); err == nil && page > 0 {
		filter.Page = page
	}
	if pageSize, err := strconv.Atoi(c.Query("page_size")); err == nil && pageSize > 0 {
		filter.PageSize = pageSize
	}

	resp, err := h.configService.ListConfigs(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:      500,
			Message:   err.Error(),
			Timestamp: 0,
		})
		return
	}

	c.JSON(http.StatusOK, dto.PagedResponse{
		Code:      0,
		Message:   "success",
		Data:      resp.Configs,
		Total:     resp.Total,
		Page:      resp.Page,
		PageSize:  resp.PageSize,
		Timestamp: 0,
	})
}
