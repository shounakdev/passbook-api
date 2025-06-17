package rbac

import (
	"encoding/json"
	"fmt"
	"os"
)

// Role -> Allowed Commands
var rolePermissions = map[string][]string{
	"admin":     {"commit", "status", "log", "add", "init", "push", "pull", "create-branch", "checkout", "branch"},
	"developer": {"commit", "status", "log", "add", "branch"},
	"viewer":    {"status", "log"},
}

func HasPermission(username, command string) bool {
	rolesFile := ".mygit/roles.json"
	data, err := os.ReadFile(rolesFile)
	if err != nil {
		fmt.Println("⚠️  Error reading roles.json:", err)
		return false
	}

	var userRoles map[string]string
	if err := json.Unmarshal(data, &userRoles); err != nil {
		fmt.Println("⚠️  Error parsing roles.json:", err)
		return false
	}

	role := userRoles[username]
	perms := rolePermissions[role]
	for _, p := range perms {
		if p == command {
			return true
		}
	}
	return false
}

func CanCreateBranch(user string) bool {
	data, err := os.ReadFile(".mygit/branch_permissions.json")
	if err != nil {
		return true // default: allowed
	}

	var perms map[string]struct {
		Global struct {
			CanCreateBranch bool `json:"canCreateBranch"`
		} `json:"global"`
	}

	if err := json.Unmarshal(data, &perms); err != nil {
		return true // fallback: allowed
	}

	return perms[user].Global.CanCreateBranch
}

func CanAccessBranch(user, branch, mode string) bool {
	// mode = "view" or "edit"
	data, err := os.ReadFile(".mygit/branch_permissions.json")
	if err != nil {
		return true // default: allowed
	}

	var perms map[string]struct {
		BranchAccess map[string]string `json:"branchAccess"`
	}

	if err := json.Unmarshal(data, &perms); err != nil {
		return true
	}

	level := perms[user].BranchAccess[branch]
	if level == "edit" {
		return true
	}
	if level == "view" && mode == "view" {
		return true
	}

	return false
}
