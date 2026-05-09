package appointments

import (
	"appointment-system/internal/centers"
	"appointment-system/internal/domain"
	"appointment-system/internal/messaging-client/whatsapp"
	"context"
	"fmt"
	"time"
)

const (
	ButtonTitleConfirm = "CONFIRMO"
	ButtonTitleCancel  = "NO PUEDO ASISTIR"
)

type AppointmentService struct {
	appointmentsRepo        AppointmentRepository
	centersRepo             centers.CentersRepository
	managementSystemFactory domain.AppointmentSystemFactory
	messagingClientFactory  domain.MessagingClientFactory
	env                     string
}

func NewAppointmentService(
	appointmentsRepo AppointmentRepository,
	centersRepo centers.CentersRepository,
	asFactory domain.AppointmentSystemFactory,
	mcFactory domain.MessagingClientFactory,
	env string,
) *AppointmentService {
	return &AppointmentService{
		appointmentsRepo:        appointmentsRepo,
		centersRepo:             centersRepo,
		managementSystemFactory: asFactory,
		messagingClientFactory:  mcFactory,
		env:                     env,
	}
}

func (s *AppointmentService) SendConfirmations(ctx context.Context,
	event AppointmentConfirmationSenderEvent,
) error {
	// Attain appointments and send confirmations logic would go here
	system := s.managementSystemFactory.Build(event.ManagementSystem)
	msgClient := s.messagingClientFactory.GetMessagingClient(
		domain.MessagingClientOption(event.MessagingClient),
	)

	center, err := s.centersRepo.GetByID(ctx, event.CenterId)
	if err != nil {
		return fmt.Errorf(
			"AppointmentService.SendConfirmations: Error retrieving center data:%w",
			err,
		)
	}

	appointments, err := system.ListAppointments(ctx, event.CenterId)
	if err != nil {
		return fmt.Errorf("AppointmentService.SendConfirmations: Error listing appointments:%w", err)
	}

	// In test mode, take only the first appointment and override the phone number
	if s.env == "test" && len(appointments) > 0 {
		appointments = appointments[5:6]
		appointments[0].CustomerPhoneNumber = "+34644409148"
	}

	// Once we have got the appointments we should persist them into DynamoDB and send the confirmation message
	for _, appointment := range appointments {
		appointment.CustomerPhoneNumber = "+34644409148"
		appointment.ConfirmationExpirationTimestamp = time.Now().Add(2 * 24 * time.Hour).Unix()
		existingAp, err := s.appointmentsRepo.PutIfNotExists(ctx, &appointment)
		if err != nil {
			return fmt.Errorf("AppointmentService.SendConfirmations: Error persisting appointment:%w", err)
		}

		// If we already persisted the appointment before we skip
		if existingAp != nil && existingAp.Status == domain.AppoinmentStatusSent {
			continue
		}

		// Sending appointment confirmation message via whatsapp using whatsapp client
		messageId, err := msgClient.SendConfirmationMessage(
			ctx,
			center,
			&appointment,
		)
		if err != nil {
			return fmt.Errorf("AppointmentService.SendConfirmations: Error sending confirmation message:%w", err)
		}

		// Mark the message as sent
		err = s.appointmentsRepo.MarkSent(
			ctx,
			appointment.ID,
			messageId,
			event.MessagingClient,
		)
		if err != nil {
			return fmt.Errorf("AppointmentService.SendConfirmations: Error marking confirmation as sent:%w", err)
		}
	}

	return nil
}

func (s *AppointmentService) HandleWebhookEvent(ctx context.Context, event whatsapp.WebhookEvent) error {
	for _, entry := range event.Entry {
		for _, change := range entry.Changes {
			if change.Field != "messages" {
				continue
			}
			for _, msg := range change.Value.Messages {
				if err := s.processWebhookReply(ctx, msg); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *AppointmentService) processWebhookReply(ctx context.Context, msg whatsapp.WebhookEventMessage) error {
	const op = "AppointmentService.ProcessWebhookReply"

	fmt.Println("[webhook] processWebhookReply called, msg.Type:", msg.Type)

	var buttonTitle string
	switch msg.Type {
	case "interactive":
		if msg.Interactive == nil || msg.Interactive.Type != "button_reply" || msg.Interactive.ButtonReply == nil {
			fmt.Println("[webhook] skipping: interactive but not a button_reply")
			return nil
		}
		buttonTitle = msg.Interactive.ButtonReply.Title
	case "button":
		if msg.Button == nil {
			fmt.Println("[webhook] skipping: button type but no button payload")
			return nil
		}
		buttonTitle = msg.Button.Text
	default:
		fmt.Println("[webhook] skipping: unsupported msg type:", msg.Type)
		return nil
	}

	if msg.Context == nil || msg.Context.ID == "" {
		return fmt.Errorf("%s: Button reply without context message_id:%w", op, domain.ErrInvalidInput)
	}

	originalMessageID := msg.Context.ID
	fmt.Println("[webhook] button reply received, originalMessageID:", originalMessageID, "buttonTitle:", buttonTitle)

	appointment, err := s.appointmentsRepo.GetByMessageID(ctx, originalMessageID)
	if err != nil {
		return fmt.Errorf("%s: Error retrieving appointment by message_id:%w", op, err)
	}
	fmt.Println("[webhook] appointment found, ID:", appointment.ID, "centerID:", appointment.CenterID)

	center, err := s.centersRepo.GetByID(ctx, appointment.CenterID)
	if err != nil {
		return fmt.Errorf("%s: Error retrieving center data:%w", op, err)
	}
	fmt.Println("[webhook] center found, ID:", center.ID)

	system := s.managementSystemFactory.Build("flowww")
	if system == nil {
		return fmt.Errorf("%s: Failed to build management system:%w", op, domain.ErrInternal)
	}

	msgClient := s.messagingClientFactory.GetMessagingClient(domain.WhatsappClient)

	switch buttonTitle {
	case ButtonTitleConfirm:
		fmt.Println("[webhook] confirming appointment on Flowww...")
		if err := system.ConfirmAppointment(ctx, appointment.CenterID, appointment.ID, appointment.Date); err != nil {
			return fmt.Errorf("%s: Error confirming appointment on Flowww:%w", op, err)
		}
		fmt.Println("[webhook] confirmed on Flowww, sending ack...")
		if err := msgClient.SendConfirmationAck(ctx, center, appointment); err != nil {
			return fmt.Errorf("%s: Error sending confirmation ack:%w", op, err)
		}
		fmt.Println("[webhook] confirmation ack sent")

	case ButtonTitleCancel:
		fmt.Println("[webhook] cancelling appointment on Flowww...")
		if err := system.CancelAppointment(ctx, appointment.CenterID, appointment.ID, appointment.Date); err != nil {
			return fmt.Errorf("%s: Error cancelling appointment on Flowww:%w", op, err)
		}
		fmt.Println("[webhook] cancelled on Flowww, sending ack...")
		if err := msgClient.SendCancellationAck(ctx, center, appointment); err != nil {
			return fmt.Errorf("%s: Error sending cancellation ack:%w", op, err)
		}
		fmt.Println("[webhook] cancellation ack sent")

	default:
		return fmt.Errorf("%s: Unknown button reply title '%s':%w", op, buttonTitle, domain.ErrInvalidInput)
	}

	fmt.Println("[webhook] deleting appointment", appointment.ID)
	if err := s.appointmentsRepo.Delete(ctx, appointment.ID); err != nil {
		return fmt.Errorf("%s: Error deleting appointment after processing:%w", op, err)
	}
	fmt.Println("[webhook] done, appointment deleted")

	return nil
}
