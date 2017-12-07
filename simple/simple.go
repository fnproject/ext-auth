package simple

import (
	"fmt"
	"os"

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
	return "github.com/treeder/fn-ext-auth/simple"
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
