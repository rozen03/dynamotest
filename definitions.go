package dynamotest

var (
	// DynamoDBLocalRepo is the repository for the DynamoDB Local image
	// this is at least in 2004/06/23 https://hub.docker.com/r/amazon/dynamodb-local
	dynamoDBLocalRepo = "amazon/dynamodb-local"

	// DynamoDBLocalTag is the tag for the DynamoDB Local image
	// this is set to latest just due to the fact that dynamoDB is a managed service
	// and we want tests that are as close to the real thing as possible
	dynamoDBLocalTag = "latest"
)
