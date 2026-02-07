// Package logger предоставляет обертку над zap logger для логирования приложения.
package logger

import "go.uber.org/zap"

type Logger struct {
	Log zap.SugaredLogger
}

func GetLogger() Logger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	return Logger{Log: *logger.Sugar()}
}
