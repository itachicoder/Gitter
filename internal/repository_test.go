// internal/repository_test.go
package internal

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// setupTestRepo creates a temporary directory for testing
func setupTestRepo(t *testing.T) (string, func()) {
	tempDir, err := ioutil.TempDir("", "gitter-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current dir: %v", err)
	}

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		os.Chdir(originalDir)
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

func TestInitRepository(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(string) error
		wantErr bool
	}{
		{
			name:    "Initialize new repository",
			setup:   func(dir string) error { return nil },
			wantErr: false,
		},
		{
			name: "Initialize already initialized repository",
			setup: func(dir string) error {
				return os.MkdirAll(filepath.Join(dir, GITTER_DIR), 0755)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, cleanup := setupTestRepo(t)
			defer cleanup()

			// Setup test conditions
			if err := tt.setup(tempDir); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			// Run the test
			err := InitRepository()
			if (err != nil) != tt.wantErr {
				t.Errorf("InitRepository() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify repository structure was created
				expectedDirs := []string{
					GITTER_DIR,
					filepath.Join(GITTER_DIR, REFS_DIR),
					filepath.Join(GITTER_DIR, REFS_DIR, HEADS_DIR),
					filepath.Join(GITTER_DIR, OBJECTS_DIR),
				}

				for _, dir := range expectedDirs {
					if _, err := os.Stat(dir); os.IsNotExist(err) {
						t.Errorf("Expected directory %s was not created", dir)
					}
				}

				// Verify files were created
				expectedFiles := []string{
					filepath.Join(GITTER_DIR, HEAD_FILE),
					filepath.Join(GITTER_DIR, INDEX_FILE),
					filepath.Join(GITTER_DIR, LOG_FILE),
				}

				for _, file := range expectedFiles {
					if _, err := os.Stat(file); os.IsNotExist(err) {
						t.Errorf("Expected file %s was not created", file)
					}
				}

				// Verify HEAD content
				headContent, err := ioutil.ReadFile(filepath.Join(GITTER_DIR, HEAD_FILE))
				if err != nil {
					t.Errorf("Failed to read HEAD file: %v", err)
				}
				expectedHead := "ref: refs/heads/main\n"
				if string(headContent) != expectedHead {
					t.Errorf("HEAD content = %v, want %v", string(headContent), expectedHead)
				}
			}
		})
	}
}

func TestFindGitterRepo(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(string) error
		wantErr bool
	}{
		{
			name: "Find repository in current directory",
			setup: func(dir string) error {
				return os.MkdirAll(filepath.Join(dir, GITTER_DIR), 0755)
			},
			wantErr: false,
		},
		{
			name: "Find repository in parent directory",
			setup: func(dir string) error {
				// Create .gitter in parent
				if err := os.MkdirAll(filepath.Join(dir, GITTER_DIR), 0755); err != nil {
					return err
				}
				// Create subdirectory and change to it
				subDir := filepath.Join(dir, "subdir")
				if err := os.MkdirAll(subDir, 0755); err != nil {
					return err
				}
				return os.Chdir(subDir)
			},
			wantErr: false,
		},
		{
			name: "No repository found",
			setup: func(dir string) error {
				return nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, cleanup := setupTestRepo(t)
			defer cleanup()

			// Setup test conditions
			if err := tt.setup(tempDir); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			// Run the test
			repo, err := FindGitterRepo()
			if (err != nil) != tt.wantErr {
				t.Errorf("FindGitterRepo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if repo == nil {
					t.Error("FindGitterRepo() returned nil repository")
					return
				}

				// Verify repository paths
				expectedGitDir := filepath.Join(repo.WorkingDir, GITTER_DIR)
				if repo.GitDir != expectedGitDir {
					t.Errorf("Repository GitDir = %v, want %v", repo.GitDir, expectedGitDir)
				}
			}
		})
	}
}

func TestLoadAndSaveIndex(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	// Initialize repository
	err := InitRepository()
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	// Test loading empty index
	index, err := LoadIndex()
	if err != nil {
		t.Errorf("LoadIndex() error = %v", err)
		return
	}

	if len(index) != 0 {
		t.Errorf("LoadIndex() returned non-empty index, got %d entries", len(index))
	}

	// Test saving index
	testIndex := []IndexEntry{
		{
			FilePath: "test.txt",
			Hash:     "abc123",
			Modified: true,
		},
		{
			FilePath: "src/main.go",
			Hash:     "def456",
			Modified: false,
		},
	}

	err = SaveIndex(testIndex)
	if err != nil {
		t.Errorf("SaveIndex() error = %v", err)
		return
	}

	// Test loading saved index
	loadedIndex, err := LoadIndex()
	if err != nil {
		t.Errorf("LoadIndex() after save error = %v", err)
		return
	}

	if len(loadedIndex) != len(testIndex) {
		t.Errorf("LoadIndex() returned %d entries, want %d", len(loadedIndex), len(testIndex))
		return
	}

	// Verify index contents
	for i, entry := range loadedIndex {
		if entry.FilePath != testIndex[i].FilePath {
			t.Errorf("Index entry %d FilePath = %v, want %v", i, entry.FilePath, testIndex[i].FilePath)
		}
		if entry.Hash != testIndex[i].Hash {
			t.Errorf("Index entry %d Hash = %v, want %v", i, entry.Hash, testIndex[i].Hash)
		}
		if entry.Modified != testIndex[i].Modified {
			t.Errorf("Index entry %d Modified = %v, want %v", i, entry.Modified, testIndex[i].Modified)
		}
	}
}

func TestAddFile(t *testing.T) {
	tests := []struct {
		name    string
		files   map[string]string // filename -> content
		addFile string
		wantErr bool
	}{
		{
			name: "Add single file",
			files: map[string]string{
				"test.txt": "Hello World",
			},
			addFile: "test.txt",
			wantErr: false,
		},
		{
			name: "Add non-existent file",
			files: map[string]string{
				"test.txt": "Hello World",
			},
			addFile: "nonexistent.txt",
			wantErr: false, // Should not error, just skip the file
		},
		{
			name: "Add with wildcard",
			files: map[string]string{
				"test1.go": "package main",
				"test2.go": "package main",
				"test.txt": "not go file",
			},
			addFile: "*.go",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, cleanup := setupTestRepo(t)
			defer cleanup()

			// Initialize repository
			err := InitRepository()
			if err != nil {
				t.Fatalf("Failed to initialize repository: %v", err)
			}

			// Create test files
			for filename, content := range tt.files {
				err := ioutil.WriteFile(filename, []byte(content), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file %s: %v", filename, err)
				}
			}

			// Run AddFile
			err = AddFile(tt.addFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify index was updated
				index, err := LoadIndex()
				if err != nil {
					t.Errorf("Failed to load index: %v", err)
					return
				}

				// Count expected files
				expectedFiles := 0
				if tt.addFile == "*.go" {
					expectedFiles = 2 // test1.go and test2.go
				} else if tt.addFile == "test.txt" {
					expectedFiles = 1
				}

				if len(index) != expectedFiles {
					t.Errorf("Index has %d entries, want %d", len(index), expectedFiles)
				}

				// Verify all entries are marked as modified
				for _, entry := range index {
					if !entry.Modified {
						t.Errorf("Index entry %s not marked as modified", entry.FilePath)
					}
				}
			}
		})
	}
}

func TestGetAndUpdateHead(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	// Initialize repository
	err := InitRepository()
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	// Test getting HEAD when no commits exist
	head, err := GetCurrentHead()
	if err != nil {
		t.Errorf("GetCurrentHead() error = %v", err)
		return
	}

	if head != "" {
		t.Errorf("GetCurrentHead() = %v, want empty string", head)
	}

	// Test updating HEAD
	testCommitHash := "abc123456789"
	err = UpdateHead(testCommitHash)
	if err != nil {
		t.Errorf("UpdateHead() error = %v", err)
		return
	}

	// Verify HEAD was updated
	newHead, err := GetCurrentHead()
	if err != nil {
		t.Errorf("GetCurrentHead() after update error = %v", err)
		return
	}

	if newHead != testCommitHash {
		t.Errorf("GetCurrentHead() = %v, want %v", newHead, testCommitHash)
	}
}

func TestCalculateHash(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Empty string",
			input: "",
			want:  "da39a3ee5e6b4b0d3255bfef95601890afd80709",
		},
		{
			name:  "Hello World",
			input: "Hello World",
			want:  "0a4d55a8d778e5022fab701977c5d840bbc486d0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateHash(tt.input)
			if got != tt.want {
				t.Errorf("CalculateHash() = %v, want %v", got, tt.want)
			}
		})
	}
}
