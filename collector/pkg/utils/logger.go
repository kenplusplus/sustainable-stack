package utils

import "go.uber.org/zap"

var Logger *zap.Logger

var Sugar *zap.SugaredLogger

func InitializeLogger() {
	Logger, _ = zap.NewProduction()

	Sugar = Logger.Sugar()
}
