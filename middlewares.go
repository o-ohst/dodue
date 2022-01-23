package main

import (
	"net/http"
	"log"
	"os"
	"fmt"
	"encoding/json"
	jwt "github.com/dgrijalva/jwt-go"
)

func authApi(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("api_key")

		if apiKey == "" {
			log.Print("no api key provided")
			w.WriteHeader(401)
			return
		}

		if apiKey != os.Getenv("API_SECRET") {
			w.WriteHeader(401)
			return
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func authJwt(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		tokenCookie, err := r.Cookie("token")

		if err == nil {

			token, err := jwt.Parse(tokenCookie.Value, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf(("invalid signing method"))
				}

				iss := "dodue"
				checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(iss, true)
				if !checkIss {
					return nil, fmt.Errorf(("invalid iss"))
				}

				return []byte(os.Getenv("JWT_SECRET")), nil

			})

			if err != nil {
				logError("authJwt Parse", err)
			}

			if token.Valid {
				next.ServeHTTP(w, r)
			}

		} else {
			log.Print("no authorization token provided")
			e := Error{
				Error: "no authorization token provided",
			}
			w.WriteHeader(401)
			json.NewEncoder(w).Encode(e)
			return
		}
	})
}

//check if user_id cookie is included
func checkUserCookie(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := r.Cookie("user_id")
		if err != nil {
			writeErrorMessageToResponse(w, "no user_id cookie provided")
			return
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func cors(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "https://dodue.netlify.app")
    	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
    	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type,  Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, DNT, Referer, token, user_id, api_key, task_id, category_id, username, password, set-cookie, done")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(200)
			return
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
