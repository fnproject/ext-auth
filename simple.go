package simple

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/fnproject/fn/api/id"
	"github.com/fnproject/fn/api/models"
	"github.com/fnproject/fn/api/server"
	"github.com/fnproject/fn/fnext"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	server.RegisterExtension(&SimpleAuth{})
}

const (
	EnvSecret      = "SIMPLE_SECRET"
	EnvMasterToken = "MASTER_TOKEN"
)

type SimpleAuth struct {
	ds                  models.Datastore
	insertUserAppQuery  string
	deleteUserAppQuery  string
	getUserByUsername   string
	getUserAppIDName    string
	getUserAppsByUserID string
}

func (e *SimpleAuth) Name() string {
	return "github.com/fnproject/ext-auth"
}

func (e *SimpleAuth) Setup(s fnext.ExtServer) error {
	fmt.Println("SETTING UP SIMPLE AUTH")
	if os.Getenv(EnvSecret) == "" {
		return fmt.Errorf("%s env var is required for simple auth extension", EnvSecret)
	}

	if os.Getenv(EnvMasterToken) == "" {
		return fmt.Errorf("%s env var is required for simple auth extension", EnvMasterToken)
	}

	// setup database for auth
	simple := &SimpleAuth{
		ds: s.Datastore(),
	}

	err := simple.initDB()
	if err != nil {
		return err
	}

	// for letting user login:
	s.AddEndpoint("POST", "/login", &SimpleEndpoint{simple: simple})
	// for protecting endpoints:
	s.AddAPIMiddleware(&SimpleMiddleware{simple: simple})
	// for checking app access
	s.AddAppListener(&listener{simple: simple})

	return nil
}

var tables = []string{
	`CREATE TABLE IF NOT EXISTS users (
    id varchar(256) NOT NULL,
	username varchar(256) NOT NULL,
	passhash varchar(256) NOT NULL
);`,
	`CREATE TABLE IF NOT EXISTS user_apps (
    user_id varchar(256) NOT NULL,
    app_name varchar(256) NOT NULL
);`,
}

// User represents a user in the auth system.
type User struct {
	ID       string `db:"id"`
	Username string `db:"username"` // we have a username and ID so user can change his username
	PassHash string `json:"passhash"`
}

// UserApps represents a correlation of a user with an app in functions.
type UserApps struct {
	UserID string `db:"user_id"`
	// TODO: should we have app_id's in core? What if user wants to change app name?
	AppName string `db:"app_name"`
}

// InitDB initializes the database tables in the datastore for this auth package.
func (s *SimpleAuth) initDB() error {
	db := s.ds.GetDatabase()
	for _, v := range tables {
		_, err := db.Exec(v)
		if err != nil {
			return err
		}
	}
	s.insertUserAppQuery = db.Rebind("INSERT INTO user_apps (user_id, app_name) VALUES (?, ?)")
	s.deleteUserAppQuery = db.Rebind("DELETE FROM user_apps WHERE user_id=? and app_name=?")
	s.getUserByUsername = db.Rebind("SELECT * FROM users WHERE username=?")
	s.getUserAppIDName = db.Rebind("SELECT * FROM user_apps WHERE user_id=? and app_name=?")
	s.getUserAppsByUserID = db.Rebind(`SELECT * FROM user_apps WHERE user_id=?`)
	return nil
}

func (s *SimpleAuth) findUser(ctx context.Context, username, password string) (*User, error) {
	db := s.ds.GetDatabase()
	query := db.Rebind(`SELECT * FROM users WHERE username=?`)
	row := db.QueryRowxContext(ctx, query, username)
	var user User
	err := row.StructScan(&user)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *SimpleAuth) createUser(ctx context.Context, username, password string) (*User, error) {
	h, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	user := &User{ID: id.New().String(), Username: username, PassHash: h}
	db := s.ds.GetDatabase()
	query := db.Rebind("INSERT INTO users (id, username, passhash) VALUES (:id, :username, :passhash);")
	_, err = db.NamedExecContext(ctx, query, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func (s *SimpleAuth) canAccessApp(ctx context.Context, userID, appName string) error {
	var uapp UserApps
	err := s.ds.GetDatabase().Get(&uapp, s.getUserAppIDName, userID, appName)
	if err != nil {
		if err == sql.ErrNoRows {
			return authErr{code: http.StatusForbidden, reason: "forbidden"}
		}
		return err
	}
	return nil
}
