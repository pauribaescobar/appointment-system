package main

import (
	"appointment-system/internal/appointments"
	"appointment-system/internal/config"
	"appointment-system/internal/domain"
	"appointment-system/pkg/logger"
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/aws/aws-lambda-go/lambda"
)

var (
	log                 *logger.Logger
	appointmentsService *appointments.AppointmentService
	lambdaConfig        *config.LambdaConfig
)

func init() {
	lambdaConfig = config.InitializeLambdaConfig("appointment-confirmation-sender")
	log = &lambdaConfig.Log
	appointmentsService = appointments.NewAppointmentService(
		lambdaConfig.AppointmentsRepo,
		lambdaConfig.CentersRepo,
		lambdaConfig.ManagementSystemFactory,
		lambdaConfig.MessagingClientFactory,
		lambdaConfig.Cfg.Env,
	)
}

func handleSendConfirmationsErrors(err error) error {
	// We try to parse the error into an application error.
	// If that's the case we can handle the error
	var appErr *domain.AppError
	if errors.As(err, &appErr) {
		log.Error("Send confirmations failed",
			slog.String("op", appErr.Op),
			slog.Any("kind", appErr.Kind),
			slog.Any("cause", appErr.Cause),
			slog.Any("error", err),
		)
	} else {
		log.Error("Send confirmations failed with unknown error",
			slog.Any("error", err),
		)
	}

	switch {
	case errors.Is(err, domain.ErrUnauthorized):
		log.Warn(
			"Unauthorized error - check configuration, skipping retry",
			slog.Any("error", err),
		)
		return nil
	case errors.Is(err, domain.ErrTransient),
		errors.Is(err, domain.ErrRateLimited): // Retry from Lambda / Eventbridge
		return err

	case errors.Is(err, domain.ErrAppointmentAlreadyProcessed),
		errors.Is(err, domain.ErrApointmentNotFound): // Idempotency, normal races
		log.Info("Appointment already handled, skipping")
		return nil

	default: // Internal server errors not handled, for MVP better to throw and take notice
		return err
	}
}

func handleRequest(ctx context.Context, event json.RawMessage) error {
	var appointmentConfirmationSenderEvent appointments.AppointmentConfirmationSenderEvent
	if err := json.Unmarshal(event, &appointmentConfirmationSenderEvent); err != nil {
		log.Error("Failed to unmarshal event",
			slog.Any("error", err),
		)
		return err
	}

	log.WithAttrs(
		slog.String("managementSystem", appointmentConfirmationSenderEvent.ManagementSystem),
		slog.String("messagingClient", appointmentConfirmationSenderEvent.MessagingClient),
		slog.String("centerId", appointmentConfirmationSenderEvent.CenterId),
	)

	err := appointmentsService.SendConfirmations(
		ctx,
		appointmentConfirmationSenderEvent,
	)
	if err == nil {
		log.Info("Appointment confirmations sent successfully")

		return nil
	}

	return handleSendConfirmationsErrors(err)
}

func main() {
	lambdaConfig = config.InitializeLambdaConfig("appointment-confirmation-sender")
	log = &lambdaConfig.Log
	appointmentsService = appointments.NewAppointmentService(
		lambdaConfig.AppointmentsRepo,
		lambdaConfig.CentersRepo,
		lambdaConfig.ManagementSystemFactory,
		lambdaConfig.MessagingClientFactory,
		lambdaConfig.Cfg.Env,
	)
	if lambdaConfig.Cfg.Env == "local" || lambdaConfig.Cfg.Env == "test" {
		handleRequest(
			context.Background(),
			[]byte(`{"managementSystem": "flowww", "messagingClient": "whatsapp", "centerId": "183"}`),
		)
	} else {
		lambda.Start(handleRequest)
	}
}
