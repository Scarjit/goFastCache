package logger

import (
	"fmt"
	"go.uber.org/zap"
	"os"
)

func InitLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Printf("Can't create logger: %v\n", err)
		os.Exit(1)
	}
	zap.ReplaceGlobals(logger)
}
