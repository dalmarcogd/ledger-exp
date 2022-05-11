package main

import (
	"log"

	"github.com/dalmarcogd/blockchain-exp/internal/api"
	"github.com/dalmarcogd/blockchain-exp/pkg/zapctx"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	err := zapctx.StartZapCtx()
	if err != nil {
		log.Fatal(err)
	}

	app := fx.New(api.Module, fx.NopLogger)
	err = app.Err()
	if err != nil {
		zap.L().Fatal("fx", zap.Error(err))
	}

	app.Run()
}
