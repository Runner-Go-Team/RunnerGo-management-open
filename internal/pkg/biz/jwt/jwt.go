package jwt

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"

	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/conf"
)

func GenerateToken(userID string) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(24 * time.Hour * 365)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"iss":     conf.Conf.JWT.Issuer,
		"iat":     now.Unix(),
		"nbf":     now.Unix(),
		"exp":     exp.Unix(),
	})
	tokenString, err := token.SignedString([]byte(conf.Conf.JWT.Secret))

	return tokenString, exp, err
}

func GenerateTokenByTime(userID string, d time.Duration) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(d)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"iss":     conf.Conf.JWT.Issuer,
		"iat":     now.Unix(),
		"nbf":     now.Unix(),
		"exp":     exp.Unix(),
	})
	tokenString, err := token.SignedString([]byte(conf.Conf.JWT.Secret))

	return tokenString, exp, err
}

func ParseToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(conf.Conf.JWT.Secret), nil
	})

	if err != nil {
		return "", jwt.ErrHashUnavailable
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if userID, ok := claims["user_id"]; ok {

			if i, ok := userID.(string); ok {
				return i, nil
			}

		}
	}

	return "", jwt.ErrTokenInvalidClaims
}

func RefreshToken(tokenString string) (string, time.Time, error) {
	userID, err := ParseToken(tokenString)
	if err != nil {
		return "", time.Now(), err
	}

	return GenerateToken(userID)
}

func GetUserIDByCtx(c *gin.Context) string {
	return c.GetString("user_id")
}
