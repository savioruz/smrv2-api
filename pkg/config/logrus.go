package config

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func NewLogrus(v *viper.Viper) *logrus.Logger {
	log := logrus.New()
	env := v.GetString("APP_ENV")

	if env == "production" {
		log.SetFormatter(
			&logrus.JSONFormatter{
				TimestampFormat: "2006-01-02 15:04:05",
				FieldMap: logrus.FieldMap{
					logrus.FieldKeyTime:  "@timestamp",
					logrus.FieldKeyLevel: "@level",
					logrus.FieldKeyMsg:   "@message",
					logrus.FieldKeyFunc:  "@caller",
				},
			},
		)
	}

	logLevel := v.GetInt("APP_LOG_LEVEL")
	log.SetLevel(logrus.Level(logLevel))

	return log
}
