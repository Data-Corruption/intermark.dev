package tailwind

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"intermark/go/stringsx"
	"intermark/go/system"
)

const (
	INPUT_PATH  = "./public/.meta/app.css"
	DIST_PATH   = "./public/.meta/app.dist.css"
	OUTPUT_DIR  = "./assets/css"
	OUTPUT_PATH = "./assets/css/out.css"
)

// Version returns the version of Tailwind CSS CLI installed via npx.
// Also serves as a check to see if git is installed.
func Version(ctx context.Context) (string, error) {
	/*
		-h first line out will be like `≈ tailwindcss v4.0.9` if installed. Else will be like:
		```
		Need to install the following packages:
		  @tailwindcss/cli@4.1.5
		Ok to proceed? (y)
		...
		```
	*/
	cmd := exec.CommandContext(ctx, "npx", "--no-install", "@tailwindcss/cli", "-h")
	cmdOut, err := system.RunCommand(ctx, cmd)
	if err != nil {
		return "", err
	}

	// get the first line
	lines := strings.Split(cmdOut, "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("error parsing tailwindcss version: %s", cmdOut)
	}
	if strings.Contains(lines[0], "Need to install") || !strings.Contains(lines[0], "≈") {
		return "", fmt.Errorf("tailwindcss is not installed, please run `npm install`")
	}

	// parse and return the version
	re := regexp.MustCompile(`\d+(?:\.\d+)+`) // matches version numbers (one or more groups of digits separated by dots, e.g. "2.34.1" or "10.0.18362.1")
	version := re.FindString(lines[0])
	if version == "" {
		return "", fmt.Errorf("error parsing tailwindcss version: %s", lines[0])
	}
	return version, nil
}

func Run(ctx context.Context, pathToHash map[string]string) (string, error) {
	// ensure the output directory exists
	if err := os.MkdirAll(OUTPUT_DIR, 0o755); err != nil {
		return "", fmt.Errorf("error creating output directory: %w", err)
	}
	// if pathToHash is not nil, replace the links in app.ccs, write to dist.css, use that as input
	inPath := INPUT_PATH
	if pathToHash != nil {
		// read the input file
		data, err := os.ReadFile(INPUT_PATH)
		if err != nil {
			return "", fmt.Errorf("error reading input file: %w", err)
		}
		// replace the links
		replaced := stringsx.FastLinkReplace(string(data), pathToHash)
		// write the output file
		if err := os.WriteFile(DIST_PATH, []byte(replaced), 0o644); err != nil {
			return "", fmt.Errorf("error writing output file: %w", err)
		}
		inPath = DIST_PATH
	}
	// run the tailwindcss command
	cmd := exec.CommandContext(ctx, "npx", "@tailwindcss/cli", "-i", inPath, "-o", OUTPUT_PATH, "-m")
	return system.RunCommand(ctx, cmd)
}
