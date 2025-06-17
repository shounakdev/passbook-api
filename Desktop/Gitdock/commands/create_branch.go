package commands

import (
	"fmt"
	"mygit/rbac"
	"os"
	"strings"
)

func CreateBranch(username, branchName string) {

	if !rbac.CanCreateBranch(username) {
		fmt.Printf("❌ User '%s' is not allowed to create new branches.\n", username)
		return
	}
	headRef := ".mygit/HEAD"
	data, err := os.ReadFile(headRef)
	if err != nil {
		fmt.Println("Error reading HEAD:", err)
		return
	}

	currentRef := strings.TrimSpace(string(data))
	if !strings.HasPrefix(currentRef, "ref: ") {
		fmt.Println("Invalid HEAD")
		return
	}
	currentBranchRef := strings.TrimPrefix(currentRef, "ref: ")

	commitHash, err := os.ReadFile(".mygit/" + currentBranchRef)
	if err != nil {
		fmt.Println("Error reading current branch commit:", err)
		return
	}

	newRefPath := ".mygit/refs/heads/" + branchName
	err = os.WriteFile(newRefPath, commitHash, 0644)
	if err != nil {
		fmt.Println("Error creating branch:", err)
		return
	}

	fmt.Println("✅ Branch created:", branchName)
}
