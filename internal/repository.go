// internal/repository.go
package internal // Changed from "package main" to "package internal"

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Repository structure
type Repository struct {
	WorkingDir string
	GitDir     string
}

// Commit structure
type Commit struct {
	Hash     string    `json:"hash"`
	Author   string    `json:"author"`
	Date     time.Time `json:"date"`
	Message  string    `json:"message"`
	Parent   string    `json:"parent"`
	TreeHash string    `json:"tree_hash"`
}

// IndexEntry structure
type IndexEntry struct {
	FilePath string `json:"file_path"`
	Hash     string `json:"hash"`
	Modified bool   `json:"modified"`
}

// Configuration constants
const (
	GITTER_DIR  = ".gitter"
	HEAD_FILE   = "HEAD"
	INDEX_FILE  = "index"
	REFS_DIR    = "refs"
	HEADS_DIR   = "heads"
	OBJECTS_DIR = "objects"
	LOG_FILE    = "log"
)

// GetCurrentDir returns the current working directory
func GetCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return dir
}

// FindGitterRepo finds the gitter repository starting from current directory
func FindGitterRepo() (*Repository, error) {
	dir := GetCurrentDir()
	for {
		gitterPath := filepath.Join(dir, GITTER_DIR)
		if _, err := os.Stat(gitterPath); err == nil {
			return &Repository{
				WorkingDir: dir,
				GitDir:     gitterPath,
			}, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return nil, fmt.Errorf("not a gitter repository")
		}
		dir = parent
	}
}

// InitRepository initializes a new gitter repository
func InitRepository() error {
	gitterPath := filepath.Join(GetCurrentDir(), GITTER_DIR)

	// Check if already initialized
	if _, err := os.Stat(gitterPath); err == nil {
		return fmt.Errorf("repository already initialized")
	}

	// Create directory structure
	dirs := []string{
		gitterPath,
		filepath.Join(gitterPath, REFS_DIR),
		filepath.Join(gitterPath, REFS_DIR, HEADS_DIR),
		filepath.Join(gitterPath, OBJECTS_DIR),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// Create HEAD file
	headPath := filepath.Join(gitterPath, HEAD_FILE)
	if err := ioutil.WriteFile(headPath, []byte("ref: refs/heads/main\n"), 0644); err != nil {
		return err
	}

	// Create empty index
	indexPath := filepath.Join(gitterPath, INDEX_FILE)
	emptyIndex := []IndexEntry{}
	indexData, err := json.Marshal(emptyIndex)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(indexPath, indexData, 0644); err != nil {
		return err
	}

	// Create log file
	logPath := filepath.Join(gitterPath, LOG_FILE)
	if err := ioutil.WriteFile(logPath, []byte(""), 0644); err != nil {
		return err
	}

	return nil
}

// hashFile calculates SHA1 hash of a file
func hashFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha1.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// LoadIndex loads the current index from file
func LoadIndex() ([]IndexEntry, error) {
	repo, err := FindGitterRepo()
	if err != nil {
		return nil, err
	}

	indexPath := filepath.Join(repo.GitDir, INDEX_FILE)
	data, err := ioutil.ReadFile(indexPath)
	if err != nil {
		return nil, err
	}

	var index []IndexEntry
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, err
	}

	return index, nil
}

// SaveIndex saves the index to file
func SaveIndex(index []IndexEntry) error {
	repo, err := FindGitterRepo()
	if err != nil {
		return err
	}

	indexPath := filepath.Join(repo.GitDir, INDEX_FILE)
	data, err := json.Marshal(index)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(indexPath, data, 0644)
}

// AddFile adds a file to the index
func AddFile(filePath string) error {
	repo, err := FindGitterRepo()
	if err != nil {
		return err
	}

	// Handle glob patterns
	var files []string
	if strings.Contains(filePath, "*") {
		matches, err := filepath.Glob(filePath)
		if err != nil {
			return err
		}
		files = matches
	} else {
		files = []string{filePath}
	}

	index, err := LoadIndex()
	if err != nil {
		return err
	}

	for _, file := range files {
		// Skip if file doesn't exist
		if _, err := os.Stat(file); os.IsNotExist(err) {
			continue
		}

		// Calculate hash
		hash, err := hashFile(file)
		if err != nil {
			return err
		}

		// Update or add to index
		found := false
		for i := range index {
			if index[i].FilePath == file {
				index[i].Hash = hash
				index[i].Modified = true
				found = true
				break
			}
		}

		if !found {
			index = append(index, IndexEntry{
				FilePath: file,
				Hash:     hash,
				Modified: true,
			})
		}

		// Copy file to objects directory
		objectPath := filepath.Join(repo.GitDir, OBJECTS_DIR, hash)
		if err := copyFile(file, objectPath); err != nil {
			return err
		}
	}

	return SaveIndex(index)
}

// copyFile copies a file from source to destination
func copyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

// GetCurrentHead returns the current HEAD commit hash
func GetCurrentHead() (string, error) {
	repo, err := FindGitterRepo()
	if err != nil {
		return "", err
	}

	headPath := filepath.Join(repo.GitDir, HEAD_FILE)
	data, err := ioutil.ReadFile(headPath)
	if err != nil {
		return "", err
	}

	headRef := strings.TrimSpace(string(data))
	if strings.HasPrefix(headRef, "ref: ") {
		refPath := strings.TrimPrefix(headRef, "ref: ")
		refFile := filepath.Join(repo.GitDir, refPath)
		refData, err := ioutil.ReadFile(refFile)
		if err != nil {
			if os.IsNotExist(err) {
				return "", nil // No commits yet
			}
			return "", err
		}
		return strings.TrimSpace(string(refData)), nil
	}

	return headRef, nil
}

// UpdateHead updates the HEAD to point to a new commit
func UpdateHead(commitHash string) error {
	repo, err := FindGitterRepo()
	if err != nil {
		return err
	}

	// Update the main branch reference
	mainRef := filepath.Join(repo.GitDir, REFS_DIR, HEADS_DIR, "main")
	return ioutil.WriteFile(mainRef, []byte(commitHash+"\n"), 0644)
}

// CalculateHash calculates SHA1 hash of a string
func CalculateHash(data string) string {
	hasher := sha1.New()
	hasher.Write([]byte(data))
	return hex.EncodeToString(hasher.Sum(nil))
}

// UpdateLog appends a commit to the log file
func UpdateLog(commit Commit) error {
	repo, err := FindGitterRepo()
	if err != nil {
		return err
	}

	logPath := filepath.Join(repo.GitDir, LOG_FILE)

	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	logEntry := fmt.Sprintf("%s\n", commit.Hash)
	_, err = file.WriteString(logEntry)
	return err
}
