package wefact

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testApiKey = os.Getenv("WEFACT_API_KEY")

func TestApi(t *testing.T) {
	client := New(&Config{Key: testApiKey})
	//var results map[string]interface{}

	_, err := client.Request("invoice", "list", nil)
	assert.Nil(t, err)
}
