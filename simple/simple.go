package simple

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/fnproject/fn/api/server"
	"github.com/fnproject/fn/fnext"
)

func init() {
	server.RegisterExtension(&SimpleAuth{})
}

const (
	EnvSecret = "SIMPLE_SECRET"
)

type SimpleAuth struct {
}

func (e *SimpleAuth) Name() string {
	return "github.com/treeder/fn-ext-example/logspam"
}

func (e *SimpleAuth) Setup(s fnext.ExtServer) error {
	fmt.Println("SETTING UP SIMPLE")
	if os.Getenv(EnvSecret) == "" {
		return fmt.Errorf("%s env var is required for simple auth extension", EnvSecret)
	}
	// for letting user login:
	s.AddEndpoint("POST", "/login", &SimpleEndpoint{})
	// for protecting endpoints:
	s.AddAPIMiddleware(&SimpleMiddleware{})
	return nil
}

// SimpleEndpoint is used for logging in. Returns a JWT token if successful.
type SimpleEndpoint struct {
}

func (e *SimpleEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("SIMPLEENDPOINT SERVEHTTP")
	ctx := r.Context()
	// parse JSON input containing username and password
	var login Login
	err := json.NewDecoder(r.Body).Decode(&login)
	if err != nil {
		server.HandleErrorResponse(ctx, w, err)
		return
	}
	/////////////////////////////
	// TODO: This is where you verify the credentials.
	/////////////////////////////

	// Since this is dumb, we'll just automatically authenticate and return a token.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iat":      time.Now().Unix(),
		"username": login.Username,
	})
	tokenString, err := token.SignedString([]byte(os.Getenv(EnvSecret)))
	if err != nil {
		server.HandleErrorResponse(ctx, w, err)
		return
	}
	json.NewEncoder(w).Encode(JwtToken{Token: tokenString})
}

type Login struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type JwtToken struct {
	Token string `json:"token"`
}

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
