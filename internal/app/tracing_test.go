package app

import (
	"context"
	"testing"

	"github.com/Sugar-pack/users-manager/pkg/logging"
)

func TestNewResource(t *testing.T) {
	r, err := newResource()
	if err != nil || r == nil {
		t.Fatalf("newResource error: %v", err)
	}
}

func TestInitTracing(t *testing.T) {
	tp, err := InitTracing(context.Background(), logging.GetLogger())
	if err != nil || tp == nil {
		t.Fatalf("InitTracing error: %v", err)
	}
	_ = tp.Shutdown(context.Background())
}
