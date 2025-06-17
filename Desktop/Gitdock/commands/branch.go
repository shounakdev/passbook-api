package commands

import (
	"fmt"
	"io/ioutil"
)

func Branch() {
	branches, err := ioutil.ReadDir(".mygit/refs/heads")
	if err != nil {
		fmt.Println("Error reading branches:", err)
		return
	}
	fmt.Println("Available branches:")
	for _, branch := range branches {
		fmt.Println("- " + branch.Name())
	}
}
