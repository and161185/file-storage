package files_test

import (
	"context"
	"file-storage/internal/storage/inmemory"
	"testing"
)

func TestUpdate(t *testing.T) {

	ctx := context.Background()
	//	cfg := config.Image{Ext: "jpeg", MaxDimention: 2000}
	strg := inmemory.New()
	f, err := strg.Info(ctx, "1")
	//s := files.NewService(&cfg, strg)

	//s.Update()
}
