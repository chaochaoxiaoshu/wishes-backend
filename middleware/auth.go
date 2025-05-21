package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"

	"wishes/models"
	"wishes/utils"
)

var jwtSecret []byte

func InitJWTSecret(secret []byte) {
	jwtSecret = secret
}

type UserType = string

const (
	UserTypeUser  UserType = "user"
	UserTypeAdmin UserType = "admin"
)

type JWTClaims struct {
	UserID  uint
	Type    UserType
	IsAdmin bool
	jwt.RegisteredClaims
}

func GenerateUserToken(user models.User) (string, error) {
	claims := JWTClaims{
		UserID:  user.ID,
		Type:    UserTypeUser,
		IsAdmin: user.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * 7 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "wishes-api",
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtSecret)
}

func GenerateAdminToken(admin models.Admin) (string, error) {
	claims := JWTClaims{
		UserID:  admin.ID,
		Type:    UserTypeAdmin,
		IsAdmin: true,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * 7 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "wishes-api",
			Subject:   fmt.Sprintf("%d", admin.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtSecret)
}

func ParseToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, utils.CreateResponse(nil, "未提供 token"))
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, utils.CreateResponse(nil, "无效的 Authorization 格式"))
			c.Abort()
			return
		}

		claims, err := ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, utils.CreateResponse(nil, "无效的 token"))
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("userType", claims.Type)
		c.Set("isAdmin", claims.IsAdmin)
		c.Next()
	}
}
