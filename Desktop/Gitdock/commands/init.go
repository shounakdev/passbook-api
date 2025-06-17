package commands

import (
	"fmt"
	"os"
)

func InitRepo() {
	os.Mkdir(".mygit", 0755)
	os.MkdirAll(".mygit/refs/heads", 0755)
	os.Mkdir(".mygit/objects", 0755)
	os.WriteFile(".mygit/HEAD", []byte("ref: refs/heads/main\n"), 0644)

	fmt.Println("Initialized empty MyGit repository.")
}
