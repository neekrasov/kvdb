package tcp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWithClientIdleTimeout(t *testing.T) {
	t.Parallel()

	idleTimeout := time.Second
	option := WithClientIdleTimeout(time.Second)

	var client Client
	option(&client)

	assert.Equal(t, idleTimeout, client.idleTimeout)
}

func TestWithClientBufferSize(t *testing.T) {
	t.Parallel()

	var bufferSize uint = 10 << 10
	option := WithClientBufferSize(bufferSize)

	var client Client
	option(&client)

	assert.Equal(t, bufferSize, uint(client.bufferSize))
}
