package middleware

import (
	"ai-cloud/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims 结构体定义了JWT中包含的用户相关声明。
// UserID 是用户在JWT中的唯一标识。
// RegisteredClaims 包含了一些标准的JWT声明，如过期时间、发行时间等。
type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateToken 为给定的用户ID生成JWT。
// userID 是要编码到JWT中的用户ID。
// 函数返回生成的JWT字符串和可能的错误。
func GenerateToken(userID uint) (string, error) {
	cfg := config.AppConfigInstance.JWT
	expirationTime := time.Now().Add(time.Duration(cfg.ExpirationHours) * time.Hour)

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

// ParseToken 解析JWT并验证其有效性。
// tokenString 是待解析的JWT字符串。
// 函数返回解析出的Claims指针和可能的错误。
func ParseToken(tokenString string) (*Claims, error) {
	cfg := config.AppConfigInstance.JWT

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrInvalidKey
}
