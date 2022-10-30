package logger

import "github.com/hashicorp/go-hclog"

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
			Name:  "oidc-client",
			Level: logLevel,
		})
	} else {
		appLogger = hclog.New(&hclog.LoggerOptions{
			Name:  "oidc-client",
			Level: logLevel,
			Color: hclog.AutoColor,
		})
	}

	return appLogger
}
