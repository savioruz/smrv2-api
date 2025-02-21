package config

import (
	"strings"
	"time"

	"github.com/savioruz/smrv2-api/pkg/jwt"
	"github.com/spf13/viper"
)

func parseDuration(s string) time.Duration {
	if strings.HasSuffix(s, "d") {
		days := strings.TrimSuffix(s, "d")
		if d, err := time.ParseDuration(days + "h"); err == nil {
			return d * 24
		}
	}
	d, _ := time.ParseDuration(s)
	return d
}

func NewJWT(viper *viper.Viper) *jwt.JWTConfig {
	return &jwt.JWTConfig{
		Secret:        viper.GetString("JWT_SECRET"),
		AccessExpiry:  parseDuration(viper.GetString("JWT_ACCESS_EXPIRY")),
		RefreshExpiry: parseDuration(viper.GetString("JWT_REFRESH_EXPIRY")),
	}
}
