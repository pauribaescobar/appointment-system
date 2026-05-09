package config

import (
	"appointment-system/internal/appointments"
	"appointment-system/internal/centers"
	"appointment-system/internal/domain"
	"appointment-system/internal/dynamo"
	managementsystem "appointment-system/internal/management-system"
	messagingclient "appointment-system/internal/messaging-client"
	"appointment-system/pkg/logger"
	"context"
	"log/slog"
	"os"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	lambdaTypes "github.com/aws/aws-sdk-go-v2/service/lambda"
)

type Config struct {
	AWSRegion                     string
	DynamoDBAppointmentsTableName string
	DynamoDBCentersTableName      string
	Env                           string
	PhoneHashSalt                 string
	LogLevel                      string
	WebhookVerifyToken            string
}

func Load() *Config {
	return &Config{
		AWSRegion:                     os.Getenv("AWS_REGION"),
		DynamoDBAppointmentsTableName: os.Getenv("DYNAMODB_APPOINTMENTS_TABLE"),
		DynamoDBCentersTableName:      os.Getenv("DYNAMODB_CENTERS_TABLE"),
		Env:                           os.Getenv("ENV"),
		PhoneHashSalt:                 os.Getenv("PHONE_HASH_SALT"),
		LogLevel:                      os.Getenv("LOG_LEVEL"),
		WebhookVerifyToken:            os.Getenv("WEBHOOK_VERIFY_TOKEN"),
	}
}

type LambdaConfig struct {
	Cfg                     Config
	Log                     logger.Logger
	ManagementSystemFactory domain.AppointmentSystemFactory
	AppointmentsRepo        appointments.AppointmentRepository
	CentersRepo             centers.CentersRepository
	MessagingClientFactory  domain.MessagingClientFactory
}

func InitializeLambdaConfig(serviceName string) *LambdaConfig {
	cfg := *Load()

	log := *logger.NewLogger(logger.LoggerConfig{
		Level:   cfg.LogLevel,
		Service: serviceName,
	})

	log.Info("Logger Initialized")

	awsCfg, err := awsConfig.LoadDefaultConfig(
		context.TODO(),
		awsConfig.WithRegion(cfg.AWSRegion),
	)

	if err != nil {
		log.Error("Failed to load AWS config",
			slog.Any("error", err),
		)
		panic(err)
	}

	lambdaClient := lambdaTypes.NewFromConfig(awsCfg)
	log.Info("Lambda Client Initialized")

	dynamoClient := dynamodb.NewFromConfig(awsCfg)
	log.Info("DynamoDB Client Initialized")

	appointmentsRepo := dynamo.NewDynamoDBAppointmentsRepo(
		dynamoClient,
		cfg.DynamoDBAppointmentsTableName,
		log,
	)
	log.Info("Appointments Repository Initialized")

	centersRepo := dynamo.NewDynamoDBCentersRepo(
		dynamoClient,
		cfg.DynamoDBCentersTableName,
		log,
	)
	log.Info("Centers Repository Initialized")

	messagingClientFactory := messagingclient.GetMessagingClientFactory()

	managementSystemFactory := managementsystem.NewManagementSystemFactory(
		lambdaClient,
	)
	return &LambdaConfig{
		Cfg:                     cfg,
		Log:                     log,
		ManagementSystemFactory: managementSystemFactory,
		AppointmentsRepo:        appointmentsRepo,
		CentersRepo:             centersRepo,
		MessagingClientFactory:  messagingClientFactory,
	}
}
