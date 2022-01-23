package main

import (
	"os"
	"net/http"
	"log"
	"encoding/json"
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	// "github.com/joho/godotenv" //DEV
)

// func initLocalEnv() {
// 	err := godotenv.Load(".env")
// 	if err != nil {
// 		log.Fatalf("Error loading .env file")
// 	}
// }


//respond with 400 and an "error" value in json
func writeErrorToResponse(w http.ResponseWriter, err error) {
	if err != nil {
		w.WriteHeader(400)
		log.Print(err)
		e := Error{
			Error: err.Error(),
		}
		json.NewEncoder(w).Encode(e)
	}
}

//respond with 400 and an "error" value in json
func writeErrorMessageToResponse(w http.ResponseWriter, errorMessage string) {
	w.WriteHeader(400)
	log.Print(errorMessage)
	e := Error{
		Error: errorMessage,
	}
	json.NewEncoder(w).Encode(e)
}

func logError(msg string, err error) {
	if err != nil {
		log.Print(fmt.Errorf(msg+"; %w", err))
	}
}

func hash(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	logError("hashPassword", err)
	return string(bytes)
}

func checkPassword(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	logError("checkPasswordHash", err)
	return err
}

func GetJWT(username string) string {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["username"] = username
	claims["iss"] = "dodue"

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	logError("JWT", err)

	return tokenString
}