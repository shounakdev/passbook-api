package commands

import (
	"fmt"
	"os"

	"mygit/rbac"
)

func Checkout(username, branchName string) {
	if !rbac.CanAccessBranch(username, branchName, "view") {
		fmt.Printf("❌ User '%s' is not allowed to view branch '%s'.\n", username, branchName)
		return
	}

	refPath := ".mygit/refs/heads/" + branchName
	if _, err := os.Stat(refPath); os.IsNotExist(err) {
		fmt.Println("Branch does not exist:", branchName)
		return
	}

	err := os.WriteFile(".mygit/HEAD", []byte("ref: refs/heads/"+branchName), 0644)
	if err != nil {
		fmt.Println("Error switching branch:", err)
		return
	}

	fmt.Println("✅ Switched to branch:", branchName)
}
