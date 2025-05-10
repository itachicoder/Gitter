// internal/repository/operations.go
package internal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pmezard/go-difflib/difflib"
)

// ShowStatus displays the current repository status
func ShowStatus() error {
	repo, err := FindGitterRepo()
	if err != nil {
		return err
	}

	index, err := LoadIndex()
	if err != nil {
		return err
	}

	// Get all files in working directory
	workingFiles := make(map[string]string)
	err = filepath.Walk(repo.WorkingDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip gitter directory
		if strings.Contains(path, GITTER_DIR) {
			return nil
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(repo.WorkingDir, path)
		if err != nil {
			return err
		}

		hash, err := hashFile(path)
		if err != nil {
			return err
		}

		workingFiles[relPath] = hash
		return nil
	})
	if err != nil {
		return err
	}

	// Classify files
	staged := []string{}
	notStaged := []string{}
	untracked := []string{}

	// Check indexed files
	indexedFiles := make(map[string]IndexEntry)
	for _, entry := range index {
		indexedFiles[entry.FilePath] = entry
		if entry.Modified {
			staged = append(staged, entry.FilePath)
		}
	}

	// Check all working files
	for filePath, currentHash := range workingFiles {
		if entry, exists := indexedFiles[filePath]; exists {
			// File is tracked
			if !entry.Modified && entry.Hash != currentHash {
				notStaged = append(notStaged, filePath)
			}
		} else {
			// File is untracked
			untracked = append(untracked, filePath)
		}
	}

	// Print status
	if len(staged) > 0 {
		fmt.Println("Changes to be committed:")
		for _, file := range staged {
			fmt.Printf("  modified: %s\n", file)
		}
		fmt.Println()
	}

	if len(notStaged) > 0 {
		fmt.Println("Changes not staged for commit:")
		for _, file := range notStaged {
			fmt.Printf("  modified: %s\n", file)
		}
		fmt.Println()
	}

	if len(untracked) > 0 {
		fmt.Println("Untracked files:")
		for _, file := range untracked {
			fmt.Printf("  %s\n", file)
		}
	}

	if len(staged) == 0 && len(notStaged) == 0 && len(untracked) == 0 {
		fmt.Println("nothing to commit, working tree clean")
	}

	return nil
}

// CommitChanges creates a new commit
// CommitChanges creates a new commit
func CommitChanges(message string, all bool) error {
	repo, err := FindGitterRepo()
	if err != nil {
		return err
	}

	index, err := LoadIndex()
	if err != nil {
		return err
	}

	// If -a flag is used, add all modified files
	if all {
		err = filepath.Walk(repo.WorkingDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip directories and gitter directory
			if info.IsDir() || strings.Contains(path, GITTER_DIR) {
				return nil
			}

			relPath, err := filepath.Rel(repo.WorkingDir, path)
			if err != nil {
				return err
			}

			// Check if file is already tracked
			var isTracked bool
			for _, entry := range index {
				if entry.FilePath == relPath {
					isTracked = true
					break
				}
			}

			// Only add tracked files for -a flag
			if isTracked {
				return AddFile(relPath)
			}

			return nil
		})
		if err != nil {
			return err
		}

		// Reload index after adding files
		index, err = LoadIndex()
		if err != nil {
			return err
		}
	}

	// Check if there are staged changes
	var stagedFiles []IndexEntry
	for _, entry := range index {
		if entry.Modified {
			stagedFiles = append(stagedFiles, entry)
		}
	}

	if len(stagedFiles) == 0 {
		return fmt.Errorf("nothing to commit")
	}

	// Create tree object and save it
	treeData, err := json.Marshal(stagedFiles)
	if err != nil {
		return err
	}
	treeHash := CalculateHash(string(treeData))

	// Save tree object to objects directory
	treePath := filepath.Join(repo.GitDir, OBJECTS_DIR, treeHash)
	if err := ioutil.WriteFile(treePath, treeData, 0644); err != nil {
		return err
	}

	// Create commit object
	commit := Commit{
		Hash:     "",     // Will be calculated
		Author:   "user", // You can make this configurable
		Date:     time.Now(),
		Message:  message,
		TreeHash: treeHash, // Use the saved tree hash
	}

	// Get parent commit (current HEAD)
	head, err := GetCurrentHead()
	if err != nil {
		return err
	}
	commit.Parent = head

	// Calculate commit hash
	commitData, err := json.Marshal(commit)
	if err != nil {
		return err
	}
	commit.Hash = CalculateHash(string(commitData))

	// Save commit object
	commitPath := filepath.Join(repo.GitDir, OBJECTS_DIR, commit.Hash)
	if err := ioutil.WriteFile(commitPath, commitData, 0644); err != nil {
		return err
	}

	// Update HEAD
	if err := UpdateHead(commit.Hash); err != nil {
		return err
	}

	// Update log
	if err := UpdateLog(commit); err != nil {
		return err
	}

	// Reset index (mark all as not modified)
	for i := range index {
		index[i].Modified = false
	}

	if err := SaveIndex(index); err != nil {
		return err
	}

	fmt.Printf("[main %s] %s\n", commit.Hash[:7], commit.Message)
	return nil
}

// ShowDiff displays differences between HEAD and working tree
func ShowDiff(path string) error {
	repo, err := FindGitterRepo()
	if err != nil {
		return err
	}

	// Get current HEAD
	head, err := GetCurrentHead()
	if err != nil {
		return err
	}

	// If no commit exists, return error
	if head == "" {
		return fmt.Errorf("no commits yet")
	}

	// Load head commit
	commitPath := filepath.Join(repo.GitDir, OBJECTS_DIR, head)
	commitData, err := ioutil.ReadFile(commitPath)
	if err != nil {
		return err
	}

	var commit Commit
	if err := json.Unmarshal(commitData, &commit); err != nil {
		return err
	}

	// Get files to check
	var filesToCheck []string

	if path == "" {
		// Check all files in working directory
		err = filepath.Walk(repo.WorkingDir, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && !strings.Contains(p, GITTER_DIR) {
				relPath, err := filepath.Rel(repo.WorkingDir, p)
				if err != nil {
					return err
				}
				filesToCheck = append(filesToCheck, relPath)
			}
			return nil
		})
		if err != nil {
			return err
		}
	} else {
		// Check specific file or directory
		stat, err := os.Stat(path)
		if err != nil {
			return err
		}

		if stat.IsDir() {
			err = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if !info.IsDir() {
					relPath, err := filepath.Rel(repo.WorkingDir, p)
					if err != nil {
						return err
					}
					filesToCheck = append(filesToCheck, relPath)
				}
				return nil
			})
			if err != nil {
				return err
			}
		} else {
			filesToCheck = []string{path}
		}
	}

	// Show diff for each file
	for _, file := range filesToCheck {
		if err := showFileDiff(repo, commit, file); err != nil {
			continue // Skip files that don't exist in HEAD
		}
	}

	return nil
}

// showFileDiff displays diff for a single file
// showFileDiff displays diff for a single file
func showFileDiff(repo *Repository, commit Commit, filePath string) error {
	// Get current file content
	currentPath := filepath.Join(repo.WorkingDir, filePath)
	currentContent, err := ioutil.ReadFile(currentPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File was deleted
			return nil
		}
		return err
	}

	// Get the tree data from the commit
	var headContent []byte

	// Parse the tree hash to find the file
	// Since our implementation stores the tree as JSON of staged files at commit time,
	// we need to load the tree and find the file hash
	treePath := filepath.Join(repo.GitDir, OBJECTS_DIR, commit.TreeHash)
	treeData, err := ioutil.ReadFile(treePath)
	if err != nil {
		// Tree might not exist in simple implementation, try to find file directly
		headContent = []byte("") // Empty content for new files
	} else {
		// Parse tree data
		var files []IndexEntry
		if err := json.Unmarshal(treeData, &files); err == nil {
			// Find the file in the tree
			for _, file := range files {
				if file.FilePath == filePath {
					// Load the file content from objects
					objectPath := filepath.Join(repo.GitDir, OBJECTS_DIR, file.Hash)
					content, err := ioutil.ReadFile(objectPath)
					if err == nil {
						headContent = content
					}
					break
				}
			}
		}
	}

	// Generate diff
	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(headContent)),
		B:        difflib.SplitLines(string(currentContent)),
		FromFile: fmt.Sprintf("a/%s", filePath),
		ToFile:   fmt.Sprintf("b/%s", filePath),
		Context:  2,
	}

	result, err := difflib.GetUnifiedDiffString(diff)
	if err != nil {
		return err
	}

	if result != "" {
		fmt.Print(result)
	}

	return nil
}

// ShowLog displays the commit history
func ShowLog() error {
	repo, err := FindGitterRepo()
	if err != nil {
		return err
	}

	// Get current HEAD
	head, err := GetCurrentHead()
	if err != nil {
		return err
	}

	if head == "" {
		fmt.Println("No commits yet")
		return nil
	}

	// Traverse commit history
	currentCommit := head
	for currentCommit != "" {
		// Load commit
		commitPath := filepath.Join(repo.GitDir, OBJECTS_DIR, currentCommit)
		commitData, err := ioutil.ReadFile(commitPath)
		if err != nil {
			return err
		}

		var commit Commit
		if err := json.Unmarshal(commitData, &commit); err != nil {
			return err
		}

		// Print commit info
		fmt.Printf("commit %s\n", commit.Hash)
		fmt.Printf("Author: %s\n", commit.Author)
		fmt.Printf("Date: %s\n", commit.Date.Format("Mon Jan 2 15:04:05 2006 -0700"))
		fmt.Printf("\n    %s\n\n", commit.Message)

		// Move to parent
		currentCommit = commit.Parent
	}

	return nil
}
