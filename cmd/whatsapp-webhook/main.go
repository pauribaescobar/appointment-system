package main

import (
	"appointment-system/internal/appointments"
	"appointment-system/internal/config"
	"appointment-system/internal/domain"
	"appointment-system/internal/messaging-client/whatsapp"
	"appointment-system/pkg/logger"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var (
	log                 *logger.Logger
	lambdaConfig        *config.LambdaConfig
	appointmentsService *appointments.AppointmentService
	verifyToken         string
)

func init() {
	lambdaConfig = config.InitializeLambdaConfig("whatsapp-webhook")
	log = &lambdaConfig.Log
	verifyToken = lambdaConfig.Cfg.WebhookVerifyToken

	appointmentsService = appointments.NewAppointmentService(
		lambdaConfig.AppointmentsRepo,
		lambdaConfig.CentersRepo,
		lambdaConfig.ManagementSystemFactory,
		lambdaConfig.MessagingClientFactory,
		lambdaConfig.Cfg.Env,
	)
}

func handleWebhookErrors(err error) {
	var appErr *domain.AppError
	if errors.As(err, &appErr) {
		log.Error("Webhook processing failed",
			slog.String("op", appErr.Op),
			slog.Any("kind", appErr.Kind),
			slog.Any("cause", appErr.Cause),
			slog.Any("error", err),
		)
	} else {
		log.Error("Webhook processing failed with unknown error",
			slog.Any("error", err),
		)
	}

	switch {
	case errors.Is(err, domain.ErrUnauthorized):
		log.Warn("Unauthorized error - check configuration",
			slog.Any("error", err),
		)

	case errors.Is(err, domain.ErrTransient),
		errors.Is(err, domain.ErrRateLimited):
		log.Warn("Transient/rate-limited error - Meta will retry delivery",
			slog.Any("error", err),
		)

	case errors.Is(err, domain.ErrAppointmentAlreadyProcessed),
		errors.Is(err, domain.ErrApointmentNotFound):
		log.Info("Appointment already handled, skipping")

	case errors.Is(err, domain.ErrInvalidInput):
		log.Warn("Invalid input in webhook payload",
			slog.Any("error", err),
		)
	}
}

func handleRequest(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	log.Info("Webhook request received",
		slog.String("method", req.RequestContext.HTTP.Method),
	)

	switch req.RequestContext.HTTP.Method {
	case http.MethodGet:
		return handleVerification(req)
	case http.MethodPost:
		return handleIncomingMessage(ctx, req)
	default:
		return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusMethodNotAllowed}, nil
	}
}

func handleVerification(req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	mode := req.QueryStringParameters["hub.mode"]
	token := req.QueryStringParameters["hub.verify_token"]
	challenge := req.QueryStringParameters["hub.challenge"]

	if mode == "subscribe" && token == verifyToken {
		log.Info("Webhook verification succeeded")
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusOK,
			Body:       challenge,
		}, nil
	}

	log.Warn("Webhook verification failed",
		slog.String("mode", mode),
	)
	return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusForbidden}, nil
}

func handleIncomingMessage(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	log.Info("Raw webhook body received",
		slog.String("body", req.Body),
	)

	var webhookEvent whatsapp.WebhookEvent
	if err := json.Unmarshal([]byte(req.Body), &webhookEvent); err != nil {
		log.Error("Failed to unmarshal webhook event",
			slog.Any("error", err),
		)
		return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusOK}, nil
	}

	log.Info("Webhook event parsed, calling HandleWebhookEvent")

	err := appointmentsService.HandleWebhookEvent(ctx, webhookEvent)
	if err != nil {
		handleWebhookErrors(err)
	} else {
		log.Info("Webhook event processed successfully")
	}

	return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusOK}, nil
}

func main() {
	lambdaConfig = config.InitializeLambdaConfig("whatsapp-webhook")
	log = &lambdaConfig.Log
	verifyToken = lambdaConfig.Cfg.WebhookVerifyToken

	appointmentsService = appointments.NewAppointmentService(
		lambdaConfig.AppointmentsRepo,
		lambdaConfig.CentersRepo,
		lambdaConfig.ManagementSystemFactory,
		lambdaConfig.MessagingClientFactory,
		lambdaConfig.Cfg.Env,
	)

	if lambdaConfig.Cfg.Env == "local" {
		testPayload := `{"object":"whatsapp_business_account","entry":[{"id":"test","changes":[{"value":{"messaging_product":"whatsapp","metadata":{"display_phone_number":"123","phone_number_id":"456"},"messages":[{"from":"34644409148","id":"wamid.test","timestamp":"1234567890","type":"interactive","interactive":{"type":"button_reply","button_reply":{"id":"confirm","title":"CONFIRMO"}},"context":{"from":"123","id":"wamid.original"}}]},"field":"messages"}]}]}`
		req := events.APIGatewayV2HTTPRequest{
			RequestContext: events.APIGatewayV2HTTPRequestContext{
				HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
					Method: http.MethodPost,
				},
			},
			Body: testPayload,
		}
		handleRequest(context.Background(), req)
	} else {
		lambda.Start(handleRequest)
	}
}
