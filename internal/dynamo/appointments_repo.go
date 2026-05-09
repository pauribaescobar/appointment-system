package dynamo

import (
	"appointment-system/internal/domain"
	"appointment-system/pkg/logger"
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/smithy-go"
)

type DynamoDBAppointmentsRepo struct {
	dynamoDbClient *dynamodb.Client
	table          string
	logger         logger.Logger
}

func NewDynamoDBAppointmentsRepo(
	client *dynamodb.Client,
	appointmentsTableName string,
	log logger.Logger,
) *DynamoDBAppointmentsRepo {
	return &DynamoDBAppointmentsRepo{
		dynamoDbClient: client,
		table:          appointmentsTableName,
		logger:         log,
	}
}

var (
	APPOINTMENT_TABLE_STATUS_INDEX                     = "status-index"
	APPOINTMENT_TABLE_MESSAGE_ID_INDEX                 = "message_id-index"
	APPOINTMENT_NOT_EXISTS_CONDITION_EXPRESSION string = "attribute_not_exists(id)"
)

func (r *DynamoDBAppointmentsRepo) PutIfNotExists(ctx context.Context, appointment *domain.Appointment) (*domain.Appointment, error) {
	const op = "DynamoDBAppointmentsRepo.PutIfNotExists"
	item, err := attributevalue.MarshalMap(appointment)
	if err != nil {
		return nil, domain.E(op, domain.ErrInternal, err)
	}

	_, err = r.dynamoDbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           &r.table,
		Item:                item,
		ConditionExpression: &APPOINTMENT_NOT_EXISTS_CONDITION_EXPRESSION,
	})
	if err != nil {
		var condErr *types.ConditionalCheckFailedException
		if errors.As(err, &condErr) {
			existingAp, getErr := r.GetByID(ctx, appointment.ID)
			if getErr != nil {
				if errors.Is(getErr, domain.ErrApointmentNotFound) {
					return nil, domain.E(op, domain.ErrTransient, err)
				}
				return nil, domain.E(op, domain.ErrTransient, getErr)
			}
			return existingAp, nil
		}
		return nil, r.mapDynamoErr(op, err)
	}

	return nil, nil
}

func (r *DynamoDBAppointmentsRepo) MarkSent(
	ctx context.Context,
	id string,
	messageId string,
	messageChannel string,
) error {
	const op = "DynamoDBAppointmentsRepo.MarkSent"
	key, err := r.buildAppointmentTableKey(id)
	if err != nil {
		return domain.E(op, domain.ErrInternal, err)
	}

	cond := expression.And(
		expression.Name("status").Equal(expression.Value(string(domain.AppointmentStatusPending))),
		expression.AttributeExists(expression.Name("id")),
	)

	update := expression.
		Set(expression.Name("status"), expression.Value(string(domain.AppoinmentStatusSent))).
		Set(expression.Name("message_id"), expression.Value(messageId)).
		Set(expression.Name("message_channel"), expression.Value(messageChannel)).
		Set(expression.Name("message_sent_at"), expression.Value(time.Now().Unix()))

	expr, err := expression.NewBuilder().
		WithCondition(cond).
		WithUpdate(update).
		Build()
	if err != nil {
		return domain.E(op, domain.ErrInternal, err)
	}

	_, err = r.dynamoDbClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		Key:                       key,
		TableName:                 &r.table,
		ConditionExpression:       expr.Condition(),
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	if err == nil {
		return nil
	}

	var condErr *types.ConditionalCheckFailedException
	if errors.As(err, &condErr) {
		return domain.E(op, domain.ErrAppointmentAlreadyProcessed, err)
	}

	return r.mapDynamoErr(op, err)
}

func (r *DynamoDBAppointmentsRepo) GetByMessageID(ctx context.Context, messageID string) (*domain.Appointment, error) {
	const op = "DynamoDBAppointmentsRepo.GetByMessageID"

	keyEx := expression.Key("message_id").Equal(expression.Value(messageID))
	expr, err := expression.NewBuilder().WithKeyCondition(keyEx).Build()
	if err != nil {
		return nil, domain.E(op, domain.ErrInternal, err)
	}

	response, err := r.dynamoDbClient.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(r.table),
		IndexName:                 aws.String(APPOINTMENT_TABLE_MESSAGE_ID_INDEX),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		Limit:                     aws.Int32(1),
	})
	if err != nil {
		return nil, r.mapDynamoErr(op, err)
	}

	if len(response.Items) == 0 {
		return nil, domain.E(op, domain.ErrApointmentNotFound, nil)
	}

	var appointment domain.Appointment
	if err := attributevalue.UnmarshalMap(response.Items[0], &appointment); err != nil {
		return nil, domain.E(op, domain.ErrInternal, err)
	}

	return &appointment, nil
}

func (r *DynamoDBAppointmentsRepo) GetByID(ctx context.Context, id string) (*domain.Appointment, error) {
	const op = "DynamoDBAppointmentsRepo.GetById"
	key, err := r.buildAppointmentTableKey(id)
	if err != nil {
		return nil, domain.E(op, domain.ErrInternal, err)
	}

	response, err := r.dynamoDbClient.GetItem(ctx, &dynamodb.GetItemInput{
		Key:       key,
		TableName: &r.table,
	})
	if err != nil {
		return nil, r.mapDynamoErr(op, err)
	}

	if len(response.Item) == 0 {
		return nil, domain.E(op, domain.ErrApointmentNotFound, nil)
	}

	var appointment domain.Appointment
	err = attributevalue.UnmarshalMap(response.Item, &appointment)
	if err != nil {
		return nil, domain.E(op, domain.ErrInternal, err)
	}

	return &appointment, nil
}

func (r *DynamoDBAppointmentsRepo) Delete(ctx context.Context, id string) error {
	const op = "DynamoDBAppointmentsRepo.Delete"
	key, err := r.buildAppointmentTableKey(id)
	if err != nil {
		return domain.E(op, domain.ErrInternal, err)
	}

	_, err = r.dynamoDbClient.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		Key:       key,
		TableName: &r.table,
	})

	if err == nil {
		return nil
	}

	return r.mapDynamoErr(op, err)
}

func (r *DynamoDBAppointmentsRepo) QueryPendingAppointments(
	ctx context.Context,
) (*[]domain.Appointment, error) {
	const op = "DynamoDBAppointmentsRepo.QueryPendingAppointments"

	var appointments []domain.Appointment

	keyEx := expression.Key("status").
		Equal(expression.Value("PENDING"))
	expr, err := expression.NewBuilder().WithKeyCondition(keyEx).Build()
	if err != nil {
		return nil, domain.E(op, domain.ErrInternal, err)
	}

	queryPaginator := dynamodb.NewQueryPaginator(r.dynamoDbClient, &dynamodb.QueryInput{
		TableName:                 aws.String(r.table),
		IndexName:                 aws.String(APPOINTMENT_TABLE_STATUS_INDEX),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ScanIndexForward:          aws.Bool(true),
	})

	for queryPaginator.HasMorePages() {
		response, err := queryPaginator.NextPage(ctx)
		if err != nil {
			return nil, r.mapDynamoErr(op, err)
		}

		var appointmentsPage []domain.Appointment
		err = attributevalue.UnmarshalListOfMaps(response.Items, &appointmentsPage)
		if err != nil {
			return nil, domain.E(op, domain.ErrInternal, err)
		}

		appointments = append(appointments, appointmentsPage...)
	}

	return &appointments, nil
}

func (r *DynamoDBAppointmentsRepo) buildAppointmentTableKey(
	id string,
) (map[string]types.AttributeValue, error) {
	const op = "DynamoDBAppointmentsRepo.buildAppointmentTableKey"
	marshalledId, err := attributevalue.Marshal(id)
	if err != nil {
		return nil, domain.E(op, domain.ErrInternal, err)
	}

	return map[string]types.AttributeValue{
		"id": marshalledId,
	}, nil
}

func (r *DynamoDBAppointmentsRepo) mapDynamoErr(
	op string,
	err error,
) error {
	if r.dynamoErrIsErrTransient(err) {
		return domain.E(op, domain.ErrTransient, err)
	}

	if r.dynamoErrIsErrRateLimited(err) {
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

func (r *DynamoDBAppointmentsRepo) dynamoErrIsErrTransient(
	err error,
) bool {
	var (
		ise *types.InternalServerError
	)
	return errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, context.Canceled) ||
		errors.As(err, &ise)
}

func (r *DynamoDBAppointmentsRepo) dynamoErrIsErrRateLimited(
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
