package commands

import (
	"bufio"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"mygit/storage"
)

func Status() {
	// 1. Load staged files from index
	staged := make(map[string]string)
	indexFile, err := os.Open(".mygit/index")
	if err == nil {
		scanner := bufio.NewScanner(indexFile)
		for scanner.Scan() {
			parts := strings.SplitN(scanner.Text(), " ", 2)
			if len(parts) == 2 {
				staged[parts[1]] = parts[0]
			}
		}
		indexFile.Close()
	}

	// 2. Load files from last commit (HEAD)
	lastCommitFiles := make(map[string]string)
	head, err := os.ReadFile(".mygit/refs/heads/main")
	if err == nil {
		commit, err := storage.LoadCommit(strings.TrimSpace(string(head)))
		if err == nil {
			lastCommitFiles = commit.Tree
		}
	}

	// 3. Scan working directory
	files, err := ioutil.ReadDir(".")
	if err != nil {
		fmt.Println("Error reading working dir:", err)
		return
	}

	fmt.Println("=== MyGit Status ===")

	for _, file := range files {
		name := file.Name()

		if name == ".mygit" || file.IsDir() {
			continue
		}

		content, err := os.ReadFile(name)
		if err != nil {
			continue
		}
		hash := fmt.Sprintf("%x", sha1.Sum(content))

		// 4. Compare and categorize
		if stagedHash, ok := staged[name]; ok {
			if hash != stagedHash {
				fmt.Printf("Modified (staged): %s\n", name)
			}
		} else if commitHash, ok := lastCommitFiles[name]; ok {
			if hash != commitHash {
				fmt.Printf("Modified (not staged): %s\n", name)
			}
		} else {
			fmt.Printf("Untracked: %s\n", name)
		}
	}

	for file := range staged {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			fmt.Printf("Deleted (staged): %s\n", file)
		}
	}
}
