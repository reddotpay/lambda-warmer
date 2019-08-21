package warmer_test

import (
	"context"
	"testing"

	warmer "github.com/reddotpay/lambda-warmer"
	"github.com/stretchr/testify/assert"
)

func TestWarmer_Handler(t *testing.T) {
	err := warmer.Handler(context.Background(), map[string]interface{}{
		"warmer":      true,
		"concurrency": 1,
	})
	assert.NoError(t, err)
	assert.True(t, warmer.Warm)
}

func TestWarmer_Handler_Concurreny(t *testing.T) {
	err := warmer.Handler(context.Background(), map[string]interface{}{
		"warmer":      true,
		"concurrency": 3,
	})
	assert.NoError(t, err)
	assert.True(t, warmer.Warm)
}

func TestWarmer_Handler_ConcurrentInvocation(t *testing.T) {
	err := warmer.Handler(context.Background(), map[string]interface{}{
		"warmer":            true,
		"warmerInvocation":  2,
		"warmerConcurrency": 3,
	})
	assert.NoError(t, err)
	assert.True(t, warmer.Warm)
}

func TestWarmer_Handler_NotWarmerEvent(t *testing.T) {
	err := warmer.Handler(context.Background(), map[string]interface{}{})
	assert.EqualError(t, err, "not a lambda warmer event")
	assert.True(t, warmer.Warm)
}
