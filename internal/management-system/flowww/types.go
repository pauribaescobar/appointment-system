package flowww

import "encoding/json"

type ListAppointmentsRequest struct {
	Operation string `json:"operation"`
	CenterId  string `json:"centerId"`
}

type UpdateConfirmationRequest struct {
	Operation          string `json:"operation"`
	CenterId           string `json:"centerId"`
	ConfirmationStatus bool   `json:"confirmationStatus"`
	AppointmentDate    string `json:"appointmentDate"`
	AppointmentId      string `json:"appointmentId"`
}

type FlowwwAppointmentDTO struct {
	ID                  string `json:"id"`
	CustomerId          string `json:"customerId"`
	CustomerName        string `json:"customerName"`
	CustomerPhoneNumber string `json:"customerPhoneNumber"`
	Date                string `json:"date"`
	StartTime           string `json:"startTime"`
	EndTime             string `json:"endTime"`
}

type UpdateConfirmationData struct {
	Skipped bool `json:"skipped"`
}

type LambdaHTTPResponse struct {
	Body       string `json:"body"`
	StatusCode int64  `json:"statusCode"`
}

type FlowwwEnvelope struct {
	Ok    bool            `json:"ok"`
	Data  json.RawMessage `json:"data,omitempty"`
	Error *FlowwwErrorDTO `json:"error,omitempty"`
}

type FlowwwErrorDTO struct {
	Kind    string `json:"kind"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func decodeEnvelope(payload []byte) (*FlowwwEnvelope, error) {
	var httpResp LambdaHTTPResponse
	if err := json.Unmarshal(payload, &httpResp); err != nil {
		return nil, err
	}

	var envelope FlowwwEnvelope
	if err := json.Unmarshal([]byte(httpResp.Body), &envelope); err != nil {
		return nil, err
	}

	return &envelope, nil
}
