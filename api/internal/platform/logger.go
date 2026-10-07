package platform

import (
	"context"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewLogger(lc fx.Lifecycle, c Config) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	if err := cfg.Level.UnmarshalText([]byte(c.LogLevel)); err != nil {
		return nil, err
	}
	log, err := cfg.Build()
	if err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{OnStop: func(context.Context) error { _ = log.Sync(); return nil }})
	return log, nil
}
