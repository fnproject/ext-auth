package simple

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/fnproject/fn/api/server"
)

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

	user, created, err := authenticate(&login)
	if err != nil {
		server.WriteError(ctx, w, http.StatusUnauthorized, err)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iat":      time.Now().Unix(),
		"username": user.Username,
	})
	tokenString, err := token.SignedString([]byte(os.Getenv(EnvSecret)))
	if err != nil {
		server.HandleErrorResponse(ctx, w, err)
		return
	}
	var msg string
	if created {
		msg = "New user created"
	} else {
		msg = "Thanks for coming back!"
	}
	json.NewEncoder(w).Encode(LoginResponse{Token: tokenString, Msg: msg})
}

func authenticate(login *Login) (*User, bool, error) {
	user, err := findUser(login.Username, login.Password)
	if err != nil {
		return nil, false, err
	}
	if user != nil {
		return user, false, nil
	}
	// Since this is dumb, we'll just automatically create a user and return a token.
	user, err = createUser(login.Username, login.Password)
	if err != nil {
		return nil, false, err
	}
	return user, true, nil
}

type Login struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	Msg   string `json:"msg"`
}
