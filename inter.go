//go:build ignore
// +build ignore

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	binDir       = "bin"
	upstreamRepo = "https://github.com/Data-Corruption/Intermark.git"
	binName      = "intermark"
	goMainPath   = "./go/main"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run inter.go [setup|clean|build|edit|prod|update_intermark]")
		os.Exit(1)
	}
	cmd := os.Args[1]

	switch cmd {
	case "setup":
		run("go", "mod", "tidy")
		run("npm", "install")
		fmt.Println("🟢 Dependencies installed.")
		gitAddLFS()
		checkGitLFS()
	case "clean":
		clean()
	case "build":
		clean()
		build()
	case "edit":
		clean()
		build()
		run(binOut(), "-e")
	case "prod":
		clean()
		build()
		run(binOut())
	case "update_intermark":
		updateIm()
	default:
		fmt.Println("Unknown command:", cmd)
		os.Exit(1)
	}
}

// Helpers

func clean() {
	rmDir(binDir)
	os.MkdirAll(binDir, 0755)
	fmt.Println("🟢 Cleaned binary directory")
}

func build() {
	run("go", "build", "-o", binOut(), goMainPath)
	fmt.Printf("🟢 Built %s\n", binOut())
}

func binOut() string {
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	return filepath.Join(binDir, binName+ext)
}

func updateIm() {
	confirm := prompt("🟡  This will fetch & merge upstream/main(intermark) into your current branch. Continue? [y/N] ")
	if confirm != "y" && confirm != "Y" {
		fmt.Println("Update aborted.")
		return
	}
	if !gitRemoteExists("upstream") {
		run("git", "remote", "add", "upstream", upstreamRepo)
	}

	run("git", "fetch", "upstream")

	mergeArgs := []string{"merge", "--allow-unrelated-histories", "upstream/main"}

	cmd := exec.Command("git", mergeArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if bytes.Contains(output, []byte("CONFLICT")) {
			fmt.Println("🟡 Merge completed, but there are conflicts. “CONFLICT” markers have been added to your files.")
			fmt.Println("Please open your preferred merge editor (e.g. `git mergetool`) to resolve them, then `git add` and `git commit`.")
			return
		}
		log.Fatalf("🔴 Merge failed: %v\nOutput: %s", err, output)
	}

	fmt.Println("🟢 Merge succeeded with no conflicts.")
}

func run(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "🔴 %s %v failed: %v\n", name, args, err)
		os.Exit(1)
	}
}

func rmDir(path string) {
	if err := os.RemoveAll(path); err != nil {
		fmt.Fprintf(os.Stderr, "🔴 Failed to remove %s: %v\n", path, err)
		os.Exit(1)
	}
}

func prompt(msg string) string {
	fmt.Print(msg)
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

func gitRemoteExists(name string) bool {
	out, err := exec.Command("git", "remote").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), name)
}

func gitAddLFS() {
	const attrLine = "assets/** filter=lfs diff=lfs merge=lfs -text"
	gitAttrPath := filepath.Join(".", ".gitattributes")

	// read existing .gitattributes if exists
	var lines []string
	if data, err := os.ReadFile(gitAttrPath); err != nil {
		if os.IsNotExist(err) {
			lines = []string{}
		} else {
			fmt.Fprintf(os.Stderr, "🔴 Failed to read .gitattributes: %v\n", err)
			os.Exit(1)
		}
	} else {
		content := string(data)
		lines = strings.Split(strings.TrimRight(content, "\n"), "\n")
	}

	// Append the LFS line if missing
	for _, l := range lines {
		if l == attrLine {
			fmt.Println("🟢 .gitattributes is configured for LFS.")
			return
		}
	}
	lines = append(lines, attrLine)

	// Write back with a trailing newline
	out := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(gitAttrPath, []byte(out), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "🔴 Failed to write .gitattributes for LFS: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("🟢 .gitattributes is configured for LFS.")
}

func checkGitLFS() {
	// should work on all platforms
	_, err := exec.LookPath("git-lfs")
	if err != nil {
		fmt.Println("🔴 Git LFS was not detected on your system.")
		fmt.Println("You can install it as follows:")
		switch runtime.GOOS {
		case "darwin":
			fmt.Println("  • macOS (with Homebrew):")
			fmt.Println("      brew install git-lfs")
			fmt.Println("    Then run: git lfs install")
		case "linux":
			fmt.Println("  • Debian/Ubuntu:")
			fmt.Println("      sudo apt-get update && sudo apt-get install git-lfs")
			fmt.Println("    Then run: git lfs install")
			fmt.Println()
			fmt.Println("  • Fedora/CentOS (with dnf/yum):")
			fmt.Println("      sudo dnf install git-lfs  # or sudo yum install git-lfs")
			fmt.Println("    Then run: git lfs install")
			fmt.Println()
			fmt.Println("  • From the official binary (all distros):")
			fmt.Println("      curl -s https://packagecloud.io/install/repositories/github/git-lfs/script.deb.sh | sudo bash")
			fmt.Println("      sudo apt-get install git-lfs  # or sudo yum install git-lfs")
			fmt.Println("    Then run: git lfs install")
		case "windows":
			fmt.Println("  • Windows (using Chocolatey):")
			fmt.Println("      choco install git-lfs")
			fmt.Println("    Then run: git lfs install")
			fmt.Println()
			fmt.Println("  • Or download the Windows installer:")
			fmt.Println("      https://github.com/git-lfs/git-lfs/releases")
			fmt.Println("    and run the .msi, then open a new terminal and run: git lfs install")
		default:
			fmt.Println("  • Visit https://git-lfs.github.com/ for your platform and follow the instructions there.")
		}
		return
	}
	fmt.Println("🟢 Found Git LFS")
}
