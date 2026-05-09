package flowww

import (
	"appointment-system/internal/domain"
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

type ILambdaInvoker interface {
	Invoke(ctx context.Context, params *lambda.InvokeInput, optFns ...func(*lambda.Options)) (*lambda.InvokeOutput, error)
	InvokeAsync(ctx context.Context, params *lambda.InvokeAsyncInput, optFns ...func(*lambda.Options)) (*lambda.InvokeAsyncOutput, error)
}

type FlowwwClient struct {
	invoker      ILambdaInvoker
	functionName string
}

func NewFlowwwClient(invoker ILambdaInvoker, functionName string) *FlowwwClient {
	return &FlowwwClient{
		invoker:      invoker,
		functionName: functionName,
	}
}

// TODO: IMPROVE ABSTRACTING INVOKE LOGIC INTO A FUNCTION

func (c *FlowwwClient) ListAppointments(
	ctx context.Context,
	centerId string,
) (*[]FlowwwAppointmentDTO, error) {
	const op = "FlowwwClient.ListAppointments"
	req := ListAppointmentsRequest{
		Operation: "get-appointments",
		CenterId:  centerId,
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return &[]FlowwwAppointmentDTO{}, domain.E(op, domain.ErrInternal, err)
	}

	out, err := c.invoker.Invoke(
		ctx,
		&lambda.InvokeInput{
			FunctionName: &c.functionName,
			Payload:      payload,
		},
	)

	// Error invoking Lambda, AWS error
	if err != nil {
		return &[]FlowwwAppointmentDTO{}, domain.E(op, domain.ErrTransient, err)
	}

	// Error while executing the function
	if out.FunctionError != nil {
		return &[]FlowwwAppointmentDTO{}, domain.E(op, domain.ErrInternal, fmt.Errorf("Flowww Operations Error: %s", *out.FunctionError))
	}

	envelope, err := decodeEnvelope(out.Payload)
	if err != nil {
		return &[]FlowwwAppointmentDTO{}, domain.E(op, domain.ErrInternal, err)
	}

	if !envelope.Ok && envelope.Error != nil {
		return &[]FlowwwAppointmentDTO{}, c.mapFlowwwError(op, *envelope.Error)
	}

	var appointments []FlowwwAppointmentDTO
	if err = json.Unmarshal(envelope.Data, &appointments); err != nil {
		return &[]FlowwwAppointmentDTO{}, domain.E(op, domain.ErrInternal, err)
	}

	return &appointments, nil
}

func (c *FlowwwClient) UpdateConfirmation(
	ctx context.Context,
	confirmationStatus bool,
	centerId, appointmentId, appointmentDate string,
) error {
	const op = "FlowwwClient.UpdateConfirmation"
	req := UpdateConfirmationRequest{
		Operation:          "confirmation-response",
		CenterId:           centerId,
		AppointmentId:      appointmentId,
		AppointmentDate:    appointmentDate,
		ConfirmationStatus: confirmationStatus,
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return err
	}

	out, err := c.invoker.Invoke(
		ctx,
		&lambda.InvokeInput{
			FunctionName: &c.functionName,
			Payload:      payload,
		},
	)

	// Error invoking Lambda, AWS error
	if err != nil {
		return domain.E(op, domain.ErrTransient, err)
	}

	// Error while executing the function
	if out.FunctionError != nil {
		return domain.E(op, domain.ErrInternal, fmt.Errorf("Flowww Operations Error: %s", *out.FunctionError))
	}

	envelope, err := decodeEnvelope(out.Payload)
	if err != nil {
		return domain.E(op, domain.ErrInternal, err)
	}

	if !envelope.Ok && envelope.Error != nil {
		return c.mapFlowwwError(op, *envelope.Error)
	}

	var data UpdateConfirmationData
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		return domain.E(op, domain.ErrInternal, err)
	}

	return nil
}

func (c *FlowwwClient) mapFlowwwError(op string, flowwwErr FlowwwErrorDTO) error {
	cause := fmt.Errorf("%s: %s", flowwwErr.Code, flowwwErr.Message)

	switch flowwwErr.Kind {
	case "TRANSIENT":
		return domain.E(op, domain.ErrTransient, cause)

	case "RATE_LIMITED":
		return domain.E(op, domain.ErrRateLimited, cause)

	case "UNAUTHORIZED":
		return domain.E(op, domain.ErrUnauthorized, cause)

	case "INVALID_INPUT":
		return domain.E(op, domain.ErrInvalidInput, cause)

	default:
		return domain.E(op, domain.ErrInternal, cause)
	}
}
