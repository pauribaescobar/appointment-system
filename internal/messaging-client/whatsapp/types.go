package whatsapp

type WhatsappTemplatePayloadLanguage struct {
	Code string `json:"code"`
}

type WhatsappTemplatePayloadComponentParameter struct {
	Type          string `json:"type"`
	ParameterName string `json:"parameter_name,omitempty"`
	Text          string `json:"text,omitempty"`
}

type WhatsappTemplatePayloadComponent struct {
	Type       string                                      `json:"type"`
	Parameters []WhatsappTemplatePayloadComponentParameter `json:"parameters,omitempty"`
}

type WhatsappTemplatePayload struct {
	Name       string                             `json:"name"`
	Language   WhatsappTemplatePayloadLanguage    `json:"language"`
	Components []WhatsappTemplatePayloadComponent `json:"components,omitempty"`
}

type WhatsappMessageRequest struct {
	MessagingProduct string                   `json:"messaging_product"`
	To               string                   `json:"to"`
	Type             string                   `json:"type"`
	Template         *WhatsappTemplatePayload `json:"template"`
}

type WhatsappMessageResponseContact struct {
	Input      string `json:"input"`
	WathsappId string `json:"wa_id"`
}

type WhatsappMessageResponseMessage struct {
	ID            string `json:"id"`
	GroupID       string `json:"group_id"`
	MessageStatus string `json:"message_status"`
}

type WhatsappMessageResponse struct {
	MessagingProduct string                           `json:"messaging_product"`
	Contacts         []WhatsappMessageResponseContact `json:"contacts"`
	Messages         []WhatsappMessageResponseMessage `json:"messages"`
}

type WhatsappMessageErrorData struct {
	MessagingProduct string `json:"messaging_product"`
	Details          string `json:"details"`
}

type WhatsappMessageError struct {
	Message      string                    `json:"message"`
	Type         string                    `json:"type"`
	Code         int                       `json:"code"`
	ErrorData    *WhatsappMessageErrorData `json:"error_data"`
	ErrorSubcode int                       `json:"error_subcode"`
	FBTraceID    string                    `json:"fbtrace_id"`
}

type WhatsappMessageErrorResponse struct {
	Error WhatsappMessageError `json:"error"`
}
