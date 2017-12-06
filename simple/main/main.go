package main

import (
	"context"

	"github.com/fnproject/fn/api/server"
	"github.com/treeder/fn-ext-auth/simple"
)

func main() {
	ctx := context.Background()
	funcServer := server.NewFromEnv(ctx)
	funcServer.AddExtension(&simple.SimpleAuth{})
	funcServer.Start(ctx)
}
