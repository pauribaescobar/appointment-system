package domain

import (
	"context"
)

type MessagingClientOption string

const (
	WhatsappClient MessagingClientOption = "whatsapp"
)

type MessagingClient interface {
	SendConfirmationMessage(
		ctx context.Context,
		center *Center,
		appointment *Appointment,
	) (string, error)
	SendConfirmationAck(
		ctx context.Context,
		center *Center,
		appointment *Appointment,
	) error
	SendCancellationAck(
		ctx context.Context,
		center *Center,
		appointment *Appointment,
	) error
}

type MessagingClientFactory interface {
	GetMessagingClient(client MessagingClientOption) MessagingClient
}
