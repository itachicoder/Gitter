// internal/operations_test.go
package internal

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// captureOutput captures stdout during test execution
func captureOutput(t *testing.T, f func()) string {
	originalStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w

	// Execute function
	f()

	// Restore stdout
	w.Close()
	os.Stdout = originalStdout

	// Read captured output
	output, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	return string(output)
}

func TestShowStatus(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(string) error
		wantOutput []string
	}{
		{
			name:       "Clean working tree",
			setup:      func(dir string) error { return nil },
			wantOutput: []string{"nothing to commit, working tree clean"},
		},
		{
			name: "Untracked files",
			setup: func(dir string) error {
				return ioutil.WriteFile("test.txt", []byte("content"), 0644)
			},
			wantOutput: []string{"Untracked files:", "test.txt"},
		},
		{
			name: "Staged files",
			setup: func(dir string) error {
				if err := ioutil.WriteFile("test.txt", []byte("content"), 0644); err != nil {
					return err
				}
				return AddFile("test.txt")
			},
			wantOutput: []string{"Changes to be committed:", "modified: test.txt"},
		},
		{
			name: "Mixed state",
			setup: func(dir string) error {
				// Create and stage one file
				if err := ioutil.WriteFile("staged.txt", []byte("staged"), 0644); err != nil {
					return err
				}
				if err := AddFile("staged.txt"); err != nil {
					return err
				}

				// Create untracked file
				if err := ioutil.WriteFile("untracked.txt", []byte("untracked"), 0644); err != nil {
					return err
				}

				// Create and modify already tracked file
				if err := ioutil.WriteFile("tracked.txt", []byte("original"), 0644); err != nil {
					return err
				}
				if err := AddFile("tracked.txt"); err != nil {
					return err
				}
				// Reset index to simulate committed state
				index, err := LoadIndex()
				if err != nil {
					return err
				}
				for i := range index {
					if index[i].FilePath == "tracked.txt" {
						index[i].Modified = false
					}
				}
				if err := SaveIndex(index); err != nil {
					return err
				}
				// Modify file
				return ioutil.WriteFile("tracked.txt", []byte("modified"), 0644)
			},
			wantOutput: []string{
				"Changes to be committed:",
				"modified: staged.txt",
				"Changes not staged for commit:",
				"modified: tracked.txt",
				"Untracked files:",
				"untracked.txt",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, cleanup := setupTestRepo(t)
			defer cleanup()

			// Initialize repository
			err := InitRepository()
			if err != nil {
				t.Fatalf("Failed to initialize repository: %v", err)
			}

			// Setup test conditions
			if err := tt.setup(tempDir); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			// Capture output
			output := captureOutput(t, func() {
				err := ShowStatus()
				if err != nil {
					t.Errorf("ShowStatus() error = %v", err)
				}
			})

			// Verify output contains expected strings
			for _, expected := range tt.wantOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("ShowStatus() output missing expected string: %s\nGot: %s", expected, output)
				}
			}
		})
	}
}

func TestCommitChanges(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(string) error
		message   string
		all       bool
		wantErr   bool
		errString string
	}{
		{
			name:      "Nothing to commit",
			setup:     func(dir string) error { return nil },
			message:   "Test commit",
			wantErr:   true,
			errString: "nothing to commit",
		},
		{
			name: "Simple commit",
			setup: func(dir string) error {
				if err := ioutil.WriteFile("test.txt", []byte("content"), 0644); err != nil {
					return err
				}
				return AddFile("test.txt")
			},
			message: "First commit",
			wantErr: false,
		},
		{
			name: "Commit with -a flag",
			setup: func(dir string) error {
				// Create and commit initial file
				if err := ioutil.WriteFile("test.txt", []byte("original"), 0644); err != nil {
					return err
				}
				if err := AddFile("test.txt"); err != nil {
					return err
				}
				if err := CommitChanges("Initial commit", false); err != nil {
					return err
				}
				// Modify the file without staging
				return ioutil.WriteFile("test.txt", []byte("modified"), 0644)
			},
			message: "Update test file",
			all:     true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, cleanup := setupTestRepo(t)
			defer cleanup()

			// Initialize repository
			err := InitRepository()
			if err != nil {
				t.Fatalf("Failed to initialize repository: %v", err)
			}

			// Setup test conditions
			if err := tt.setup(tempDir); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			// Run commit
			output := captureOutput(t, func() {
				err := CommitChanges(tt.message, tt.all)
				if (err != nil) != tt.wantErr {
					t.Errorf("CommitChanges() error = %v, wantErr %v", err, tt.wantErr)
				}
				if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errString) {
					t.Errorf("CommitChanges() error = %v, want error containing %v", err, tt.errString)
				}
			})

			if !tt.wantErr {
				// Verify commit was created
				if !strings.Contains(output, tt.message) {
					t.Errorf("CommitChanges() output missing commit message: %s", tt.message)
				}

				// Verify HEAD was updated
				head, err := GetCurrentHead()
				if err != nil {
					t.Errorf("Failed to get HEAD: %v", err)
				}
				if head == "" {
					t.Error("HEAD not updated after commit")
				}

				// Verify commit object exists
				commitPath := filepath.Join(GITTER_DIR, OBJECTS_DIR, head)
				if _, err := os.Stat(commitPath); os.IsNotExist(err) {
					t.Errorf("Commit object not found: %s", commitPath)
				}

				// Verify index is clean
				index, err := LoadIndex()
				if err != nil {
					t.Errorf("Failed to load index: %v", err)
				}
				for _, entry := range index {
					if entry.Modified {
						t.Errorf("Index entry %s still marked as modified", entry.FilePath)
					}
				}
			}
		})
	}
}

func TestShowLog(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(string) error
		wantOutput []string
	}{
		{
			name:       "No commits",
			setup:      func(dir string) error { return nil },
			wantOutput: []string{"No commits yet"},
		},
		{
			name: "Single commit",
			setup: func(dir string) error {
				if err := ioutil.WriteFile("test.txt", []byte("content"), 0644); err != nil {
					return err
				}
				if err := AddFile("test.txt"); err != nil {
					return err
				}
				return CommitChanges("First commit", false)
			},
			wantOutput: []string{"commit", "Author: user", "First commit"},
		},
		{
			name: "Multiple commits",
			setup: func(dir string) error {
				// First commit
				if err := ioutil.WriteFile("file1.txt", []byte("content1"), 0644); err != nil {
					return err
				}
				if err := AddFile("file1.txt"); err != nil {
					return err
				}
				if err := CommitChanges("First commit", false); err != nil {
					return err
				}

				// Second commit
				if err := ioutil.WriteFile("file2.txt", []byte("content2"), 0644); err != nil {
					return err
				}
				if err := AddFile("file2.txt"); err != nil {
					return err
				}
				return CommitChanges("Second commit", false)
			},
			wantOutput: []string{
				"Second commit", // Most recent first
				"First commit",
				"Author: user",
				"Date:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, cleanup := setupTestRepo(t)
			defer cleanup()

			// Initialize repository
			err := InitRepository()
			if err != nil {
				t.Fatalf("Failed to initialize repository: %v", err)
			}

			// Setup test conditions
			if err := tt.setup(tempDir); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			// Capture output
			output := captureOutput(t, func() {
				err := ShowLog()
				if err != nil {
					t.Errorf("ShowLog() error = %v", err)
				}
			})

			// Verify output contains expected strings
			for _, expected := range tt.wantOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("ShowLog() output missing expected string: %s\nGot: %s", expected, output)
				}
			}
		})
	}
}

func TestShowDiff(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(string) error
		diffPath     string
		wantContains []string // Changed to be more flexible
		wantErr      bool
	}{
		{
			name:         "No commits",
			setup:        func(dir string) error { return nil },
			diffPath:     "",
			wantErr:      true,
			wantContains: []string{"no commits yet"},
		},
		{
			name: "Modified file",
			setup: func(dir string) error {
				// Create and commit initial file
				if err := ioutil.WriteFile("test.txt", []byte("original content"), 0644); err != nil {
					return err
				}
				if err := AddFile("test.txt"); err != nil {
					return err
				}
				if err := CommitChanges("Initial commit", false); err != nil {
					return err
				}
				// Modify the file
				return ioutil.WriteFile("test.txt", []byte("modified content"), 0644)
			},
			diffPath: "",
			wantErr:  false,
			wantContains: []string{
				"test.txt",          // File name should appear
				"@@",                // Diff header
				"+modified content", // New content
				"-original content", // Removed content
			},
		},
		{
			name: "New file added",
			setup: func(dir string) error {
				// Create and commit initial file
				if err := ioutil.WriteFile("file1.txt", []byte("content1"), 0644); err != nil {
					return err
				}
				if err := AddFile("file1.txt"); err != nil {
					return err
				}
				if err := CommitChanges("Initial commit", false); err != nil {
					return err
				}
				// Add a new file (untracked)
				return ioutil.WriteFile("newfile.txt", []byte("new content"), 0644)
			},
			diffPath: "",
			wantErr:  false,
			wantContains: []string{
				"newfile.txt",
				"+new content",
			},
		},
		{
			name: "Specific file diff",
			setup: func(dir string) error {
				// Create multiple files
				if err := ioutil.WriteFile("file1.txt", []byte("content1"), 0644); err != nil {
					return err
				}
				if err := ioutil.WriteFile("file2.txt", []byte("content2"), 0644); err != nil {
					return err
				}
				if err := AddFile("file1.txt"); err != nil {
					return err
				}
				if err := AddFile("file2.txt"); err != nil {
					return err
				}
				if err := CommitChanges("Initial commit", false); err != nil {
					return err
				}
				// Modify both files
				if err := ioutil.WriteFile("file1.txt", []byte("modified1"), 0644); err != nil {
					return err
				}
				return ioutil.WriteFile("file2.txt", []byte("modified2"), 0644)
			},
			diffPath: "file1.txt",
			wantErr:  false,
			wantContains: []string{
				"file1.txt",
				"+modified1",
				"-content1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, cleanup := setupTestRepo(t)
			defer cleanup()

			// Initialize repository
			err := InitRepository()
			if err != nil {
				t.Fatalf("Failed to initialize repository: %v", err)
			}

			// Setup test conditions
			if err := tt.setup(tempDir); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			// Run diff and capture error or output
			var output string
			if tt.wantErr {
				err := ShowDiff(tt.diffPath)
				if err == nil {
					t.Errorf("ShowDiff() error = nil, wantErr %v", tt.wantErr)
				}
				output = err.Error()
			} else {
				output = captureOutput(t, func() {
					err := ShowDiff(tt.diffPath)
					if err != nil {
						t.Errorf("ShowDiff() error = %v", err)
					}
				})
			}

			// Verify output contains expected strings
			for _, expected := range tt.wantContains {
				if !strings.Contains(output, expected) {
					t.Errorf("ShowDiff() output missing expected string: %s\nFull output:\n%s", expected, output)
				}
			}

			// Debug output (remove in production)
			if !tt.wantErr && testing.Verbose() {
				t.Logf("Diff output:\n%s", output)
			}
		})
	}
}

// Benchmarks
func BenchmarkAddFile(b *testing.B) {
	_, cleanup := setupTestRepo(nil)
	defer cleanup()

	// Initialize repository
	err := InitRepository()
	if err != nil {
		b.Fatalf("Failed to initialize repository: %v", err)
	}

	// Create test file
	err = ioutil.WriteFile("bench.txt", bytes.Repeat([]byte("test data\n"), 1000), 0644)
	if err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := AddFile("bench.txt")
		if err != nil {
			b.Errorf("AddFile() error = %v", err)
		}
	}
}

func BenchmarkCommitChanges(b *testing.B) {
	_, cleanup := setupTestRepo(nil)
	defer cleanup()

	// Initialize repository
	err := InitRepository()
	if err != nil {
		b.Fatalf("Failed to initialize repository: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create a new file for each iteration
		filename := fmt.Sprintf("bench_%d.txt", i)
		err := ioutil.WriteFile(filename, []byte("test content"), 0644)
		if err != nil {
			b.Errorf("Failed to create file: %v", err)
		}

		err = AddFile(filename)
		if err != nil {
			b.Errorf("AddFile() error = %v", err)
		}

		err = CommitChanges(fmt.Sprintf("Commit %d", i), false)
		if err != nil {
			b.Errorf("CommitChanges() error = %v", err)
		}
	}
}
