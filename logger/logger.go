package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func CreateLogger(env string) (*zap.Logger, error) {
	if env == "LOCAL" {
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.CallerKey = ""
		return config.Build(zap.AddStacktrace(zapcore.ErrorLevel))
	} else {
		return zap.NewProduction()
	}
}
