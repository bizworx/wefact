package wefact

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApi(t *testing.T) {
	client := New(os.Getenv("WEFACT_API_KEY"))
	//var results map[string]interface{}

	_, err := client.Request("invoice", "list", nil)
	assert.Nil(t, err)
}
