// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package diff_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/FollowTheProcess/test/internal/diff"
	"golang.org/x/tools/txtar"
)

func clean(text []byte) []byte {
	text = bytes.ReplaceAll(text, []byte("$\n"), []byte("\n"))
	text = bytes.TrimSuffix(text, []byte("^D\n"))
	return text
}

func Test(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("testdata", "*.txtar"))
	if err != nil {
		t.Fatalf("could not glob txtar files: %v", err)
	}
	if len(files) == 0 {
		t.Fatalf("no testdata")
	}

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			contents, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("could not read %s: %v", file, err)
			}
			// Stupid windows
			contents = bytes.ReplaceAll(contents, []byte("\r\n"), []byte("\n"))
			archive := txtar.Parse(contents)
			if len(archive.Files) != 3 || archive.Files[2].Name != "diff" {
				t.Fatalf("%s: want three files, third named \"diff\", got: %v", file, archive.Files)
			}
			diffs := diff.Diff(
				archive.Files[0].Name,
				clean(archive.Files[0].Data),
				archive.Files[1].Name,
				clean(archive.Files[1].Data),
			)
			want := clean(archive.Files[2].Data)
			if !bytes.Equal(diffs, want) {
				t.Fatalf("%s: have:\n%s\nwant:\n%s\n%s", file,
					diffs, want, diff.Diff("have", diffs, "want", want))
			}
		})
	}
}
