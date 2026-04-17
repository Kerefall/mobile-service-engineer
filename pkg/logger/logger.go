package logger

import (
    "os"
    "github.com/sirupsen/logrus"
)

func Init() {
    logrus.SetFormatter(&logrus.TextFormatter{
        FullTimestamp: true,
        ForceColors:   true,
    })
    logrus.SetOutput(os.Stdout)
    logrus.SetLevel(logrus.InfoLevel)
}