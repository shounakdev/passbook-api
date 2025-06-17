package storage

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Commit struct {
	Message   string            `json:"message"`
	Timestamp string            `json:"timestamp"`
	Author    string            `json:"author"`
	Tree      map[string]string `json:"tree"` // filename -> blob hash
	Parent    string            `json:"parent"`
}

func SaveCommit(commit Commit) (string, error) {
	data, err := json.MarshalIndent(commit, "", "  ")
	if err != nil {
		return "", err
	}

	hash := fmt.Sprintf("%x", sha1.Sum(data))
	path := filepath.Join(".mygit", "objects", hash)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.WriteFile(path, data, 0644)
		if err != nil {
			return "", err
		}
	}

	return hash, nil
}

func LoadCommit(hash string) (Commit, error) {
	path := filepath.Join(".mygit", "objects", hash)
	data, err := os.ReadFile(path)
	if err != nil {
		return Commit{}, err
	}

	var commit Commit
	err = json.Unmarshal(data, &commit)
	if err != nil {
		return Commit{}, err
	}

	return commit, nil
}
