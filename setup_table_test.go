package dynamotest_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/go-cmp/cmp"

	"github.com/rozen03/dynamotest"
)

func compareAttributeValues(expected, actual types.AttributeValue) bool {
	switch e := expected.(type) {
	case *types.AttributeValueMemberS:
		a, ok := actual.(*types.AttributeValueMemberS)
		return ok && e.Value == a.Value
	case *types.AttributeValueMemberN:
		a, ok := actual.(*types.AttributeValueMemberN)
		return ok && e.Value == a.Value
	case *types.AttributeValueMemberBOOL:
		a, ok := actual.(*types.AttributeValueMemberBOOL)
		return ok && e.Value == a.Value
	default:
		return false
	}
}

type testData struct {
	PK string `dynamodbav:"test_PK" json:"test_PK"`

	X string `dynamodbav:"X" json:"X"`
	Y string `dynamodbav:"Y" json:"Y"`
	Z string `dynamodbav:"Z" json:"Z"`
}

func Test_Query(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		schema         dynamodb.CreateTableInput
		initialData    []any
		query          *dynamodb.QueryInput
		expectedOutput []map[string]types.AttributeValue
	}{
		"simple table with single data": {
			schema: dynamodb.CreateTableInput{
				AttributeDefinitions: []types.AttributeDefinition{
					{
						AttributeName: aws.String("id"),
						AttributeType: types.ScalarAttributeTypeS,
					},
				},
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("id"),
						KeyType:       types.KeyTypeHash,
					},
				},
				TableName:   aws.String("my-table"),
				BillingMode: types.BillingModePayPerRequest,
			},
			initialData: []any{
				map[string]interface{}{
					"id":    "123",
					"name":  "John Doe",
					"email": "john@doe.io",
				},
			},
			query: &dynamodb.QueryInput{
				TableName:              aws.String("my-table"),
				KeyConditionExpression: aws.String("id = :hashKey"),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":hashKey": &types.AttributeValueMemberS{Value: "123"},
				},
			},
			expectedOutput: []map[string]types.AttributeValue{
				{
					"id":    &types.AttributeValueMemberS{Value: "123"},
					"name":  &types.AttributeValueMemberS{Value: "John Doe"},
					"email": &types.AttributeValueMemberS{Value: "john@doe.io"},
				},
			},
		},

		"sortable table": {
			schema: dynamodb.CreateTableInput{
				TableName: aws.String("sortable-table"),
				AttributeDefinitions: []types.AttributeDefinition{
					{
						AttributeName: aws.String("s_id"),
						AttributeType: types.ScalarAttributeTypeS,
					},
					{
						AttributeName: aws.String("date"),
						AttributeType: types.ScalarAttributeTypeS,
					},
				},
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("s_id"),
						KeyType:       types.KeyTypeHash,
					},
					{
						AttributeName: aws.String("date"),
						KeyType:       types.KeyTypeRange,
					},
				},
				BillingMode: types.BillingModePayPerRequest,
			},
			initialData: []any{
				map[string]interface{}{
					"s_id": "111",
					"date": "2022-02-15",
				},
				map[string]interface{}{
					"s_id": "111",
					"date": "2022-02-16",
				},
				map[string]interface{}{
					"s_id": "111",
					"date": "2022-02-17",
				},
			},
			query: &dynamodb.QueryInput{
				TableName:              aws.String("sortable-table"),
				ScanIndexForward:       aws.Bool(false), // Descending
				Limit:                  aws.Int32(5),
				KeyConditionExpression: aws.String("s_id = :hashKey and #date > :sortKey"),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":hashKey": &types.AttributeValueMemberS{Value: "111"},
					":sortKey": &types.AttributeValueMemberS{Value: "2022-02-15"}, // Filtering first item
				},
				ExpressionAttributeNames: map[string]string{
					"#date": "date",
				},
			},
			expectedOutput: []map[string]types.AttributeValue{
				{
					"s_id": &types.AttributeValueMemberS{Value: "111"},
					"date": &types.AttributeValueMemberS{Value: "2022-02-17"},
				},
				{
					"s_id": &types.AttributeValueMemberS{Value: "111"},
					"date": &types.AttributeValueMemberS{Value: "2022-02-16"},
				},
			},
		},

		"table with multiple items and complex queries": {
			schema: dynamodb.CreateTableInput{
				TableName: aws.String("complex-table"),
				AttributeDefinitions: []types.AttributeDefinition{
					{
						AttributeName: aws.String("id"),
						AttributeType: types.ScalarAttributeTypeS,
					},
					{
						AttributeName: aws.String("category"),
						AttributeType: types.ScalarAttributeTypeS,
					},
				},
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("id"),
						KeyType:       types.KeyTypeHash,
					},
					{
						AttributeName: aws.String("category"),
						KeyType:       types.KeyTypeRange,
					},
				},
				BillingMode: types.BillingModePayPerRequest,
			},
			initialData: []any{
				map[string]interface{}{
					"id":       "1",
					"category": "A",
					"name":     "Item 1",
				},
				map[string]interface{}{
					"id":       "2",
					"category": "B",
					"name":     "Item 2",
				},
				map[string]interface{}{
					"id":       "3",
					"category": "A",
					"name":     "Item 3",
				},
			},
			query: &dynamodb.QueryInput{
				TableName:              aws.String("complex-table"),
				KeyConditionExpression: aws.String("id = :hashKey and category = :rangeKey"),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":hashKey":  &types.AttributeValueMemberS{Value: "1"},
					":rangeKey": &types.AttributeValueMemberS{Value: "A"},
				},
			},
			expectedOutput: []map[string]types.AttributeValue{
				{
					"id":       &types.AttributeValueMemberS{Value: "1"},
					"category": &types.AttributeValueMemberS{Value: "A"},
					"name":     &types.AttributeValueMemberS{Value: "Item 1"},
				},
			},
		},

		"table with no initial data": {
			schema: dynamodb.CreateTableInput{
				TableName: aws.String("empty-table"),
				AttributeDefinitions: []types.AttributeDefinition{
					{
						AttributeName: aws.String("id"),
						AttributeType: types.ScalarAttributeTypeS,
					},
				},
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("id"),
						KeyType:       types.KeyTypeHash,
					},
				},
				BillingMode: types.BillingModePayPerRequest,
			},
			query: &dynamodb.QueryInput{
				TableName:              aws.String("empty-table"),
				KeyConditionExpression: aws.String("id = :hashKey"),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":hashKey": &types.AttributeValueMemberS{Value: "123"},
				},
			},
			expectedOutput: []map[string]types.AttributeValue{},
		},

		"table with secondary indexes": {
			schema: dynamodb.CreateTableInput{
				TableName: aws.String("index-table"),
				AttributeDefinitions: []types.AttributeDefinition{
					{
						AttributeName: aws.String("id"),
						AttributeType: types.ScalarAttributeTypeS,
					},
					{
						AttributeName: aws.String("name"),
						AttributeType: types.ScalarAttributeTypeS,
					},
				},
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("id"),
						KeyType:       types.KeyTypeHash,
					},
				},
				BillingMode: types.BillingModePayPerRequest,
				GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
					{
						IndexName: aws.String("name-index"),
						KeySchema: []types.KeySchemaElement{
							{
								AttributeName: aws.String("name"),
								KeyType:       types.KeyTypeHash,
							},
						},
						Projection: &types.Projection{
							ProjectionType: types.ProjectionTypeAll,
						},
					},
				},
			},
			initialData: []any{
				map[string]interface{}{
					"id":   "1",
					"name": "Alice",
				},
				map[string]interface{}{
					"id":   "2",
					"name": "Bob",
				},
			},
			query: &dynamodb.QueryInput{
				TableName:              aws.String("index-table"),
				IndexName:              aws.String("name-index"),
				KeyConditionExpression: aws.String("#nm = :name"),
				ExpressionAttributeNames: map[string]string{
					"#nm": "name",
				},
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":name": &types.AttributeValueMemberS{Value: "Alice"},
				},
			},
			expectedOutput: []map[string]types.AttributeValue{
				{
					"id":   &types.AttributeValueMemberS{Value: "1"},
					"name": &types.AttributeValueMemberS{Value: "Alice"},
				},
			},
		},

		"query with nonexistent key": {
			schema: dynamodb.CreateTableInput{
				TableName: aws.String("nonexistent-key-table"),
				AttributeDefinitions: []types.AttributeDefinition{
					{
						AttributeName: aws.String("id"),
						AttributeType: types.ScalarAttributeTypeS,
					},
				},
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("id"),
						KeyType:       types.KeyTypeHash,
					},
				},
				BillingMode: types.BillingModePayPerRequest,
			},
			initialData: []any{
				map[string]interface{}{
					"id": "1",
				},
			},
			query: &dynamodb.QueryInput{
				TableName:              aws.String("nonexistent-key-table"),
				KeyConditionExpression: aws.String("id = :hashKey"),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":hashKey": &types.AttributeValueMemberS{Value: "999"},
				},
			},
			expectedOutput: []map[string]types.AttributeValue{},
		},

		"table with different data types": {
			schema: dynamodb.CreateTableInput{
				TableName: aws.String("mixed-types-table"),
				AttributeDefinitions: []types.AttributeDefinition{
					{
						AttributeName: aws.String("id"),
						AttributeType: types.ScalarAttributeTypeS,
					},
				},
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("id"),
						KeyType:       types.KeyTypeHash,
					},
				},
				BillingMode: types.BillingModePayPerRequest,
			},
			initialData: []any{
				map[string]interface{}{
					"id":      "1",
					"number":  42,
					"boolean": true,
				},
			},
			query: &dynamodb.QueryInput{
				TableName:              aws.String("mixed-types-table"),
				KeyConditionExpression: aws.String("id = :hashKey"),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":hashKey": &types.AttributeValueMemberS{Value: "1"},
				},
			},
			expectedOutput: []map[string]types.AttributeValue{
				{
					"id":      &types.AttributeValueMemberS{Value: "1"},
					"number":  &types.AttributeValueMemberN{Value: "42"},
					"boolean": &types.AttributeValueMemberBOOL{Value: true},
				},
			},
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client, clean := dynamotest.NewDynamoDB()
			defer clean()

			tableName := client.CreateTestingTable(t, "test", tc.schema, tc.initialData...)
			tc.query.TableName = aws.String(tableName)

			out, err := client.Query(context.Background(), tc.query)
			if err != nil {
				t.Errorf("failed to query: %v", err)
				return
			}

			if len(out.Items) != len(tc.expectedOutput) {
				t.Errorf("expected %d items, got %d", len(tc.expectedOutput), len(out.Items))
			}

			for i, expectedItem := range tc.expectedOutput {
				if i >= len(out.Items) {
					t.Errorf("missing item at index %d", i)
					continue
				}
				for key, expectedValue := range expectedItem {
					actualValue, ok := out.Items[i][key]
					if !ok {
						t.Errorf("missing key %s in result", key)
						continue
					}
					if !compareAttributeValues(expectedValue, actualValue) {
						t.Errorf("value mismatch for key %s: expected %v, got %v", key, expectedValue, actualValue)
					}
				}
			}
		})
	}
}

func TestQueryWithUnmarshal(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		schema      dynamodb.CreateTableInput
		initialData []any
		query       *dynamodb.QueryInput
		want        interface{}
	}{
		"simple table with query to unmarshal": {
			schema: dynamodb.CreateTableInput{
				AttributeDefinitions: []types.AttributeDefinition{
					{
						AttributeName: aws.String("test_PK"),
						AttributeType: types.ScalarAttributeTypeS,
					},
				},
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("test_PK"),
						KeyType:       types.KeyTypeHash,
					},
				},
				TableName:   aws.String("test-table"),
				BillingMode: types.BillingModePayPerRequest,
			},
			initialData: []any{
				map[string]interface{}{
					"test_PK": "XYZ",
					"X":       "Data for X",
					"Y":       "Data for Y",
					"Z":       "Data for Z",
				},
			},
			query: &dynamodb.QueryInput{
				TableName:              aws.String("test-table"),
				KeyConditionExpression: aws.String("test_PK = :hashKey"),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":hashKey": &types.AttributeValueMemberS{Value: "XYZ"},
				},
			},
			want: &testData{
				PK: "XYZ",
				X:  "Data for X",
				Y:  "Data for Y",
				Z:  "Data for Z",
			},
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			// dynamotest can be safely used in parallel testing.
			t.Parallel()

			client, clean := dynamotest.NewDynamoDB()
			defer clean()

			// Data prep, use simple context.
			tableName := client.CreateTestingTable(t, "test", tc.schema, tc.initialData...)
			tc.query.TableName = aws.String(tableName)

			out, err := client.Query(context.Background(), tc.query)
			if err != nil {
				t.Errorf("failed to query, %v", err)
				return
			}

			if len(out.Items) == 0 {
				t.Fatalf("expected to find items, got none")
			}

			var got testData
			err = attributevalue.UnmarshalMap(out.Items[0], &got)
			if err != nil {
				t.Fatalf("failed to unmarshal result, %v", err)
			}

			if diff := cmp.Diff(tc.want, &got); diff != "" {
				t.Errorf("received data didn't match (-want / +got)\n%s", diff)
			}
		})
	}
}
