package dynamotest

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

// NewDynamoDB creates a Docker container with DynamoDB Local, and returns the
// connected DynamoDB client. Clean up function is returned as well to ensure
// container gets removed after test is complete.
func NewDynamoDB() (Client, func()) {

	pool, err := dockertest.NewPool("")
	if err != nil {
		panic("Could not connect to docker" + err.Error())
	}

	runOpt := &dockertest.RunOptions{
		Repository: dynamoDBLocalRepo,
		Tag:        dynamoDBLocalTag,

		PortBindings: map[docker.Port][]docker.PortBinding{
			"0/tcp": {{HostIP: "localhost", HostPort: "8000/tcp"}},
		},
	}
	resource, err := pool.RunWithOptions(runOpt)
	if err != nil {
		panic("Could not start DynamoDB Local " + err.Error())
	}
	port := resource.GetHostPort("8000/tcp")
	fmt.Println("Using host:port of", port)

	dynamoClient := createDB(pool, port)
	client := Client{Client: dynamoClient, ContainerID: resource.Container.ID}

	return client, func() {
		if err = pool.Purge(resource); err != nil {
			panic("Could not purge DynamoDB " + err.Error())
		}
	}
}

func createDB(pool *dockertest.Pool, port string) *dynamodb.Client {
	ctx := context.Background()
	var dynamoClient *dynamodb.Client
	err := pool.Retry(func() error {
		cfg, err := config.LoadDefaultConfig(ctx,
			config.WithRegion("us-east-1"),
			config.WithCredentialsProvider(
				credentials.StaticCredentialsProvider{
					Value: aws.Credentials{
						AccessKeyID: "dummy", SecretAccessKey: "dummy", SessionToken: "dummy",
						Source: "Hard-coded credentials; values are irrelevant for local DynamoDB",
					},
				}),
		)
		if err != nil {
			return err
		}

		dynamoClient = dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
			o.BaseEndpoint = aws.String("http://" + port)
		})
		return nil
	})
	if err != nil {
		panic("Could not connect to the Docker instance of DynamoDB Local" + err.Error())
	}
	return dynamoClient
}

type InitialTableSetup struct {
	Table       *dynamodb.CreateTableInput
	InitialData []*types.PutRequest
}

type dynamoClient interface {
	CreateTable(ctx context.Context, params *dynamodb.CreateTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.CreateTableOutput, error)
	BatchWriteItem(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error)
}

// PrepTable iterates the provided input, creates tables and put items.
func PrepTable(t testing.TB, client dynamoClient, input ...InitialTableSetup) {
	t.Helper()

	// Add extra retry setup in case Docker instance is busy. This can happen
	// especially within a CI environment, and the default retry count of 3
	// times is too fragile.
	opt := func(o *dynamodb.Options) { o.RetryMaxAttempts = 10 }
	for _, i := range input {
		_, err := client.CreateTable(context.Background(), i.Table, opt)
		if err != nil {
			t.Fatalf("Could not create table '%s': %v", *i.Table.TableName, err)
		}

		if len(i.InitialData) == 0 {
			t.Logf("Table '%s' has been created, and no initial data has been added", *i.Table.TableName)
			continue
		}

		puts := []types.WriteRequest{}
		for _, d := range i.InitialData {
			puts = append(puts, types.WriteRequest{PutRequest: d})
		}
		_, err = client.BatchWriteItem(context.Background(), &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				*i.Table.TableName: puts,
			},
		})
		if err != nil {
			t.Fatalf("Could not write data to table '%s': %v", *i.Table.TableName, err)
		}

		t.Logf("Table '%s' has been created", *i.Table.TableName)
	}
}
