package warmer_test

import (
	"testing"

	warmer "github.com/reddotpay/lambda-warmer"
	"github.com/stretchr/testify/assert"
)

func TestErrors_New(t *testing.T) {
	err := warmer.New(warmer.ErrCodeNotWarmerEvent)
	assert.EqualError(t, err, "not a lambda warmer event")
}
