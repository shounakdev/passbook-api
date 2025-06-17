package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"mygit/rbac"
	"mygit/storage"
)

func CommitChanges(username, author, message string) {
	indexPath := ".mygit/index"
	file, err := os.Open(indexPath)
	if err != nil {
		fmt.Println("Nothing to commit. Staging area is empty.")
		return
	}
	defer file.Close()

	tree := make(map[string]string)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), " ", 2)
		if len(parts) == 2 {
			tree[parts[1]] = parts[0]
		}
	}

	// 🔍 Determine current branch
	headData, err := os.ReadFile(".mygit/HEAD")
	if err != nil {
		fmt.Println("Error reading HEAD:", err)
		return
	}
	headRef := strings.TrimPrefix(strings.TrimSpace(string(headData)), "ref: ")
	branchName := strings.TrimPrefix(headRef, "refs/heads/")

	// 🔒 Check if user can edit this branch
	if !rbac.CanAccessBranch(username, branchName, "edit") {
		fmt.Printf("❌ User '%s' is not allowed to edit branch '%s'.\n", username, branchName)
		return
	}

	// Read current HEAD commit hash
	parent := ""
	if content, err := os.ReadFile(".mygit/" + headRef); err == nil {
		parent = strings.TrimSpace(string(content))
	}

	commit := storage.Commit{
		Message:   message,
		Timestamp: time.Now().Format(time.RFC3339),
		Author:    author,
		Tree:      tree,
		Parent:    parent,
	}

	hash, err := storage.SaveCommit(commit)
	if err != nil {
		fmt.Println("Error saving commit:", err)
		return
	}

	// Update branch pointer
	err = os.WriteFile(".mygit/"+headRef, []byte(hash), 0644)
	if err != nil {
		fmt.Println("Error writing branch pointer:", err)
		return
	}

	// Clear index
	os.Remove(indexPath)

	fmt.Printf("✅ Committed to '%s' with hash %s\n", branchName, hash)
}
