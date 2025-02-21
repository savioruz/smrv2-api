package config

import (
	"context"
	"fmt"

	cache "github.com/TrinityKnights/Backend/pkg/cache"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// NewRedisClient creates a new redis client
func NewRedisClient(viper *viper.Viper, log *logrus.Logger) *cache.ImplCache {
	addr := fmt.Sprintf("%s:%s", viper.GetString("REDIS_HOST"), viper.GetString("REDIS_PORT"))
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: viper.GetString("REDIS_PASSWORD"),
		DB:       viper.GetInt("REDIS_DB"),
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("failed to connect redis: %v", err)
	}

	return cache.NewCache(client)
}
