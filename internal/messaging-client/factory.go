package messagingclient

import (
	"appointment-system/internal/domain"
	"appointment-system/internal/messaging-client/whatsapp"
	"os"
)

type messagingClientFactory struct{}

var msgClientFactory *messagingClientFactory

func GetMessagingClientFactory() *messagingClientFactory {
	if msgClientFactory != nil {
		return msgClientFactory
	}

	return &messagingClientFactory{}
}

func (f *messagingClientFactory) GetMessagingClient(client domain.MessagingClientOption) domain.MessagingClient {
	switch client {
	case domain.WhatsappClient:
		return whatsapp.NewWhatsappClient(
			os.Getenv("WHATSAPP_AUTH_TOKEN"),
		)
	default:
		return nil
	}
}
