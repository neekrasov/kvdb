package client

import (
	"errors"
	"io"
)

type MockWriter interface {
	io.Writer
	String() string
}

type mockWriter struct {
}

func (*mockWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("some write error")
}

func (*mockWriter) String() string {
	return ""
}
