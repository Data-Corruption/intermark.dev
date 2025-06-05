//go:build ignore
// +build ignore

package main

import (
	"bufio"
	"fmt"
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
		confirm := prompt("🟡  This will fetch & merge upstream/main(intermark) into your current branch. Continue? [y/N] ")
		if confirm != "y" && confirm != "Y" {
			fmt.Println("Update aborted.")
			return
		}
		if !gitRemoteExists("upstream") {
			run("git", "remote", "add", "upstream", upstreamRepo)
		}
		run("git", "fetch", "upstream")
		run("git", "merge", "--ff-only", "upstream/main")
		run("git", "merge", "upstream/main")
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
			fmt.Println("🟢 .gitattributes already contains LFS settings.")
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
	fmt.Println("🟢 Updated .gitattributes for LFS.")
}
