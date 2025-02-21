package config

import (
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func NewLogrus(v *viper.Viper) *logrus.Logger {
	log := logrus.New()
	env := v.GetString("APP_ENV")

	if strings.ToUpper(env) == "PRODUCTION" {
		log.SetFormatter(
			&logrus.JSONFormatter{},
		)
	}

	logLevel := v.GetInt("APP_LOG_LEVEL")
	log.SetLevel(logrus.Level(logLevel))

	return log
}
