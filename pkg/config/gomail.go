package config

import (
	"github.com/savioruz/smrv2-api/pkg/mail"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/gomail.v2"
)

// NewGomail creates a new gomail client
func NewGomail(viper *viper.Viper, log *logrus.Logger) *mail.ImplGomail {
	dialer := gomail.NewDialer(
		viper.GetString("SMTP_HOST"),
		viper.GetInt("SMTP_PORT"),
		viper.GetString("SMTP_USERNAME"), // fromEmail
		viper.GetString("SMTP_PASSWORD"),
	)

	if _, err := dialer.Dial(); err != nil {
		log.Fatalf("failed to connect to SMTP server: %v", err)
	}

	return mail.NewGomail(dialer, viper.GetString("SMTP_USERNAME"))
}
