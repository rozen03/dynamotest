package usage_example

import (
	"os"
	"testing"

	"github.com/rozen03/dynamotest"
)

func TestMain(m *testing.M) {
	code := dynamotest.RunTestAndCleanup(m)
	os.Exit(code)
}
