package handler

import (
	"github.com/k8s/kube-app-operator/internal/approval/models"
	services "github.com/k8s/kube-app-operator/internal/approval/services"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)


type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    TokenData `json:"data"`
}
type TokenData struct {
	Token string `json:"token"`
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type GetUserDTO struct {
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Email string   `json:"email"`
	Roles []string `json:"roles"`
}


type UserListResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}



// 定义列结构

type Column struct {
	Alias string `json:"alias"`
	Name  string `json:"name"`
}

// 定义返回的用户数据（已经“扁平化”roles）

type UserDTO struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Roles     string `json:"roles"` // roles 拼成字符串
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type UsersData struct {
	Total int64        `json:"total"`
	Page  int        `json:"page"`
	Limit int        `json:"limit"`
	Users []UserDTO  `json:"users"`
}

type UserQueryResponse struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Data    UsersData `json:"data"`
	Columns []Column  `json:"columns"`
}


type UserInfoData struct {
	UserID string   `json:"user_id"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
	Status bool     `json:"status"`
	Avatar string   `json:"avatar"`
	Name string `json:"name""`
	Introduction string `json:"introduction"`
}

type CreateUserRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Roles    []string `json:"roles"` // 新增
}


type AssignRolesRequest struct {
	Roles []string `json:"roles" binding:"required"` // 角色名数组
}


// UpdateUserRequest 用于接收更新请求
type UpdateUserRequest struct {
    Name     string `json:"name,omitempty"`
    Email    string `json:"email,omitempty"`
    Password string `json:"password,omitempty"` // 可选更新
    Roles    []string `json:"roles"` // 新增
}


type StandardResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// UserHandler exposes user-related endpoints

type UserHandler struct {
	svc *services.UserService
}

func NewUserHandler(svc *services.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// POST /users

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Code:    40000,
			Message: "请求参数错误: " + err.Error(),
			Data:    nil,
		})
		return
	}

	user, err := h.svc.CreateUserWithRoles(req.Name, req.Email, req.Password, req.Roles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Code:    50000,
			Message: "创建用户失败: " + err.Error(),
			Data:    nil,
		})
		return
	}

	// 返回时不暴露 PasswordHash

	c.JSON(http.StatusCreated, StandardResponse{
		Code:    20000,
		Message: "创建用户成功",
		Data: map[string]interface{}{
			"id":         user.ID,
			"name":       user.Name,
			"email":      user.Email,
			"roles":      user.Roles,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		},
	})
}

// POST /users/:id/roles

func (h *UserHandler) AssignRolesToUser(c *gin.Context) {
	id := c.Param("id")

	var req AssignRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.svc.AssignRoles(c.Request.Context(), id, req.Roles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}


// DELETE /users/:id/roles

func (h *UserHandler) RemoveRoles(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Roles []string `json:"roles" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.RemoveRoles(c.Request.Context(), id, req.Roles); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "roles removed successfully",
		"user_id": id,
		"roles":   req.Roles,
	})
}

// GET /users/:id

func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")
	user, err := h.svc.GetUser(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    50000,
			Message: err.Error(),
		})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, Response{
			Code:    50000,
			Message: "user not found",
		})
		return
	}

	// 提取 roles 的 name
	roleNames := make([]string, 0, len(user.Roles))
	for _, r := range user.Roles {
		roleNames = append(roleNames, r.Name)
	}

	dto := GetUserDTO{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Roles: roleNames,
	}

	c.JSON(http.StatusOK, Response{
		Code:    20000,
		Message: "success",
		Data:    dto,
	})
}

// PUT /users/:id

func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	// 构造要更新的用户，只更新非空字段
	user := &models.User{ID: id}
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}

	updated, err := h.svc.UpdateUser(c.Request.Context(), user, req.Password, req.Roles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	// 成功返回统一格式
	c.JSON(http.StatusOK, gin.H{
		"code":    20000,
		"message": "用户更新成功",
		"data": gin.H{
			"id":         updated.ID,
			"name":       updated.Name,
			"email":      updated.Email,
			"roles":      updated.Roles,
			"created_at": updated.CreatedAt,
			"updated_at": updated.UpdatedAt,
		},
	})
}


// DELETE /users/:id

func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	err := h.svc.DeleteUser(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "The deleted user was not found in the system" {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    40400,
				"message": "user not found",
				"data":    nil,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    50000,
				"message": err.Error(),
				"data":    nil,
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    20000,
		"message": "success",
		"data": gin.H{
			"deleted": id,
		},
	})
}

// GET /users?page=1&limit=10

// handler 修改

func (h *UserHandler) ListUsers(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("page_size", "10")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	users, total, err := h.svc.ListUsers(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 50000, "message": err.Error()})
		return
	}

	// ⚡ 转换为 DTO（roles 拼接成字符串）
	var userDTOs []UserDTO
	for _, u := range users {
		var roleNames []string
		for _, r := range u.Roles {
			roleNames = append(roleNames, r.Name)
		}
		userDTOs = append(userDTOs, UserDTO{
			ID:        u.ID,
			Name:      u.Name,
			Email:     u.Email,
			Roles:     strings.Join(roleNames, ", "), // roles 拼接
			CreatedAt: u.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: u.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	// 定义表头
	columns := []Column{
		{Name: "id", Alias: "标识"},
		{Name: "name", Alias: "用户名"},
		{Name: "email", Alias: "邮箱"},
		{Name: "roles", Alias: "角色"},
		{Name: "created_at", Alias: "创建时间"},
		{Name: "updated_at", Alias: "更新时间"},
	}

	resp := UserQueryResponse{
		Code:    20000,
		Message: "查询用户成功",
		Data: UsersData{
			Total: total,
			Page:  page,
			Limit: limit,
			Users: userDTOs,
		},
		Columns: columns,
	}

	c.JSON(http.StatusOK, resp)
}


// Login 用户登录

func (h *UserHandler) Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "code":    50008,
            "message": "请求参数错误，请检查输入用户 email 和 password",
            "detail":  err.Error(),
        })
        return
    }
    token, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
    if err != nil {
        switch err.Error() {
        case "user not found":
            c.JSON(http.StatusUnauthorized, LoginResponse{
                Code:    40101,
                Message: "用户不存在，请先注册用户",
            })
        case "invalid credentials":
            c.JSON(http.StatusUnauthorized, LoginResponse{
                Code:    40102,
                Message: "邮箱或密码错误，请检查后重试",
            })
        default:
            c.JSON(http.StatusInternalServerError, LoginResponse{
                Code:    50000,
                Message: "登录失败，请检查后再试",
            })
        }
        return
    }
    c.JSON(http.StatusOK, LoginResponse{
        Code:    20000,
        Message: "登录成功",
		Data: TokenData{
			Token: token,
		},
    })
}

func (h *UserHandler) GetUserInfo(c *gin.Context) {
	// 从 JWT 中间件取 user_id 和 email
	userIDVal, _ := c.Get("user_id")
	emailVal, _ := c.Get("email")

	var userID, email string
	if v, ok := userIDVal.(string); ok {
		userID = v
	}
	if v, ok := emailVal.(string); ok {
		email = v
	}

	if userID == "" {
		c.JSON(http.StatusUnauthorized, Response{
			Code:    40100,
			Message: "未提供有效的 user_id（请检查 token）",
		})
		return
	}

	// 数据库查询用户
	user, err := h.svc.GetUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    50000,
			Message: err.Error(),
		})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, Response{
			Code:    40400,
			Message: "用户不存在",
		})
		return
	}

	// 提取角色名
	roleNames := make([]string, 0, len(user.Roles))
	for _, r := range user.Roles {
		roleNames = append(roleNames, r.Name)
	}

	// 构造响应
	resp := Response{
		Code:    20000,
		Message: "获取用户信息成功",
		Data: UserInfoData{
			Name:         user.Name,
			UserID:       userID,
			Email:        email,
			Roles:        roleNames,
			Status:       true,
			Avatar:       "https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif",
			Introduction: "user info from database query",
		},
	}

	c.JSON(http.StatusOK, resp)
}


func UserLogout(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:    20000,
		Message: "Logged out successfully!",
		Data: map[string]interface{}{},
	})
}