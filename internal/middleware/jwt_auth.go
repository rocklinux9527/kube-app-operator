package middleware

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
	"time"
)
// JWT 密钥（建议用环境变量管理）
var jwtSecret = []byte("JbCYgwaoMH9Xhzt4")



// jwt 定义结构体

type Claims struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
    jwt.RegisteredClaims
}


// ErrorResponse 统一错误响应结构
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}


// JWTAuthMiddleware 是 Gin 中间件，用于验证 JWT

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取 Authorization 头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Code:    40100,
				Message: "缺少或非法的 Authorization 头",
			})
			return
		}

		// 2. 提取 token 字符串
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == "" {
			tokenStr = c.Query("token")
		}

		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Code:    40100,
				Message: "缺少 Token，请在 Header 或 Query 中传入",
			})
			return
		}

		// 3. 解析 token
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			// 确保是 HMAC 签名
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret, nil
		})

		// 4. 校验 token
		if err != nil || !token.Valid {
			msg := "Token 非法"
			code := 40103

			if errors.Is(err, jwt.ErrTokenExpired) {
				msg = "Token 已过期"
				code = 40101
			} else if errors.Is(err, jwt.ErrSignatureInvalid) {
				msg = "Token 签名无效"
				code = 40102
			}

			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Code:    code,
				Message: msg,
				Detail:  err.Error(),
			})
			return
		}

		// 5. 再次手动校验过期时间（双保险）
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Code:    40101,
				Message: "Token 已过期",
			})
			return
		}

		// 6. 校验 user_id
		if claims.UserID == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, ErrorResponse{
				Code:    40301,
				Message: "Token 中缺少 user_id",
			})
			return
		}

		// 7. 将 user_id 写入上下文
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		// 继续处理
		c.Next()
	}
}

// GenerateToken 生成 JWT
func GenerateToken(userID, email string) (string, error) {
    claims := Claims{
        UserID: userID,
        Email:  email,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(8 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
            Issuer:    "approval-system",
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}



// ParseToken 验证 JWT

func ParseToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return jwtSecret, nil
    })
    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }
    return nil, jwt.ErrTokenInvalidClaims
}
