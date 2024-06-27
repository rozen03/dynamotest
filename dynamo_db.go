package dynamotest

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var (
	client Client
	purge  func()
)

func RunTestAndCleanup(m *testing.M) int {
	client, purge = NewDynamoDB()
	defer purge()
	code := m.Run()
	return code
}
func DynamoDBClient() Client {
	if client.Client == nil {

		panic("DynamoDB client is not initialized")
	}
	return client
}

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
	client = Client{Client: dynamoClient, ContainerID: resource.Container.ID}
	purge = func() {

		if err := pool.Purge(resource); err != nil {
			panic("Could not purge DynamoDB " + err.Error())
		}

	}

	return client, purge
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
