package pkg

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const logFilePath = "./frontendlogfile"

func SetupLogger() (*zap.Logger, *os.File) {
    file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        panic(err)
    }

    fileCore := zapcore.NewCore(
        zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
        zapcore.AddSync(file),                                   
        zap.InfoLevel,                                           
    )

    consoleCore := zapcore.NewCore(
        zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()), 
        zapcore.AddSync(os.Stdout),                                   
        zap.InfoLevel,                                                
    )

    combinedCore := zapcore.NewTee(fileCore, consoleCore)

    logger := zap.New(combinedCore, zap.AddCaller())

    return logger, file
}
