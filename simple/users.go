package simple

import (
	"sync"

	"golang.org/x/crypto/bcrypt"
)

var users = &sync.Map{}

func init() {

}

func findUser(username, password string) (*User, error) {
	userI, _ := users.Load(username)
	if userI != nil {
		user := userI.(*User)
		err := CheckPasswordHash(user.PassHash, password)
		if err != nil {
			return nil, err
		}
		return user, nil
	}
	return nil, nil
}

func createUser(username, password string) (*User, error) {
	h, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	u := &User{Username: username, PassHash: h}
	users.Store(username, u)
	return u, nil
}

type User struct {
	Username string `json:"username"`
	PassHash string `json:"passhash"`
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
