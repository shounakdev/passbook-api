package commands

import (
	"fmt"
	"os"

	"mygit/storage"
)

func AddFile(filename string) {
	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Println("File does not exist:", filename)
		return
	}

	// Save the file as a blob in .mygit/objects/
	hash, err := storage.SaveBlob(filename)
	if err != nil {
		fmt.Println("Error saving blob:", err)
		return
	}

	// Append the blob info to the .mygit/index (staging area)
	indexLine := fmt.Sprintf("%s %s\n", hash, filename)
	f, err := os.OpenFile(".mygit/index", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening index file:", err)
		return
	}
	defer f.Close()

	if _, err := f.WriteString(indexLine); err != nil {
		fmt.Println("Error writing to index file:", err)
		return
	}

	fmt.Printf("Added %s to index (hash: %s)\n", filename, hash)
}
