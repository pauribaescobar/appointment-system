package whatsapp

import (
	"appointment-system/internal/domain"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	WhatsappMessagingProduct                string = "whatsapp"
	WhatsappMessageRequestTemplateType      string = "template"
	WhatsappAppointmentConfirmqtionTemplate string = "appointment_confirmation_v1"
	WhatsappAppointmentConfirmedTemplate    string = "appointment_confirmed_v1"
	WhatsappAppointmentCancelledTemplate    string = "appointment_cancelled_v1"
	WhatsappTemplateLanguageCodeSpanish     string = "es_ES"
	WhatsappTemplateLanguageCodeEnglish     string = "en"
	MetaGraphApiVersion                     string = "v25.0"
)

type WhatsappClient struct {
	authToken string
}

func NewWhatsappClient(token string) *WhatsappClient {
	return &WhatsappClient{
		authToken: token,
	}
}

func (c *WhatsappClient) SendConfirmationMessage(
	ctx context.Context,
	center *domain.Center,
	appointment *domain.Appointment,
) (string, error) {
	const op = "WhatsappClient.SendConfirmationMessage"
	// Build Confirmation Message from appointment
	payload, err := c.buildConfirmationMessagePayload(center, appointment)
	if err != nil {
		return "", domain.E(op, domain.ErrInternal, err)
	}

	body := bytes.NewReader(payload)

	if center.WhatsappBusinessPhoneID == nil {
		return "", domain.E(op, domain.ErrInvalidInput, fmt.Errorf("center %s has no whatsapp business phone ID configured", center.ID))
	}

	endpoint := fmt.Sprintf(
		"https://graph.facebook.com/%s/%s/messages",
		MetaGraphApiVersion,
		*center.WhatsappBusinessPhoneID,
	)
	req, err := http.NewRequest("POST", endpoint, body)
	if err != nil {
		return "", domain.E(op, domain.ErrInternal, err)
	}
	req.Header.Set("Authorization", "Bearer "+c.authToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return "", domain.E(op, domain.ErrInternal, fmt.Errorf("read whatsapp response: %w", readErr))
	}

	if resp.StatusCode != http.StatusOK {
		return "", c.mapWhatsAppError(op, resp.StatusCode, bodyBytes)
	}

	var whatsappMessageResponse WhatsappMessageResponse
	err = json.Unmarshal(bodyBytes, &whatsappMessageResponse)
	if err != nil {
		return "", domain.E(op, domain.ErrInternal, fmt.Errorf("unmarshal whatsapp response: %w", err))
	}

	if len(whatsappMessageResponse.Messages) == 0 || len(whatsappMessageResponse.Contacts) == 0 {
		return "", domain.E(op, domain.ErrInternal, fmt.Errorf("invalid whatsapp success response"))
	}

	return whatsappMessageResponse.Messages[0].ID, nil
}

func (c *WhatsappClient) SendConfirmationAck(
	ctx context.Context,
	center *domain.Center,
	appointment *domain.Appointment,
) error {
	const op = "WhatsappClient.SendConfirmationAck"

	req := WhatsappMessageRequest{
		MessagingProduct: WhatsappMessagingProduct,
		To:               appointment.CustomerPhoneNumber,
		Type:             WhatsappMessageRequestTemplateType,
		Template: &WhatsappTemplatePayload{
			Name:     WhatsappAppointmentConfirmedTemplate,
			Language: WhatsappTemplatePayloadLanguage{Code: WhatsappTemplateLanguageCodeEnglish},
			Components: []WhatsappTemplatePayloadComponent{
				{
					Type: "body",
					Parameters: []WhatsappTemplatePayloadComponentParameter{
						{Type: "text", ParameterName: "customer_name", Text: appointment.CustomerName},
						{Type: "text", ParameterName: "center", Text: center.Name},
						{Type: "text", ParameterName: "day", Text: appointment.Date},
						{Type: "text", ParameterName: "start_time", Text: appointment.StartTime},
						{Type: "text", ParameterName: "end_time", Text: appointment.EndTime},
						{Type: "text", ParameterName: "address", Text: center.Address},
					},
				},
			},
		},
	}

	return c.sendTemplateMessage(op, center, req)
}

func (c *WhatsappClient) SendCancellationAck(
	ctx context.Context,
	center *domain.Center,
	appointment *domain.Appointment,
) error {
	const op = "WhatsappClient.SendCancellationAck"

	req := WhatsappMessageRequest{
		MessagingProduct: WhatsappMessagingProduct,
		To:               appointment.CustomerPhoneNumber,
		Type:             WhatsappMessageRequestTemplateType,
		Template: &WhatsappTemplatePayload{
			Name:     WhatsappAppointmentCancelledTemplate,
			Language: WhatsappTemplatePayloadLanguage{Code: WhatsappTemplateLanguageCodeSpanish},
			Components: []WhatsappTemplatePayloadComponent{
				{
					Type: "body",
					Parameters: []WhatsappTemplatePayloadComponentParameter{
						{Type: "text", ParameterName: "customer_name", Text: appointment.CustomerName},
						{Type: "text", ParameterName: "center", Text: center.Name},
						{Type: "text", ParameterName: "day", Text: appointment.Date},
					},
				},
			},
		},
	}

	return c.sendTemplateMessage(op, center, req)
}

func (c *WhatsappClient) sendTemplateMessage(
	op string,
	center *domain.Center,
	req WhatsappMessageRequest,
) error {
	payload, err := json.Marshal(req)
	if err != nil {
		return domain.E(op, domain.ErrInternal, err)
	}

	if center.WhatsappBusinessPhoneID == nil {
		return domain.E(op, domain.ErrInvalidInput, fmt.Errorf("center %s has no whatsapp business phone ID configured", center.ID))
	}

	endpoint := fmt.Sprintf(
		"https://graph.facebook.com/%s/%s/messages",
		MetaGraphApiVersion,
		*center.WhatsappBusinessPhoneID,
	)

	httpReq, err := http.NewRequest("POST", endpoint, bytes.NewReader(payload))
	if err != nil {
		return domain.E(op, domain.ErrInternal, err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.authToken)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return domain.E(op, domain.ErrTransient, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return c.mapWhatsAppError(op, resp.StatusCode, bodyBytes)
	}

	return nil
}

func (c *WhatsappClient) buildConfirmationMessagePayload(
	center *domain.Center,
	appointment *domain.Appointment,
) ([]byte, error) {
	req := WhatsappMessageRequest{
		MessagingProduct: WhatsappMessagingProduct,
		To:               appointment.CustomerPhoneNumber,
		Type:             WhatsappMessageRequestTemplateType,
		Template: &WhatsappTemplatePayload{
			Name: WhatsappAppointmentConfirmqtionTemplate,
			Language: WhatsappTemplatePayloadLanguage{
				Code: WhatsappTemplateLanguageCodeEnglish,
			},
			Components: []WhatsappTemplatePayloadComponent{
				{
					Type: "body",
					Parameters: []WhatsappTemplatePayloadComponentParameter{
						{Type: "text", ParameterName: "center", Text: center.Name},
						{Type: "text", ParameterName: "day", Text: appointment.Date},
						{Type: "text", ParameterName: "start_time", Text: appointment.StartTime},
						{Type: "text", ParameterName: "end_time", Text: appointment.EndTime},
						{Type: "text", ParameterName: "address", Text: center.Address},
					},
				},
			},
		},
	}

	return json.Marshal(req)
}

func (c *WhatsappClient) mapWhatsAppError(
	op string,
	status int,
	body []byte,
) error {
	var waErrResp WhatsappMessageErrorResponse
	if err := json.Unmarshal(body, &waErrResp); err != nil {
		return domain.E(op, domain.ErrInternal, fmt.Errorf("whatsapp error (unparsed): %s", string(body)))
	}

	cause := fmt.Errorf(
		"whatsapp error: status=%d code=%d subcode=%d type=%s message=%s trace=%s",
		status,
		waErrResp.Error.Code,
		waErrResp.Error.ErrorSubcode,
		waErrResp.Error.Type,
		waErrResp.Error.Message,
		waErrResp.Error.FBTraceID,
	)

	switch {
	case status == 400:
		return domain.E(op, domain.ErrInvalidInput, cause)

	case status == 401:
		return domain.E(op, domain.ErrUnauthorized, cause)

	case status == 403:
		return domain.E(op, domain.ErrUnauthorized, cause)

	case status == 404:
		return domain.E(op, domain.ErrInvalidInput, cause)

	case status == 429:
		return domain.E(op, domain.ErrRateLimited, cause)

	case status >= 500:
		return domain.E(op, domain.ErrTransient, cause)

	default:
		return domain.E(op, domain.ErrInternal, cause)
	}
}
