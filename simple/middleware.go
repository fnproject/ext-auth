package simple

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/fnproject/fn/api/server"
)

type SimpleMiddleware struct{}

func (m *SimpleMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		fmt.Println("Simple Auth middleware called")

		// skip if this is our login url
		if path.Base(r.URL.Path) == "login" {
			next.ServeHTTP(w, r)
			return
		}

		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {
			server.WriteError(ctx, w, http.StatusUnauthorized, errors.New("No Authorization header, access denied"))
			return
		}

		ahSplit := strings.Split(authorizationHeader, " ")
		if len(ahSplit) != 2 {
			server.WriteError(ctx, w, http.StatusUnauthorized, errors.New("Invalid authorization header, access denied"))
			return
		}
		token, err := jwt.Parse(ahSplit[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(os.Getenv(EnvSecret)), nil
		})
		if err != nil {
			server.WriteError(ctx, w, http.StatusUnauthorized, err)
			return
		}
		if !token.Valid {
			server.WriteError(ctx, w, http.StatusUnauthorized, errors.New("Invalid authorization token, access denied"))
			return
		}
		next.ServeHTTP(w, r)

	})
}
