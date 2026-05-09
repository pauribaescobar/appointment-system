package dynamo

import (
	"appointment-system/internal/domain"
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/smithy-go"
)

func buildAppointmentTableKey(
	id string,
) (map[string]types.AttributeValue, error) {
	const op = "DynamoUtils.buildAppointmentTableKey"
	marshalledId, err := attributevalue.Marshal(id)
	if err != nil {
		return nil, domain.E(op, domain.ErrInternal, err)
	}

	return map[string]types.AttributeValue{
		"id": marshalledId,
	}, nil
}

func mapDynamoErr(
	op string,
	err error,
) error {
	if dynamoErrIsErrTransient(err) {
		return domain.E(op, domain.ErrTransient, err)
	}

	if dynamoErrIsErrRateLimited(err) {
		return domain.E(op, domain.ErrRateLimited, err)
	}

	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "AccessDeniedException",
			"UnrecognizedClientException",
			"InvalidSignatureException":
			return domain.E(op, domain.ErrUnauthorized, err)
		}
	}

	return domain.E(op, domain.ErrInternal, err)
}

func dynamoErrIsErrTransient(
	err error,
) bool {
	var (
		ise *types.InternalServerError
	)
	return errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, context.Canceled) ||
		errors.As(err, &ise)
}

func dynamoErrIsErrRateLimited(
	err error,
) bool {
	var (
		pte *types.ProvisionedThroughputExceededException
		tle *types.ThrottlingException
		rle *types.RequestLimitExceeded
	)
	return errors.As(err, &pte) ||
		errors.As(err, &tle) ||
		errors.As(err, &rle)
}
