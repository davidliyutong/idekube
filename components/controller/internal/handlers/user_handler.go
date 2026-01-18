package handlers

import (
	"net/http"
	"strconv"

	"github.com/davidliyutong/idekube-controller/internal/middleware"
	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/davidliyutong/idekube-controller/internal/services"
	"github.com/gin-gonic/gin"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService        *services.UserService
	enableRegistration bool
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *services.UserService, enableRegistration bool) *UserHandler {
	return &UserHandler{
		userService:        userService,
		enableRegistration: enableRegistration,
	}
}

// RefreshToken godoc
// @Summary 刷新访问令牌
// @Description 使用刷新令牌获取新的访问令牌和刷新令牌
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.RefreshTokenRequest true "刷新令牌请求"
// @Success 200 {object} models.APIResponse{data=models.LoginResponse} "刷新成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "令牌无效"
// @Router /auth/refresh [post]
func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
				Details: err.Error(),
			},
		})
		return
	}

	response, err := h.userService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "REFRESH_FAILED",
				Message: "Failed to refresh token",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    response,
	})
}

// Logout godoc
// @Summary 用户登出
// @Description 撤销用户的所有刷新令牌
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse "登出成功"
// @Failure 401 {object} models.APIResponse "未认证"
// @Router /auth/logout [post]
func (h *UserHandler) Logout(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	// Revoke all refresh tokens for the user
	if err := h.userService.RevokeAllRefreshTokens(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "LOGOUT_FAILED",
				Message: "Failed to logout",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Logged out successfully",
	})
}

// GetUserProfile godoc
// @Summary 获取当前用户信息
// @Description 获取当前登录用户的详细信息
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=models.User} "成功"
// @Failure 401 {object} models.APIResponse "未认证"
// @Router /users/me/profile [get]
func (h *UserHandler) GetUserProfile(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NOT_FOUND",
				Message: "User not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    user,
	})
}

// GetUser godoc
// @Summary 获取用户信息
// @Description 根据ID获取用户的详细信息
// @Tags User Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} models.APIResponse{data=models.User} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 404 {object} models.APIResponse "用户不存在"
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid user ID",
			},
		})
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NOT_FOUND",
				Message: "User not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    user,
	})
}

// List godoc
// @Summary 列出用户
// @Description 获取所有用户列表（仅管理员）
// @Tags User Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} models.APIResponse{data=models.PaginatedResponse} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Failure 500 {object} models.APIResponse "内部服务器错误"
// @Router /users [get]
func (h *UserHandler) List(c *gin.Context) {
	var opts models.ListOptions
	if err := c.ShouldBindQuery(&opts); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid query parameters",
				Details: err.Error(),
			},
		})
		return
	}

	// Admin uses ListAllUsers which has pagination
	users, total, err := h.userService.ListAllUsers(c.Request.Context(), &opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to list users",
				Details: err.Error(),
			},
		})
		return
	}

	totalPages := int(total) / opts.PageSize
	if int(total)%opts.PageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: models.PaginatedResponse{
			Items:      users,
			Total:      total,
			Page:       opts.Page,
			PageSize:   opts.PageSize,
			TotalPages: totalPages,
		},
	})
}

// Create godoc
// @Summary 创建用户
// @Description 创建新的用户账号（仅管理员）
// @Tags User Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateUserRequest true "用户创建请求"
// @Success 201 {object} models.APIResponse{data=models.User} "创建成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Router /users [post]
func (h *UserHandler) Create(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
				Details: err.Error(),
			},
		})
		return
	}

	user, err := h.userService.CreateUser(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "CREATE_FAILED",
				Message: "Failed to create user",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    user,
	})
}

// UpdateUserProfile godoc
// @Summary 更新个人资料
// @Description 更新当前用户的个人资料
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.UpdateUserProfileRequest true "资料更新请求"
// @Success 200 {object} models.APIResponse{data=models.User} "更新成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Router /users/me/profile [put]
func (h *UserHandler) UpdateUserProfile(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	var req models.UpdateUserProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
				Details: err.Error(),
			},
		})
		return
	}

	user, err := h.userService.UpdateUserProfile(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UPDATE_FAILED",
				Message: "Failed to update profile",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    user,
	})
}

// Update godoc
// @Summary 更新用户
// @Description 更新指定用户的信息（仅管理员）
// @Tags User Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param request body models.UpdateUserRequest true "用户更新请求"
// @Success 200 {object} models.APIResponse{data=models.User} "更新成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Router /users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid user ID",
			},
		})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
				Details: err.Error(),
			},
		})
		return
	}

	// Check if caller is super admin (for role changes)
	// Note: This is a business rule check, not a permission check
	// Only super_admin can modify user roles through the service layer
	// Permission to update user is already checked by RBACCheckEndpoint middleware
	userRole := models.UserRole(c.GetString("user_role"))
	isAdmin := userRole == models.UserRoleSuperAdmin || userRole == models.UserRoleAdmin

	user, err := h.userService.UpdateUserByAdmin(c.Request.Context(), id, &req, isAdmin)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UPDATE_FAILED",
				Message: "Failed to update user",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    user,
	})
}

// Delete godoc
// @Summary 删除用户
// @Description 删除指定的用户
// @Tags User Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} models.APIResponse "删除成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Router /users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid user ID",
			},
		})
		return
	}

	err = h.userService.DeleteUser(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DELETE_FAILED",
				Message: "Failed to delete user",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "User deleted successfully",
	})
}

// ChangePassword godoc
// @Summary 修改密码
// @Description 修改当前用户的密码
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.ChangePasswordRequest true "密码修改请求"
// @Success 200 {object} models.APIResponse "修改成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Router /users/me/security/password [put]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
				Details: err.Error(),
			},
		})
		return
	}

	err := h.userService.ChangePassword(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "CHANGE_PASSWORD_FAILED",
				Message: "Failed to change password",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Password changed successfully",
	})
}

// CheckUserExistence godoc
// @Summary 检查用户是否存在
// @Description 通过用户名检查用户是否存在,仅限 power_user 及以上角色
// @Tags User Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param username query string true "用户名"
// @Success 200 {object} models.APIResponse{data=models.CheckUserExistsResponse} "检查成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未授权"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Security Bearer
// @Router /users/check [get]
func (h *UserHandler) CheckUserExistence(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Username parameter is required",
			},
		})
		return
	}

	exists, err := h.userService.CheckUserExists(c.Request.Context(), username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "CHECK_USER_FAILED",
				Message: "Failed to check user existence",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: models.CheckUserExistsResponse{
			Exists:   exists,
			Username: username,
		},
	})
}

// Search godoc
// @Summary 搜索用户
// @Description 根据查询字符串搜索用户,仅限 admin 及以上角色
// @Tags User Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param query query string true "搜索查询字符串(用户名或邮箱)"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} models.APIResponse{data=models.PaginatedResponse} "搜索成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未授权"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Security Bearer
// @Router /users/search [get]
func (h *UserHandler) Search(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Query parameter is required",
			},
		})
		return
	}

	// Parse pagination and other options
	var opts models.ListOptions
	if err := c.ShouldBindQuery(&opts); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid query parameters",
				Details: err.Error(),
			},
		})
		return
	}

	users, total, err := h.userService.SearchUsers(c.Request.Context(), query, &opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "SEARCH_USERS_FAILED",
				Message: "Failed to search users",
				Details: err.Error(),
			},
		})
		return
	}

	totalPages := int(total) / opts.PageSize
	if int(total)%opts.PageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: models.PaginatedResponse{
			Items:      users,
			Total:      total,
			Page:       opts.Page,
			PageSize:   opts.PageSize,
			TotalPages: totalPages,
		},
	})
}

// ============================================================================
// Sub-resource Handlers
// ============================================================================

// GetProfile godoc
// @Summary 获取用户Profile
// @Description 获取指定用户的Profile子资源
// @Tags User Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} models.APIResponse{data=models.UserProfileResponse} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 404 {object} models.APIResponse "用户不存在"
// @Router /users/{id}/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid user ID",
			},
		})
		return
	}

	profile, err := h.userService.GetUserProfile(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NOT_FOUND",
				Message: "User not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    profile,
	})
}

// UpdateProfile godoc
// @Summary 更新用户Profile
// @Description 更新指定用户的Profile子资源
// @Tags User Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param request body models.UpdateUserProfileRequest true "Profile更新请求"
// @Success 200 {object} models.APIResponse{data=models.UserProfileResponse} "更新成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Failure 404 {object} models.APIResponse "用户不存在"
// @Router /users/{id}/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid user ID",
			},
		})
		return
	}

	var req models.UpdateUserProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
				Details: err.Error(),
			},
		})
		return
	}

	profile, err := h.userService.UpdateUserProfileSubResource(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UPDATE_FAILED",
				Message: "Failed to update user profile",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    profile,
	})
}

// GetUserSecurity godoc
// @Summary 获取用户Security
// @Description 获取指定用户的Security子资源（不含敏感信息）
// @Tags User Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} models.APIResponse{data=models.UserSecurityResponse} "成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 404 {object} models.APIResponse "用户不存在"
// @Router /users/{id}/security [get]
func (h *UserHandler) GetUserSecurity(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid user ID",
			},
		})
		return
	}

	security, err := h.userService.GetUserSecurity(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NOT_FOUND",
				Message: "User not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    security,
	})
}

// UpdateUserSecurity godoc
// @Summary 更新用户Security
// @Description 更新指定用户的Security子资源（密码、MFA等）
// @Tags User Sub-resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param request body models.UpdateUserSecurityRequest true "Security更新请求"
// @Success 200 {object} models.APIResponse{data=models.UserSecurityResponse} "更新成功"
// @Failure 400 {object} models.APIResponse "请求参数错误"
// @Failure 401 {object} models.APIResponse "未认证"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Failure 404 {object} models.APIResponse "用户不存在"
// @Router /users/{id}/security [put]
func (h *UserHandler) UpdateUserSecurity(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid user ID",
			},
		})
		return
	}

	var req models.UpdateUserSecurityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
				Details: err.Error(),
			},
		})
		return
	}

	security, err := h.userService.UpdateUserSecurity(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UPDATE_FAILED",
				Message: "Failed to update user security",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    security,
	})
}

// GetUserEmail godoc
// @Summary Get current user email
// @Description Get email address for current user
// @Tags Users
// @Produce json
// @Success 200 {object} models.APIResponse{data=models.UserEmailResponse}
// @Failure 404 {object} models.APIResponse
// @Security BearerAuth
// @Router /users/me/email [get]
func (h *UserHandler) GetUserEmail(c *gin.Context) {
	id := middleware.MustGetUserID(c)

	email, err := h.userService.GetUserEmail(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NOT_FOUND",
				Message: "Failed to get user email",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    email,
	})
}

// GetEmail godoc
// @Summary Get user email by ID
// @Description Get email address for a specific user (admin access)
// @Tags User Sub-resources
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} models.APIResponse{data=models.UserEmailResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Security BearerAuth
// @Router /users/{id}/email [get]
func (h *UserHandler) GetEmail(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid user ID",
			},
		})
		return
	}

	email, err := h.userService.GetUserEmail(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NOT_FOUND",
				Message: "Failed to get user email",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    email,
	})
}

// UpdateUserEmail godoc
// @Summary Update current user email
// @Description Update email address for current user
// @Tags Users
// @Accept json
// @Produce json
// @Param request body models.UpdateUserEmailRequest true "Email update request"
// @Success 200 {object} models.APIResponse{data=models.UserEmailResponse}
// @Failure 400 {object} models.APIResponse
// @Security BearerAuth
// @Router /users/me/email [put]
func (h *UserHandler) UpdateUserEmail(c *gin.Context) {
	id := middleware.MustGetUserID(c)

	var req models.UpdateUserEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
				Details: err.Error(),
			},
		})
		return
	}

	email, err := h.userService.UpdateUserEmail(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UPDATE_FAILED",
				Message: "Failed to update user email",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    email,
	})
}

// UpdateEmail godoc
// @Summary Update user email by ID
// @Description Update email address for a specific user (admin access)
// @Tags User Sub-resources
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param request body models.UpdateUserEmailRequest true "Email update request"
// @Success 200 {object} models.APIResponse{data=models.UserEmailResponse}
// @Failure 400 {object} models.APIResponse
// @Security BearerAuth
// @Router /users/{id}/email [put]
func (h *UserHandler) UpdateEmail(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid user ID",
			},
		})
		return
	}

	var req models.UpdateUserEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
				Details: err.Error(),
			},
		})
		return
	}

	email, err := h.userService.UpdateUserEmail(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UPDATE_FAILED",
				Message: "Failed to update user email",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    email,
	})
}
