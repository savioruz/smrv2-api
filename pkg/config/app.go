package config

import (
	"context"
	"fmt"
	"time"

	cache "github.com/TrinityKnights/Backend/pkg/cache"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/robfig/cron/v3"
	"github.com/savioruz/smrv2-api/internal/builder"
	"github.com/savioruz/smrv2-api/internal/delivery/http"
	"github.com/savioruz/smrv2-api/internal/delivery/messaging"
	gatewayMsg "github.com/savioruz/smrv2-api/internal/gateway/messaging"
	"github.com/savioruz/smrv2-api/internal/repository"
	"github.com/savioruz/smrv2-api/internal/service"
	"github.com/savioruz/smrv2-api/pkg/jwt"
	"github.com/savioruz/smrv2-api/pkg/mail"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type BootstrapConfig struct {
	App       *fiber.App
	DB        *gorm.DB
	RabbitMQ  *amqp.Connection
	Logger    *logrus.Logger
	Validator *validator.Validate
	Cache     *cache.ImplCache
	Viper     *viper.Viper
	Mail      *mail.ImplGomail
	Jwt       *jwt.JWTConfig
}

func Bootstrap(config *BootstrapConfig) error {
	if config.Logger == nil {
		return fmt.Errorf("logger is required")
	}
	if config.DB == nil {
		return fmt.Errorf("database connection is required")
	}
	if config.RabbitMQ == nil {
		return fmt.Errorf("RabbitMQ connection is required")
	}
	if config.Validator == nil {
		return fmt.Errorf("validator is required")
	}
	if config.Cache == nil {
		return fmt.Errorf("redis is required")
	}
	if config.Viper == nil {
		return fmt.Errorf("viper configuration is required")
	}
	if config.Mail == nil {
		return fmt.Errorf("mail is required")
	}
	if config.Jwt == nil {
		return fmt.Errorf("jwt is required")
	}

	// Initialize JWT service
	jwtService := jwt.NewJWTService(config.Jwt)

	// Initialize repositories
	userRepository := repository.NewUsersRepository(config.DB, config.Logger)
	studyRepository := repository.NewStudyRepository(config.DB, config.Logger)
	scrapedScheduleRepository := repository.NewScrapedScheduleRepository(config.DB, config.Logger)
	userScheduleRepository := repository.NewUserScheduleRepository(config.DB, config.Logger)
	subscriptionRepository := repository.NewSubscriptionRepository(config.DB, config.Logger)

	// Initialize producers
	producer := gatewayMsg.NewProducer(config.RabbitMQ, config.Logger)
	userProducer := gatewayMsg.NewUserProducer(producer, config.Logger)
	scrapedScheduleProducer := gatewayMsg.NewScrapedScheduleProducer(producer, config.Logger)
	mailProducer := gatewayMsg.NewMailProducer(producer, config.Logger)

	// Initialize services
	userService := service.NewUserService(
		config.DB,
		config.Logger,
		config.Validator,
		config.Viper,
		userRepository,
		subscriptionRepository,
		mailProducer,
		userProducer,
		config.Mail,
		jwtService,
	)
	studyService := service.NewStudyService(
		config.DB,
		config.Logger,
		config.Validator,
		userRepository,
		studyRepository,
		userScheduleRepository,
		userProducer,
	)
	scrapedScheduleService := service.NewScrapedScheduleService(
		config.DB,
		config.Logger,
		config.Validator,
		config.Cache,
		scrapedScheduleRepository,
		userRepository,
		config.Mail,
	)

	scheduleOrchestratorService := service.NewScheduleOrchestratorService(
		config.Logger,
		scrapedScheduleService,
		scrapedScheduleProducer,
	)

	// Update scrapedScheduleService with orchestrator
	scrapedScheduleService.SetOrchestrator(scheduleOrchestratorService)

	userScheduleService := service.NewUserScheduleService(
		config.DB,
		config.Logger,
		config.Validator,
		config.Viper,
		config.Cache,
		userScheduleRepository,
		userRepository,
		userProducer,
		config.Mail,
	)

	// Initialize HTTP handlers
	userHandler := http.NewUserHandler(config.Logger, config.Viper, userService)
	userScheduleHandler := http.NewUserScheduleHandler(config.Logger, userScheduleService)
	scrapedScheduleHandler := http.NewScrapedScheduleHandler(config.Logger, scrapedScheduleService)

	// Initialize message consumers
	consumer := messaging.NewConsumer(config.RabbitMQ, config.Logger)
	userConsumer := messaging.NewUserConsumer(
		consumer,
		config.Logger,
		studyService,
		config.Mail,
		mailProducer,
	)
	scrapedScheduleConsumer := messaging.NewScrapedScheduleConsumer(
		consumer,
		config.Logger,
		scrapedScheduleService,
		scrapedScheduleProducer,
		mailProducer,
	)

	// Start consumers
	go func() {
		if err := scrapedScheduleConsumer.ConsumeMetadata(context.Background()); err != nil {
			config.Logger.Errorf("Metadata consumer failed: %v", err)
		}
	}()

	go func() {
		if err := userConsumer.ConsumeEmailSending(context.Background()); err != nil {
			config.Logger.Errorf("Email sending consumer failed: %v", err)
		}
	}()

	go func() {
		if err := scrapedScheduleConsumer.ConsumeScrapeRequests(context.Background()); err != nil {
			config.Logger.Errorf("Scrape worker failed: %v", err)
		}
	}()

	go func() {
		if err := userConsumer.ConsumeStudyData(context.Background()); err != nil {
			config.Logger.Errorf("Study data consumer failed: %v", err)
		}
	}()

	go func() {
		if err := scrapedScheduleConsumer.ConsumeSaveRequests(context.Background()); err != nil {
			config.Logger.Errorf("Save worker failed: %v", err)
		}
	}()

	go func() {
		// Initialize cron scheduler for periodic scraping
		c := cron.New(cron.WithSeconds(), cron.WithLocation(time.FixedZone("WIB", 7*60*60)))
		_, err := c.AddFunc("0 0 0 */7 * *", func() {
			if err := scheduleOrchestratorService.RunScheduleScraping(context.Background(), nil); err != nil {
				config.Logger.Errorf("Scheduled scraping failed: %v", err)
			}
		})
		if err != nil {
			config.Logger.Errorf("Failed to schedule scraping job: %v", err)
		}

		// Start the cron scheduler
		c.Start()
	}()

	// Build HTTP routes
	builder := builder.Config{
		App:                    config.App,
		Jwt:                    jwtService,
		UserHandler:            userHandler,
		UserScheduleHandler:    userScheduleHandler,
		ScrapedScheduleHandler: scrapedScheduleHandler,
	}

	builder.Build()

	config.Logger.Info("Application is ready to use...")
	return nil
}
