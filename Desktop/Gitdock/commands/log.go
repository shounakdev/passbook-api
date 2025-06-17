package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"mygit/storage"
)

func ShowLog() {
	headFile := ".mygit/refs/heads/main"
	headHashBytes, err := os.ReadFile(headFile)
	if err != nil {
		fmt.Println("No commits found.")
		return
	}
	current := string(headHashBytes)

	for current != "" {
		commitPath := filepath.Join(".mygit", "objects", current)
		content, err := os.ReadFile(commitPath)
		if err != nil {
			fmt.Println("Failed to read commit object:", err)
			return
		}

		var commit storage.Commit
		err = json.Unmarshal(content, &commit)
		if err != nil {
			fmt.Println("Corrupted commit object:", err)
			return
		}

		fmt.Println("====================================")
		fmt.Println("Commit:", current)
		fmt.Println("Author:", commit.Author)
		fmt.Println("Date:  ", commit.Timestamp)
		fmt.Println("Message:", commit.Message)
		fmt.Println("====================================")
		fmt.Println()

		current = commit.Parent
	}
}
