//go:build ignore
// +build ignore

// dev-token.go — 生成开发调试用的 HS256 JWT Token
// 由 dev-token bash 脚本调用

package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: go run dev-token.go <user_id> <jwt_secret>")
		os.Exit(1)
	}

	userId, _ := strconv.ParseInt(os.Args[1], 10, 64)
	secret := os.Args[2]

	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": userId,
		"iat":     now.Unix(),
		"exp":     now.Add(24 * time.Hour).Unix(), // 24h 有效期
		"jti":     fmt.Sprintf("%d-%d", userId, now.UnixNano()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		fmt.Fprintln(os.Stderr, "JWT sign error:", err)
		os.Exit(1)
	}

	fmt.Print(signed)
}
