package middlewares

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func breakBearer(bearer string) (string, error) {
	if bearer == "" {
		return "", errors.New("bearer token not found")
	}

	bearerSlice := strings.Split(bearer, " ")

	if len(bearerSlice) != 2 && bearerSlice[0] != "Bearer" {
		return "", errors.New("bearer token not found")
	}

	return bearerSlice[1], nil
}

// lets try to write monadic middlewares

func Authenticate(next http.Handler) http.Handler {

	secretKey := os.Getenv("SECRET_KEY")

	if secretKey == "" {
		panic("SECRET_KEY is required")
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bearer := r.Header.Get("Authorization")

		token, err := breakBearer(bearer)

		if err != nil {
			http.Error(w, fmt.Sprintf("%v. %v", http.StatusText(http.StatusForbidden), err), http.StatusForbidden)
			return
		}

		check := token != secretKey

		if !check {
			http.Error(w, fmt.Sprintf("%v. %v", http.StatusText(http.StatusForbidden), err), http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
