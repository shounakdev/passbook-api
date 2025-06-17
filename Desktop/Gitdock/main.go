package main

import (
	"fmt"
	"os"

	"mygit/commands"
	"mygit/rbac"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: mygit <command> <username> [args...]")
		return
	}

	cmd := os.Args[1]
	username := os.Args[2]

	// Check permission for the command
	if !rbac.HasPermission(username, cmd) {
		fmt.Printf("❌ Access denied: User '%s' is not allowed to run '%s'\n", username, cmd)
		return
	}

	switch cmd {
	case "init":
		commands.InitRepo()

	case "add":
		if len(os.Args) < 4 {
			fmt.Println("Usage: mygit add <username> <file>")
			return
		}
		commands.AddFile(os.Args[3])

	case "commit":
		if len(os.Args) < 6 {
			fmt.Println("Usage: mygit commit <username> <author> <message>")
			return
		}
		commands.CommitChanges(username, os.Args[3], os.Args[4])

	case "status":
		commands.Status()

	case "log":
		commands.ShowLog()

	case "branch":
		commands.Branch()

	case "create-branch":
		if len(os.Args) < 4 {
			fmt.Println("Usage: mygit create-branch <username> <branch-name>")
			return
		}
		commands.CreateBranch(username, os.Args[3])

	case "checkout":
		if len(os.Args) < 4 {
			fmt.Println("Usage: mygit checkout <username> <branch-name>")
			return
		}
		commands.Checkout(username, os.Args[3])

	default:
		fmt.Println("Unknown command:", cmd)
	}

}
