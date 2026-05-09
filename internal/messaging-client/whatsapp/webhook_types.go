package whatsapp

type WebhookEvent struct {
	Object string              `json:"object"`
	Entry  []WebhookEventEntry `json:"entry"`
}

type WebhookEventEntry struct {
	ID      string                    `json:"id"`
	Changes []WebhookEventEntryChange `json:"changes"`
}

type WebhookEventEntryChange struct {
	Value WebhookEventChangeValue `json:"value"`
	Field string                  `json:"field"`
}

type WebhookEventChangeValue struct {
	MessagingProduct string                  `json:"messaging_product"`
	Metadata         WebhookEventMetadata    `json:"metadata"`
	Contacts         []WebhookEventContact   `json:"contacts,omitempty"`
	Messages         []WebhookEventMessage   `json:"messages,omitempty"`
	Statuses         []WebhookEventStatus    `json:"statuses,omitempty"`
}

type WebhookEventMetadata struct {
	DisplayPhoneNumber string `json:"display_phone_number"`
	PhoneNumberID      string `json:"phone_number_id"`
}

type WebhookEventContact struct {
	Profile WebhookEventContactProfile `json:"profile"`
	WaID    string                     `json:"wa_id"`
}

type WebhookEventContactProfile struct {
	Name string `json:"name"`
}

type WebhookEventMessage struct {
	From        string                       `json:"from"`
	ID          string                       `json:"id"`
	Timestamp   string                       `json:"timestamp"`
	Type        string                       `json:"type"`
	Interactive *WebhookEventInteractive     `json:"interactive,omitempty"`
	Button      *WebhookEventButton          `json:"button,omitempty"`
	Context     *WebhookEventMessageContext  `json:"context,omitempty"`
}

type WebhookEventButton struct {
	Payload string `json:"payload"`
	Text    string `json:"text"`
}

type WebhookEventInteractive struct {
	Type        string                          `json:"type"`
	ButtonReply *WebhookEventInteractiveButton  `json:"button_reply,omitempty"`
}

type WebhookEventInteractiveButton struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type WebhookEventMessageContext struct {
	From string `json:"from"`
	ID   string `json:"id"`
}

// Status updates (delivery receipts) -- parsed but ignored by the webhook
type WebhookEventStatus struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	Timestamp   string `json:"timestamp"`
	RecipientID string `json:"recipient_id"`
}
