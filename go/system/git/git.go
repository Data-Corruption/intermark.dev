package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"intermark/go/system"

	"github.com/Data-Corruption/rlog/logger"
)

// Version returns the version of git installed on the system.
// Also serves as a check to see if git is installed.
func Version(ctx context.Context) (string, error) {
	// run git --version
	cmd := exec.CommandContext(ctx, "git", "--version")
	output, err := system.RunCommand(ctx, cmd)
	if err != nil {
		return "", fmt.Errorf("error running git --version: %w", err)
	}
	// parse and return the version
	re := regexp.MustCompile(`\d+(?:\.\d+)+`) // matches version numbers (one or more groups of digits separated by dots, e.g. "2.34.1" or "10.0.18362.1")
	version := re.FindString(output)
	if version == "" {
		return "", fmt.Errorf("error parsing git version: %s", output)
	}
	return version, nil
}

// DebugInfo returns the git version, the current commit hash, and the current upstream commit hash.
func DebugInfo(ctx context.Context, repoDirPath string) (string, string, string, error) {
	// get git version
	version, err := Version(ctx)
	if err != nil {
		return "", "", "", fmt.Errorf("error getting git version: %w", err)
	}

	// get current commit hash
	currentCommit, err := GetCommitHash(ctx, repoDirPath)
	if err != nil {
		logger.Errorf(ctx, "error getting current commit hash: %v", err)
	}

	// get upstream commit hash
	upstreamCommit, err := GetUpstreamCommitHash(ctx, repoDirPath)
	if err != nil {
		logger.Errorf(ctx, "error getting upstream commit hash: %v", err)
	}

	return version, currentCommit, upstreamCommit, nil
}

// FileChanged checks if the given file (relative to repo dir) has changed since the given commit.
// If given commit is empty, it returns true.
func FileChanged(ctx context.Context, repoDirPath, filePath, commitHash string) (bool, error) {
	if err := ensureGitDir(repoDirPath); err != nil {
		return false, err
	}

	if commitHash == "" {
		return true, nil
	}

	// run git diff
	cmd := exec.CommandContext(ctx, "git", "diff", "--exit-code", commitHash, "--", filePath)
	cmd.Dir = repoDirPath
	if _, err := system.RunCommand(ctx, cmd); err != nil {
		// If the error is an exit error and the exit code is 1, the file has changed. Otherwise, return the error
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 1 {
				return true, nil
			}
		}
		return false, fmt.Errorf("error running git diff: %w", err)
	}
	return false, nil
}

var oidRe = regexp.MustCompile(`oid sha256:([a-f0-9]{64})`)

// LFSFileChanged checks if the actual LFS object for a file changed between HEAD and the given commit.
// Returns true if the oid hashes differ, or if the file was not tracked in the given commit.
func LFSFileChanged(ctx context.Context, repoDirPath, filePath, commitHash string) (bool, error) {
	if err := ensureGitDir(repoDirPath); err != nil {
		return false, err
	}

	if commitHash == "" {
		return true, nil
	}

	// helper to get LFS OID from a commit
	getOID := func(commit string) (string, error) {
		cmd := exec.CommandContext(ctx, "git", "show", fmt.Sprintf("%s:%s", commit, filePath))
		cmd.Dir = repoDirPath
		out, err := system.RunCommand(ctx, cmd)
		if err != nil {
			if strings.Contains(out, "fatal:") {
				return "", nil // file absent in this commit
			}
			return "", err
		}
		// find oid sha256:...
		match := oidRe.FindStringSubmatch(out)
		if len(match) < 2 {
			return "", nil // not an LFS file (or malformed)
		}
		return match[1], nil
	}

	// get OIDs
	headOID, err := getOID("HEAD")
	if err != nil {
		return false, fmt.Errorf("getting HEAD LFS oid: %w", err)
	}
	baseOID, err := getOID(commitHash)
	if err != nil {
		return false, fmt.Errorf("getting base commit LFS oid: %w", err)
	}

	return headOID != baseOID, nil
}

func GetCommitHash(ctx context.Context, repoDirPath string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	cmd.Dir = repoDirPath
	return system.RunCommand(ctx, cmd)
}

func GetUpstreamCommitHash(ctx context.Context, repoDirPath string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "@{upstream}")
	cmd.Dir = repoDirPath
	return system.RunCommand(ctx, cmd)
}

// Clone clones the given repository into the given directory.
// If the directory already exists, it will be removed first.
func Clone(ctx context.Context, repoURL, repoDirPath string) error {
	// remove the directory if it exists
	if _, err := os.Stat(repoDirPath); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("error checking if repository path exists: %w", err)
		}
	} else {
		if err := os.RemoveAll(repoDirPath); err != nil {
			return fmt.Errorf("error removing existing repository path: %w", err)
		}
	}
	// create the directory
	if err := os.MkdirAll(repoDirPath, 0o755); err != nil {
		return fmt.Errorf("error creating repository path: %w", err)
	}

	logger.Debugf(ctx, "Attempting to clone %s into %s", repoURL, repoDirPath)

	// clone the repo
	cmd := exec.CommandContext(ctx, "git", "clone", repoURL, ".")
	cmd.Env = getENV(ctx)
	cmd.Dir = repoDirPath
	if _, err := system.RunCommand(ctx, cmd); err != nil {
		return fmt.Errorf("error running git clone: %w", err)
	}

	return nil
}

// Fetch fetches the latest changes from the given branch for the given repository.
func Fetch(ctx context.Context, repoDirPath, branch string) error {
	if err := ensureGitDir(repoDirPath); err != nil {
		return err
	}

	// fetch latest changes
	cmd := exec.CommandContext(ctx, "git", "fetch", "origin", branch)
	cmd.Env = getENV(ctx)
	cmd.Dir = repoDirPath
	_, err := system.RunCommand(ctx, cmd)
	return err
}

// Reset resets the given repository to the latest local commit on the given branch.
//
// Example usage:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
//	defer cancel()
//	err := git.Reset(ctx, "/path/to/repo", "main", true)
//
//	if err != nil {
//		log.Fatalf("Failed to reset repository: %v", err)
//	}
func Reset(ctx context.Context, repoDirPath, branch string, hard bool) error {
	if err := ensureGitDir(repoDirPath); err != nil {
		return err
	}

	args := []string{"reset", "origin/" + branch}
	if hard {
		args = append(args, "--hard")
	}

	// reset
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Env = getENV(ctx)
	cmd.Dir = repoDirPath
	_, err := system.RunCommand(ctx, cmd)
	return err
}

// LfsPull pulls the latest changes for LFS files in the given repository.
// Typically used after a clone or reset.
func LfsPull(ctx context.Context, repoDirPath string) error {
	if err := ensureGitDir(repoDirPath); err != nil {
		return err
	}

	// pull lfs files
	cmd := exec.CommandContext(ctx, "git", "lfs", "pull")
	cmd.Env = getENV(ctx)
	cmd.Dir = repoDirPath
	_, err := system.RunCommand(ctx, cmd)
	return err
}

func ensureGitDir(repoDirPath string) error {
	if info, err := os.Stat(filepath.Join(repoDirPath, ".git")); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("repository path does not contain a .git directory")
		}
		return fmt.Errorf("error checking if repository path exists: %w", err)
	} else if !info.IsDir() {
		return fmt.Errorf("repository path is not a directory")
	}
	return nil
}

// helper func for setting git ssh command
func getENV(ctx context.Context) []string {
	// ensure the SSH key exists
	if _, err := os.Stat(filepath.Join(os.Getenv("HOME"), ".ssh", "id_ed25519_intermark")); err != nil {
		if os.IsNotExist(err) {
			logger.Debug(ctx, "SSH key not found at ~/.ssh/id_ed25519_intermark, please ensure it exists")
			return os.Environ()
		}
		logger.Errorf(ctx, "issue checking SSH key: %v", err)
		return os.Environ()
	}
	return append(os.Environ(),
		"GIT_SSH_COMMAND=ssh -i ~/.ssh/id_ed25519_intermark -o IdentitiesOnly=yes",
	)
}
