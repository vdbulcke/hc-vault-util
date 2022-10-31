package logger

import (
	"os"

	"github.com/hashicorp/go-hclog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// GenLogger generate logger
func GenLogger(Debug, noColor bool) hclog.Logger {
	// Create Logger
	var appLogger hclog.Logger

	logLevel := hclog.LevelFromString("INFO")

	// from arg
	if Debug {
		logLevel = hclog.LevelFromString("DEBUG")
	}

	if noColor {
		appLogger = hclog.New(&hclog.LoggerOptions{

			Level: logLevel,
		})
	} else {
		appLogger = hclog.New(&hclog.LoggerOptions{

			Level: logLevel,
			Color: hclog.AutoColor,
		})
	}

	return appLogger
}

// GetZapLogger returns a zap.Logger
func GetZapLogger(Debug bool, noColor bool) *zap.Logger {

	// Zap Logger
	var logger *zap.Logger
	var core zapcore.Core

	// override time format
	zapConfig := zap.NewProductionEncoderConfig()
	if !noColor {
		zapConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		zapConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	}

	zapConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// using console encoder since CLI tool
	consoleEncoder := zapcore.NewConsoleEncoder(zapConfig)

	// default writer for logger
	consoleDebugging := zapcore.Lock(os.Stdout)
	consoleErrors := zapcore.Lock(os.Stderr)
	if Debug {
		// set log level to writer
		core = zapcore.NewTee(
			zapcore.NewCore(consoleEncoder, consoleDebugging, zap.DebugLevel),
			zapcore.NewCore(consoleEncoder, consoleDebugging, zap.InfoLevel),
			zapcore.NewCore(consoleEncoder, consoleErrors, zap.WarnLevel),
			zapcore.NewCore(consoleEncoder, consoleErrors, zap.ErrorLevel),
		)
	} else {

		// set log level to writer
		core = zapcore.NewTee(
			zapcore.NewCore(consoleEncoder, consoleDebugging, zap.InfoLevel),
			zapcore.NewCore(consoleEncoder, consoleErrors, zap.WarnLevel),
			zapcore.NewCore(consoleEncoder, consoleErrors, zap.ErrorLevel),
		)

	}

	// add function caller and stack trace on error
	logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger

}
