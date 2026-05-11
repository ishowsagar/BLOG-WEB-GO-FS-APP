package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
)

//  func that generated token string parsed with Userid for auth
func GenerateToken(userId uint) (*string,error) {
	// load and access env var "secretKey"

	err := godotenv.Load()
	if err != nil {
		slog.Error("failed to load .env file","error",err)
		os.Exit(1)
	}

	jwtSecretSigningKey := os.Getenv("JWT_SECRET_KEY")

	// create mapped claims with user id and expiry
	mappedClaims := jwt.MapClaims{
		"user_id" : userId,
		"expiry" : time.Now().Add(24 * time.Hour),
	}
	// jwt new map with claims method - create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,mappedClaims)

	// sign token with secret key
	tokenStr,err := token.SignedString([]byte(jwtSecretSigningKey))
	if err != nil {
		return nil,err
	}

	//  return string
	return &tokenStr,err
}

func HashToken(token string) ([]byte,error) {

	sum := sha256.Sum256([]byte(token))
	storedHash := hex.EncodeToString(sum[:])

	return []byte(storedHash),nil
}

// todos -
// * auth - fetches token and attach user id from its claims map - done✅ 
// * req auth - checking if it has token in req  - done✅
// * latency check mw - done✅ 
// * router - cors,protections - done✅
//* Post - let user create post 
// * Comment - let comment on that post with attached id - done✅
// * Likes - likes model,meth,controller,route