package dynamotest

import "github.com/aws/aws-sdk-go-v2/service/dynamodb"

type Client struct {
	*dynamodb.Client
	ContainerID string
}
