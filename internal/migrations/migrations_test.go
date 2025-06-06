package migrations

import (
	"context"
	"testing"

	"github.com/Sugar-pack/users-manager/internal/config"
	"github.com/Sugar-pack/users-manager/pkg/logging"
)

func TestApply_DBConnectError(t *testing.T) {
	conf := &config.DB{ConnString: "invalid"}
	ctx := logging.WithContext(context.Background(), logging.GetLogger())
	if err := Apply(ctx, conf); err == nil {
		t.Fatal("expected error")
	}
}
