// storage/objects.go
package storage

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func SaveBlob(filePath string) (string, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	hash := fmt.Sprintf("%x", sha1.Sum(data))
	objectPath := filepath.Join(".mygit", "objects", hash)

	// If object already exists, skip writing
	if _, err := os.Stat(objectPath); os.IsNotExist(err) {
		err = ioutil.WriteFile(objectPath, data, 0644)
		if err != nil {
			return "", err
		}
	}

	return hash, nil
}
