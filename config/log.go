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

	Log.SetLevel(logrus.DebugLevel)

	Log.Formatter = &logrus.TextFormatter{
		ForceColors: true,
	}

	// prod config

	// if Config("ENV") == "production" {
	// 	file, err := os.OpenFile("ws.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	Log.Formatter = &logrus.JSONFormatter{}

	// 	Log.Out = file

	// 	Log.SetLevel(logrus.ErrorLevel)
	// }
}
