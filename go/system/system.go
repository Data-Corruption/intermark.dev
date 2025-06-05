package system

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/Data-Corruption/rlog/logger"
)

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// RunCommand runs a command and returns its output as a string.
// It also handles context cancellation timeout errors and logging.
func RunCommand(ctx context.Context, cmd *exec.Cmd) (string, error) {
	bytes, err := cmd.CombinedOutput()
	output := ansiRegexp.ReplaceAllString(strings.TrimSpace(string(bytes)), "")

	if err != nil {
		if errors.Is(err, context.Canceled) {
			logger.Warnf(ctx, "command was cancelled: %s", cmd.String())
			return output, fmt.Errorf("command cancelled: %w", err)
		}
		if errors.Is(err, context.DeadlineExceeded) {
			logger.Warnf(ctx, "command timed out: %s", cmd.String())
			return output, fmt.Errorf("command timed out: %w", err)
		}
		return output, err
	}
	logger.Debugf(ctx, "Command:\n\n%s\n\nOutput:\n\n%s\n\n", cmd.String(), output)
	return output, nil
}
