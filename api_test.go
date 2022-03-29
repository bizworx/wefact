package wefact

import (
	"os"
	"testing"

	"github.com/kr/pretty"
	"github.com/stretchr/testify/assert"
)

func TestApi(t *testing.T) {
	client := New(os.Getenv("WEFACT_API_KEY"), os.Getenv("WEFACT_PROXY_HOST"))
	list, err := client.Request("invoice", "list", nil)
	assert.Nil(t, err)
	pretty.Log(list)
}
