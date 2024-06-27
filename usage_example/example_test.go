package usage_example

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rozen03/dynamotest"
)

func TestRepositoryExample_Create(t *testing.T) {
	client := dynamotest.DynamoDBClient()
	table := client.CreateTestingTable(t, "test", getSchema())

	repo := RepositoryExample{client: client, table: &table}
	model := ExampleModel{
		ID:    "1",
		SK:    "1",
		Value: "example",
	}

	err := repo.Create(context.Background(), model)
	assert.NoError(t, err)
}

func TestRepositoryExample_Read(t *testing.T) {
	client := dynamotest.DynamoDBClient()
	table := client.CreateTestingTable(t, "test", getSchema(), ExampleModel{
		ID:    "1",
		SK:    "1",
		Value: "example",
	})

	repo := RepositoryExample{client: client, table: &table}

	// Read the item
	result, err := repo.Read(context.Background(), "1")
	assert.NoError(t, err)
	assert.Equal(t, "1", result.ID)
	assert.Equal(t, "example", result.Value)
}

func TestRepositoryExample_Update(t *testing.T) {
	client := dynamotest.DynamoDBClient()
	model := ExampleModel{
		ID:    "1",
		SK:    "1",
		Value: "example",
	}
	table := client.CreateTestingTable(t, "test", getSchema(), model)

	repo := RepositoryExample{client: client, table: &table}

	// Update the item
	model.Value = "updated-value"
	err := repo.Update(context.Background(), model)
	assert.NoError(t, err)

	// Read the item to verify update
	updatedModel, err := repo.Read(context.Background(), "1")
	assert.NoError(t, err)
	assert.Equal(t, "updated-value", updatedModel.Value)
}
