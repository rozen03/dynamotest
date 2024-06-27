Here's the updated `README.md` with a more concise example and separated steps for better clarity:

# dynamotest

Leverage the power of [DynamoDB Local][1] with [ory/dockertest][2] to create your DynamoDB test cases effortlessly.

[1]: https://hub.docker.com/r/amazon/dynamodb-local/
[2]: https://github.com/ory/dockertest

## 🌄 What is `dynamotest`?

`dynamotest` is a package designed to help set up a DynamoDB Local Docker instance on your machine as part of Go test code. It uses [`ory/dockertest`][2] to start the DynamoDB Local instance in your Go test code, and is configured so that each call to `dynamotest.NewDynamoDB(t)` will create a dedicated instance, allowing parallel testing on multiple Docker instances. The function returns a new DynamoDB client which is already connected to the instance, enabling you to start using the client immediately. Additionally, it provides a clean-up function to ensure that the Docker instance gets deleted if clean-up is preferred. If you do not call the clean-up function, the instance will keep running, which may be useful for debugging and investigation.

`dynamotest` also offers a helper function, `dynamotest.PrepTable(t, client, ...dynamotest.InitialTableSetup)`, to prepare tables and datasets for setting up the table beforehand.

It is also worth noting that this package uses only the v2 version of the AWS SDK.

**NOTE**: It is a prerequisite that you are able to start up a Docker container for DynamoDB Local.

## 🚀 Example

Here's a simple example to get you started with `dynamotest`.

### Step 1: Define Table Schema

```go
package usage_example

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func getSchema() dynamodb.CreateTableInput {
	return dynamodb.CreateTableInput{
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("pk"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("sk"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("pk"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: aws.String("sk"),
				KeyType:       types.KeyTypeRange,
			},
		},
		// Not necessary yet, but left here for future reference
		// GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{},
	}
}
```

### Step 2: Write Test Code

```go
package usage_example

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/rozen03/dynamotest"
)

func TestMain(m *testing.M) {
	code := dynamotest.RunTestAndCleanup(m)
	os.Exit(code)
}

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
```

Refer to [usage_example/example_test.go](/usage_example/example_test.go) for the complete code and more detailed examples.