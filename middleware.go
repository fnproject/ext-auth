package simple

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/fnproject/fn/fnext"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/fnproject/fn/api/models"
	"github.com/fnproject/fn/api/server"
)

type contextKey string

var (
	userIDKey   = contextKey("user_id")
	usernameKey = contextKey("username")
)

type SimpleMiddleware struct {
	simple *SimpleAuth
}

func (m *SimpleMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

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
		var claims jwt.MapClaims
		var ok bool
		if claims, ok = token.Claims.(jwt.MapClaims); !ok {
			server.WriteError(ctx, w, http.StatusUnauthorized, errors.New("Invalid authorization token, invalid claims, access denied"))
			return
		}
		// ok, so finally we're good
		userID := claims["user_id"].(string)
		ctx = context.WithValue(ctx, userIDKey, userID)
		ctx = context.WithValue(ctx, usernameKey, claims["username"])
		// now if we're in a section that has an app (eg: update a route), let's ensure this user has access
		appNameV := ctx.Value(fnext.AppNameKey)
		if appNameV != nil {
			appName := appNameV.(string)
			// first check if app exists, if it doesn't, we'll just let anyone use it
			app, err := m.simple.ds.GetApp(ctx, appName)
			if err != nil && err != models.ErrAppsNotFound { // too bad GetX() doesn't just return nil, nil. the way it is the extension developer needs to find every not found error...
				server.WriteError(ctx, w, http.StatusInternalServerError, err)
				return
			}
			if app != nil {
				err = m.simple.canAccessApp(ctx, userID, appName)
				if err != nil {
					server.WriteError(ctx, w, http.StatusForbidden, errors.New("app access denied"))
					return
				}
			}
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
