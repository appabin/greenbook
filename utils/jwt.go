package utils

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

func HassPassword(pwd string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), 12)
	return string(hash), err
}

func GenerateJWT(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID":	userID,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	})
	SignedToken, err := token.SignedString([]byte("secret"))
	return "Bearer " + SignedToken, err
}

func CheckPassword(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
func ParseJWT(tokenString string) (string, error) {
    tokenString = strings.TrimPrefix(tokenString, "Bearer ")
    if tokenString == "" {
        return "", errors.New("empty token string")
    }

    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("unexpected signing method")
        }
        return []byte("secret"), nil // 建议替换为环境变量配置的密钥
    })

    if err != nil {
        return "", err // 返回具体解析错误
    }

    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        userID, ok := claims["userID"].(string) // 提取userID
        if !ok {
            return "", errors.New("invalid userID claim")
        }
        return userID, nil
    }

    return "", errors.New("invalid token")
}
