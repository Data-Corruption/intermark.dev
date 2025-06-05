package lunrjs

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"intermark/go/files"
	"intermark/go/html"
	"intermark/go/system"

	"github.com/minio/sha256-simd"
)

const (
	SCRIPT_PATH = "./go/system/lunrjs/gen_index.js"
	DOCS_PATH   = "./public/.meta/search-pre-index.json"
	INDEX_PATH  = "./public/.meta/search-index.json"
)

// Run returns:
//   - JSON.stringify({ index: JSON.stringify(idx), hash }) as a gzipped byte array.
//   - the hash of the index as a string.
//   - an error if any.
func Run(ctx context.Context, docs *[]html.Doc) ([]byte, string, error) {
	// write the search docs file
	if err := os.MkdirAll(filepath.Dir(DOCS_PATH), 0o755); err != nil {
		return nil, "", fmt.Errorf("error creating search docs directory %s: %w", DOCS_PATH, err)
	}
	if err := files.SaveJSON(DOCS_PATH, &docs, 0o644); err != nil {
		return nil, "", fmt.Errorf("error writing search docs file %s: %w", DOCS_PATH, err)
	}

	// run the tailwindcss command
	cmd := exec.CommandContext(ctx, "node", SCRIPT_PATH, "-q")
	cout, err := system.RunCommand(ctx, cmd)
	if err != nil {
		return nil, "", fmt.Errorf("error running lunrjs script %s: %w\n%s", SCRIPT_PATH, err, cout)
	}
	// read and gzip the output
	out, err := os.ReadFile(INDEX_PATH)
	if err != nil {
		return nil, "", err
	}
	// calculate the hash of the output
	sum := sha256.Sum256(out)
	hash := hex.EncodeToString(sum[:])
	// gzip the output
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	_, err = gz.Write(out)
	if err != nil {
		return nil, "", err
	}
	gz.Close()
	// return the hash and the gzipped output
	return b.Bytes(), hash, nil
}
