package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/api/dto"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService 认证服务Mock
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(ctx interface{}, req *service.LoginRequest, ipAddress, userAgent string) (*service.LoginResponse, error) {
	args := m.Called(ctx, req, ipAddress, userAgent)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.LoginResponse), args.Error(1)
}

func (m *MockAuthService) RefreshToken(ctx interface{}, refreshToken string) (*service.LoginResponse, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.LoginResponse), args.Error(1)
}

func (m *MockAuthService) ValidateToken(token string) (*auth.Claims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Claims), args.Error(1)
}

func (m *MockAuthService) Logout(ctx interface{}, userID, ipAddress, userAgent string) error {
	args := m.Called(ctx, userID, ipAddress, userAgent)
	return args.Error(0)
}

func setupAuthTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestAuthAPI_Login_Success(t *testing.T) {
	router := setupAuthTestRouter()
	mockAuthService := new(MockAuthService)

	// 准备测试数据
	loginReq := &dto.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}
	expectedResp := &service.LoginResponse{
		AccessToken:  "access-token-123",
		RefreshToken: "refresh-token-123",
		ExpiresIn:    3600,
		TokenType:    "Bearer",
		User: &service.UserInfo{
			ID:          "user-001",
			Username:    "testuser",
			Email:       "test@example.com",
			RealName:    "测试用户",
			Roles:       []string{"admin"},
			Permissions: []string{"user:read"},
		},
	}

	mockAuthService.On("Login", mock.Anything, mock.AnythingOfType("*service.LoginRequest"), mock.Anything, mock.Anything).Return(expectedResp, nil)

	// 设置路由
	router.POST("/api/v1/auth/login", func(c *gin.Context) {
		var req dto.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: "参数错误"})
			return
		}

		loginReq := &service.LoginRequest{
			Username: req.Username,
			Password: req.Password,
		}

		resp, err := mockAuthService.Login(c.Request.Context(), loginReq, c.ClientIP(), c.GetHeader("User-Agent"))
		if err != nil {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Code: 401, Message: "认证失败"})
			return
		}

		c.JSON(http.StatusOK, dto.Response{
			Code:    0,
			Message: "success",
			Data: dto.LoginResponse{
				Token:     resp.AccessToken,
				ExpiresAt: resp.ExpiresIn,
				User: dto.UserResponse{
					ID:       resp.User.ID,
					Username: resp.User.Username,
					Email:    resp.User.Email,
					RealName: resp.User.RealName,
					Roles:    resp.User.Roles,
				},
			},
		})
	})

	// 创建请求
	body, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// 执行请求
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证结果
	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	mockAuthService.AssertExpectations(t)
}

func TestAuthAPI_Login_InvalidCredentials(t *testing.T) {
	router := setupAuthTestRouter()
	mockAuthService := new(MockAuthService)

	loginReq := &dto.LoginRequest{
		Username: "wronguser",
		Password: "wrongpassword",
	}

	mockAuthService.On("Login", mock.Anything, mock.AnythingOfType("*service.LoginRequest"), mock.Anything, mock.Anything).Return(nil, service.ErrInvalidCredentials)

	router.POST("/api/v1/auth/login", func(c *gin.Context) {
		var req dto.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: "参数错误"})
			return
		}

		loginReq := &service.LoginRequest{
			Username: req.Username,
			Password: req.Password,
		}

		_, err := mockAuthService.Login(c.Request.Context(), loginReq, c.ClientIP(), c.GetHeader("User-Agent"))
		if err != nil {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Code: 401, Message: "用户名或密码错误"})
			return
		}

		c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success"})
	})

	body, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.Code)

	mockAuthService.AssertExpectations(t)
}

func TestAuthAPI_Logout_Success(t *testing.T) {
	router := setupAuthTestRouter()
	mockAuthService := new(MockAuthService)

	mockAuthService.On("Logout", mock.Anything, "user-001", mock.Anything, mock.Anything).Return(nil)

	router.POST("/api/v1/auth/logout", func(c *gin.Context) {
		// 模拟从Token中获取用户ID
		userID := "user-001"

		err := mockAuthService.Logout(c.Request.Context(), userID, c.ClientIP(), c.GetHeader("User-Agent"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: "登出失败"})
			return
		}

		c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success"})
	})

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockAuthService.AssertExpectations(t)
}

func TestAuthAPI_ValidateToken_Success(t *testing.T) {
	router := setupAuthTestRouter()
	mockAuthService := new(MockAuthService)

	expectedClaims := &auth.Claims{
		UserID:      "user-001",
		Username:    "testuser",
		Roles:       []string{"admin"},
		Permissions: []string{"user:read"},
	}

	mockAuthService.On("ValidateToken", "valid-token").Return(expectedClaims, nil)

	router.GET("/api/v1/auth/validate", func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Code: 401, Message: "未提供认证令牌"})
			return
		}

		// 移除 "Bearer " 前缀
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		claims, err := mockAuthService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Code: 401, Message: "令牌无效"})
			return
		}

		c.JSON(http.StatusOK, dto.Response{
			Code:    0,
			Message: "success",
			Data: gin.H{
				"user_id":     claims.UserID,
				"username":    claims.Username,
				"roles":       claims.Roles,
				"permissions": claims.Permissions,
			},
		})
	})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/auth/validate", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	mockAuthService.AssertExpectations(t)
}

func TestAuthAPI_RefreshToken_Success(t *testing.T) {
	router := setupAuthTestRouter()
	mockAuthService := new(MockAuthService)

	expectedResp := &service.LoginResponse{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
		ExpiresIn:    3600,
		TokenType:    "Bearer",
		User: &service.UserInfo{
			ID:       "user-001",
			Username: "testuser",
		},
	}

	mockAuthService.On("RefreshToken", mock.Anything, "refresh-token-123").Return(expectedResp, nil)

	router.POST("/api/v1/auth/refresh", func(c *gin.Context) {
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: "参数错误"})
			return
		}

		resp, err := mockAuthService.RefreshToken(c.Request.Context(), req.RefreshToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Code: 401, Message: "刷新令牌无效"})
			return
		}

		c.JSON(http.StatusOK, dto.Response{
			Code:    0,
			Message: "success",
			Data: dto.LoginResponse{
				Token:     resp.AccessToken,
				ExpiresAt: resp.ExpiresIn,
			},
		})
	})

	body, _ := json.Marshal(map[string]string{"refresh_token": "refresh-token-123"})
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	mockAuthService.AssertExpectations(t)
}

func TestAuthAPI_Login_DisabledUser(t *testing.T) {
	router := setupAuthTestRouter()
	mockAuthService := new(MockAuthService)

	loginReq := &dto.LoginRequest{
		Username: "disableduser",
		Password: "password123",
	}

	mockAuthService.On("Login", mock.Anything, mock.AnythingOfType("*service.LoginRequest"), mock.Anything, mock.Anything).Return(nil, service.ErrUserDisabled)

	router.POST("/api/v1/auth/login", func(c *gin.Context) {
		var req dto.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: "参数错误"})
			return
		}

		loginReq := &service.LoginRequest{
			Username: req.Username,
			Password: req.Password,
		}

		_, err := mockAuthService.Login(c.Request.Context(), loginReq, c.ClientIP(), c.GetHeader("User-Agent"))
		if err == service.ErrUserDisabled {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{Code: 403, Message: "用户已被禁用"})
			return
		}
		if err != nil {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Code: 401, Message: "认证失败"})
			return
		}

		c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success"})
	})

	body, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 403, resp.Code)
	assert.Equal(t, "用户已被禁用", resp.Message)

	mockAuthService.AssertExpectations(t)
}

func TestAuthAPI_Middleware_AuthRequired(t *testing.T) {
	router := setupAuthTestRouter()
	mockAuthService := new(MockAuthService)

	// 模拟认证中间件
	authMiddleware := func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Code: 401, Message: "未提供认证令牌"})
			c.Abort()
			return
		}

		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		claims, err := mockAuthService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Code: 401, Message: "令牌无效"})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}

	expectedClaims := &auth.Claims{
		UserID:   "user-001",
		Username: "testuser",
	}
	mockAuthService.On("ValidateToken", "valid-token").Return(expectedClaims, nil)

	router.GET("/api/v1/protected", authMiddleware, func(c *gin.Context) {
		userID, _ := c.Get("userID")
		c.JSON(http.StatusOK, dto.Response{
			Code:    0,
			Message: "success",
			Data: gin.H{
				"user_id": userID,
			},
		})
	})

	// 测试无Token
	req1, _ := http.NewRequest(http.MethodGet, "/api/v1/protected", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusUnauthorized, w1.Code)

	// 测试有效Token
	req2, _ := http.NewRequest(http.MethodGet, "/api/v1/protected", nil)
	req2.Header.Set("Authorization", "Bearer valid-token")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	mockAuthService.AssertExpectations(t)
}

func TestAuthAPI_PermissionCheck(t *testing.T) {
	router := setupAuthTestRouter()

	// 模拟权限检查中间件
	permissionMiddleware := func(requiredPermission string) gin.HandlerFunc {
		return func(c *gin.Context) {
			permissions, exists := c.Get("permissions")
			if !exists {
				c.JSON(http.StatusForbidden, dto.ErrorResponse{Code: 403, Message: "无权限"})
				c.Abort()
				return
			}

			permList := permissions.([]string)
			hasPermission := false
			for _, p := range permList {
				if p == requiredPermission {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				c.JSON(http.StatusForbidden, dto.ErrorResponse{Code: 403, Message: "无权限执行此操作"})
				c.Abort()
				return
			}

			c.Next()
		}
	}

	router.GET("/api/v1/users", func(c *gin.Context) {
		c.Set("permissions", []string{"user:read", "user:create"})
		c.Next()
	}, permissionMiddleware("user:read"), func(c *gin.Context) {
		c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success"})
	})

	router.GET("/api/v1/admin", func(c *gin.Context) {
		c.Set("permissions", []string{"user:read"})
		c.Next()
	}, permissionMiddleware("admin:access"), func(c *gin.Context) {
		c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success"})
	})

	// 测试有权限
	req1, _ := http.NewRequest(http.MethodGet, "/api/v1/users", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// 测试无权限
	req2, _ := http.NewRequest(http.MethodGet, "/api/v1/admin", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusForbidden, w2.Code)
}

func TestAuthAPI_RoleCheck(t *testing.T) {
	router := setupAuthTestRouter()

	// 模拟角色检查中间件
	roleMiddleware := func(requiredRoles ...string) gin.HandlerFunc {
		return func(c *gin.Context) {
			roles, exists := c.Get("roles")
			if !exists {
				c.JSON(http.StatusForbidden, dto.ErrorResponse{Code: 403, Message: "无角色信息"})
				c.Abort()
				return
			}

			roleList := roles.([]string)
			hasRole := false
			for _, r := range roleList {
				for _, required := range requiredRoles {
					if r == required {
						hasRole = true
						break
					}
				}
			}

			if !hasRole {
				c.JSON(http.StatusForbidden, dto.ErrorResponse{Code: 403, Message: "角色权限不足"})
				c.Abort()
				return
			}

			c.Next()
		}
	}

	router.DELETE("/api/v1/users/:id", func(c *gin.Context) {
		c.Set("roles", []string{"admin"})
		c.Next()
	}, roleMiddleware("admin", "super_admin"), func(c *gin.Context) {
		c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success"})
	})

	router.DELETE("/api/v1/system/config", func(c *gin.Context) {
		c.Set("roles", []string{"operator"})
		c.Next()
	}, roleMiddleware("admin", "super_admin"), func(c *gin.Context) {
		c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success"})
	})

	// 测试有角色
	req1, _ := http.NewRequest(http.MethodDelete, "/api/v1/users/user-001", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// 测试无角色
	req2, _ := http.NewRequest(http.MethodDelete, "/api/v1/system/config", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusForbidden, w2.Code)
}
