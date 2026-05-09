package dynamo

import (
	"appointment-system/internal/domain"
	"appointment-system/pkg/logger"
	"context"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type DynamoDBCentersRepo struct {
	dynamoDbClient *dynamodb.Client
	table          string
	logger         logger.Logger
}

func NewDynamoDBCentersRepo(
	client *dynamodb.Client,
	centerTableName string,
	log logger.Logger,
) *DynamoDBCentersRepo {
	return &DynamoDBCentersRepo{
		dynamoDbClient: client,
		table:          centerTableName,
		logger:         log,
	}
}

func (r *DynamoDBCentersRepo) GetByID(ctx context.Context, id string) (*domain.Center, error) {
	const op = "DynamoDBCentersRepo.GetById"
	key, err := buildAppointmentTableKey(id)
	if err != nil {
		return nil, domain.E(op, domain.ErrInternal, err)
	}

	response, err := r.dynamoDbClient.GetItem(ctx, &dynamodb.GetItemInput{
		Key:       key,
		TableName: &r.table,
	})
	if err != nil {
		return nil, mapDynamoErr(op, err)
	}

	if len(response.Item) == 0 {
		return nil, domain.E(op, domain.ErrApointmentNotFound, nil)
	}

	var center domain.Center
	err = attributevalue.UnmarshalMap(response.Item, &center)
	if err != nil {
		return nil, domain.E(op, domain.ErrInternal, err)
	}

	return &center, nil
}
