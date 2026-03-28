package filesystemstorage

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestActiveFiles(t *testing.T) {

	table := []struct {
		name        string
		id          string
		activeState *activeState
		expectFiles map[string]struct{}
	}{
		{
			name:        "legacy",
			id:          "id",
			activeState: nil,
			expectFiles: map[string]struct{}{
				"id." + lockExt:        {},
				"id." + activeStateExt: {},
				"id." + metadataExt:    {},
				"id." + binExt:         {},
			},
		},
		{
			name:        "ok",
			id:          "id",
			activeState: &activeState{Data: "A", Metadata: "B"},
			expectFiles: map[string]struct{}{
				"id." + lockExt:        {},
				"id." + activeStateExt: {},
				"id.B." + metadataExt:  {},
				"id.A." + binExt:       {},
			},
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			dirPath := t.TempDir()

			if tt.activeState != nil {
				b, err := json.Marshal(tt.activeState)
				if err != nil {
					t.Fatalf("activeState marshal error: %v", err)
				}
				filename := filepath.Join(dirPath, tt.id+"."+activeStateExt)
				err = os.WriteFile(filename, b, 0644)
				if err != nil {
					t.Fatalf("write activeState file error: %v", err)
				}
			}

			gotFiles, err := activeFiles(tt.id, dirPath)
			if err != nil {
				t.Errorf("read activeState file error: %v", err)
			}

			if !reflect.DeepEqual(gotFiles, tt.expectFiles) {
				t.Errorf("activeState mismatch got %v want %v", gotFiles, tt.expectFiles)
			}
		})
	}
}

func TestCollectGarbage(t *testing.T) {
	root := t.TempDir()
	path1 := filepath.Join(root, "aa", "bb")
	path2 := filepath.Join(root, "aa", "cc")
	err := os.MkdirAll(path1, 0755)
	if err != nil {
		t.Fatalf("create directiry %s error: %v", path1, err)
	}
	err = os.MkdirAll(path2, 0755)
	if err != nil {
		t.Fatalf("create directiry %s error: %v", path2, err)
	}

	files1, err := activeFiles("aabb", path1)
	if err != nil {
		t.Fatalf("activeFiles %s error: %v", path1, err)
	}
	files2, err := activeFiles("aacc", path2)
	if err != nil {
		t.Fatalf("activeFiles %s error: %v", path2, err)
	}

	for k := range files1 {
		filename := filepath.Join(path1, k)
		err := os.WriteFile(filename, []byte{}, 0644)
		if err != nil {
			t.Fatalf("write file %s error: %v", filename, err)
		}
	}
	for k := range files2 {
		filename := filepath.Join(path2, k)
		err := os.WriteFile(filename, []byte{}, 0644)
		if err != nil {
			t.Fatalf("write file %s error: %v", filename, err)
		}
	}

	log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	gc := NewGarbageCollector(root, 1*time.Minute, 5, log)

	ctx := context.Background()
	ch := make(chan *cleanupJob, 10)

	err = gc.collectGarbage(ctx, ch)
	if err != nil {
		t.Fatalf("collectGarbage error: %v", err)
	}

	close(ch)

	idmap := make(map[string]struct{})
	for j := range ch {
		if len(j.dirEntries) != 4 {
			t.Errorf("expect 4 files got %v", len(j.dirEntries))
		}
		idmap[j.id] = struct{}{}
	}

	if _, ok := idmap["aabb"]; !ok {
		t.Error("expect aabb id")
	}
	if _, ok := idmap["aacc"]; !ok {
		t.Error("expect aacc id")
	}

	ch2 := make(chan *cleanupJob, 10)
	ctx, cancel := context.WithCancel(ctx)
	cancel()
	err = gc.collectGarbage(ctx, ch2)
	close(ch2)

	if !errors.Is(err, ctx.Err()) {
		t.Errorf("got %v expect %v", err, ctx.Err())
	}
}

func TestRemoveGarbage(t *testing.T) {
	emptyJob := cleanupJob{}
	err := removeGarbage(&emptyJob)
	if err != nil {
		t.Errorf("expect nil error got %v", err)
	}

	root := t.TempDir()
	path := filepath.Join(root, "aa", "bb")
	err = os.MkdirAll(path, 0755)
	if err != nil {
		t.Fatalf("create directiry %s error: %v", path, err)
	}
	files, err := activeFiles("aabb", path)
	if err != nil {
		t.Fatalf("activeFiles %s error: %v", path, err)
	}

	for k := range files {
		filename := filepath.Join(path, k)
		b := []byte{}
		if k == "aabb."+activeStateExt {
			b, err = json.Marshal(&activeState{})
			if err != nil {
				t.Fatalf("activeState marshal error: %v", err)
			}
		}
		err := os.WriteFile(filename, b, 0644)
		if err != nil {
			t.Fatalf("write file %s error: %v", filename, err)
		}
	}

	garbage := make(map[string]struct{})
	garbage["aabb.C.bin"] = struct{}{}
	garbage["aabb.tmp"] = struct{}{}
	for k := range garbage {
		filename := filepath.Join(path, k)
		err := os.WriteFile(filename, []byte{}, 0644)
		if err != nil {
			t.Fatalf("write file %s error: %v", filename, err)
		}
	}

	filesEntries, err := os.ReadDir(path)
	if err != nil {
		t.Fatalf("subdirectory %s reading error: %v", path, err)
	}

	job := &cleanupJob{id: "aabb", dirPath: path, dirEntries: filesEntries}
	err = removeGarbage(job)
	if err != nil {
		t.Fatalf("removeGarbage error: %v", err)
	}

	filesEntriesNew, err := os.ReadDir(path)
	if err != nil {
		t.Fatalf("subdirectory %s reading error: %v", path, err)
	}

	cleanFiles := make(map[string]struct{})
	for _, v := range filesEntriesNew {
		cleanFiles[v.Name()] = struct{}{}
	}

	if !reflect.DeepEqual(files, cleanFiles) {
		t.Errorf("clean files %v doesn't match files %v", cleanFiles, files)
	}

}
