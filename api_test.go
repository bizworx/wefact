package wefact

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApi(t *testing.T) {
	client := New()
	//var results map[string]interface{}

	_, err := client.Request("invoice", "list", nil)
	assert.Nil(t, err)
}
