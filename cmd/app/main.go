package main

import (
	"context"
	"time"

	_ "github.com/savioruz/smrv2-api/docs"
	"github.com/savioruz/smrv2-api/pkg/config"
)

// @title smrv2-api
// @version 0.1
// @description This is an auto-generated API Docs.
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email jakueenak@gmail.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// Create a base context with timeout
	_, baseCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer baseCancel()

	// Remove ChromeDP initialization
	viper := config.NewViper()
	log := config.NewLogrus(viper)
	db := config.NewDatabase(viper, log)
	mail := config.NewGomail(viper, log)
	validate := config.NewValidator()
	jwt := config.NewJWT(viper)
	cache := config.NewRedisClient(viper, log)
	rabbitMQ, err := config.NewRabbitMQ(viper, log)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ: ", err)
	}
	defer rabbitMQ.Close()

	app, log := config.NewFiber()

	err = config.Bootstrap(&config.BootstrapConfig{
		App:       app,
		DB:        db,
		RabbitMQ:  rabbitMQ,
		Logger:    log,
		Validator: validate,
		Viper:     viper,
		Cache:     cache,
		Mail:      mail,
		Jwt:       jwt,
	})
	if err != nil {
		log.Fatal(err)
	}

	port := viper.GetString("APP_PORT")
	log.Info("Starting server on port: ", port)
	go func() {
		if err := app.Listen(":" + port); err != nil {
			log.Fatal(err)
		}
	}()

	config.GracefulShutdown(app, log, 10*time.Second)
}
