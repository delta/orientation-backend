package config

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Log is a instance of logrus.Logger, this will be used for logging
var Log = logrus.New()

func initLogger() {
	// overriding default(stderr) to stdout
	Log.Out = os.Stdout

	Log.Formatter = &logrus.TextFormatter{
		ForceColors: true,
	}
}
