package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var jwtKey = []byte("my_test_jwt_token")

type Claims struct {
	UserName string `json:"username"`
	jwt.StandardClaims
}

func GenerateToken(userName string) (string, error) {
	expirationTime := time.Now().Add(5 * time.Minute)

	claims := &Claims{
		UserName: userName,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil

}

func ValidateToken(tokenString string) (string, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return "", fmt.Errorf("invalid signature")
		}
		return "", fmt.Errorf("could not parse token")
	}

	// Ensure the token is valid
	if !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	// Return the username from the token's claims
	return claims.UserName, nil

}

func GenerateTokenHandler(w http.ResponseWriter, r *http.Request) {

	userName := r.URL.Query().Get("user_name")

	tokenString, err := GenerateToken(userName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error generating token"))
		return
	}

	w.Write([]byte(fmt.Sprintf("Generated Token: %s\n", tokenString)))
}

func ValidateTokenHandler(w http.ResponseWriter, r *http.Request) {

	tokenString := r.URL.Query().Get("token")

	username, err := ValidateToken(tokenString)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(fmt.Sprintf("Error: %v\n", err)))
		return
	}

	// Return the username if the token is valid
	w.Write([]byte(fmt.Sprintf("Token is valid. Welcome, %s!\n", username)))

}

func main() {
	// routes for generating and validating tokens
	http.HandleFunc("/generate-token", GenerateTokenHandler)
	http.HandleFunc("/validate-token", ValidateTokenHandler)

	// Start the HTTP server
	http.ListenAndServe(":8080", nil)
}
