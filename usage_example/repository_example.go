package usage_example

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DBClient interface {
	Query(
		ctx context.Context,
		params *dynamodb.QueryInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.QueryOutput, error)
	PutItem(
		context.Context,
		*dynamodb.PutItemInput,
		...func(*dynamodb.Options),
	) (*dynamodb.PutItemOutput, error)
	UpdateItem(
		context.Context,
		*dynamodb.UpdateItemInput,
		...func(*dynamodb.Options),
	) (*dynamodb.UpdateItemOutput, error)
}
type ExampleModel struct {
	ID    string `dynamodbav:"pk"`
	SK    string `dynamodbav:"sk"`
	Value string `dynamodbav:"value"`
}
type RepositoryExample struct {
	client DBClient
	table  *string
}

func (r RepositoryExample) Create(ctx context.Context, m ExampleModel) error {
	statusItem, err := attributevalue.MarshalMap(m)
	if err != nil {
		return fmt.Errorf("error marshaling: %w", err)

	}
	putItem := &dynamodb.PutItemInput{
		TableName: r.table,
		Item:      statusItem,
	}
	_, err = r.client.PutItem(ctx, putItem)
	if err != nil {
		return fmt.Errorf("error inserting: %w", err)
	}
	return nil

}

func (r RepositoryExample) Read(ctx context.Context, id string) (ExampleModel, error) {
	var zeroResponse ExampleModel
	query := &dynamodb.QueryInput{
		TableName: r.table,
		KeyConditions: map[string]types.Condition{
			"pk": {
				ComparisonOperator: types.ComparisonOperatorEq,
				AttributeValueList: []types.AttributeValue{
					&types.AttributeValueMemberS{Value: id},
				},
			},
		},
	}

	result, err := r.client.Query(ctx, query)
	if err != nil {
		return zeroResponse, fmt.Errorf("error querying dynamo: %w", err)
	}

	if len(result.Items) == 0 {
		return zeroResponse, fmt.Errorf("not found")
	}

	var response ExampleModel
	for _, item := range result.Items {
		err := attributevalue.UnmarshalMap(item, &response)
		if err != nil {
			return zeroResponse, fmt.Errorf("error unmarshaling: %w", err)
		}
	}

	return response, nil
}

func (r RepositoryExample) Update(ctx context.Context, m ExampleModel) error {

	key, err := attributevalue.MarshalMap(map[string]string{
		"pk": m.ID,
		"sk": m.SK,
	})
	if err != nil {
		return fmt.Errorf("unexpected error in users.repository.update when marshal key: %w", err)
	}

	var ue expression.UpdateBuilder

	ue = ue.Set(expression.Name("value"), expression.Value(m.Value))

	cond := expression.AttributeExists(expression.Name("pk"))
	expr, err := expression.NewBuilder().WithUpdate(ue).WithCondition(cond).Build()
	if err != nil {
		return fmt.Errorf("failed to build expression for %v: %w", m, err)
	}

	item := &dynamodb.UpdateItemInput{
		Key:                       key,
		TableName:                 r.table,
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ConditionExpression:       expr.Condition(),
	}

	_, err = r.client.UpdateItem(ctx, item)
	if err != nil {
		return fmt.Errorf("failed to update item of %v: %w", m, err)
	}

	return nil
}
