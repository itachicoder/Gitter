package main

import (
	"fmt"
	"os"

	"gitter/internal"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "gitter",
		Short: "Gitter - A git-like version control system",
		Long:  "Gitter is a simple version control system that mimics basic git functionalities",
	}

	// Add commands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(commitCmd)
	rootCmd.AddCommand(diffCmd)
	rootCmd.AddCommand(logCmd)
	rootCmd.AddCommand(helpCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Initialize command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create an empty Gitter repository",
	Run: func(cmd *cobra.Command, args []string) {
		err := internal.InitRepository()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("Initialized empty Git repository in %s/.gitter/\n", internal.GetCurrentDir())
	},
}

// Add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add file contents to the index",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		for _, file := range args {
			err := internal.AddFile(file)
			if err != nil {
				fmt.Printf("Error adding %s: %v\n", file, err)
			}
		}
	},
}

// Status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the working tree status",
	Run: func(cmd *cobra.Command, args []string) {
		err := internal.ShowStatus()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	},
}

// Commit command
var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Record changes to the repository",
	Run: func(cmd *cobra.Command, args []string) {
		message, _ := cmd.Flags().GetString("message")
		all, _ := cmd.Flags().GetBool("all")

		err := internal.CommitChanges(message, all)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	},
}

func init() {
	commitCmd.Flags().StringP("message", "m", "", "Commit message")
	commitCmd.Flags().BoolP("all", "a", false, "Stage all modified files")
	commitCmd.MarkFlagRequired("message")
}

// Diff command
var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Show changes between commits, commit and working tree, etc",
	Run: func(cmd *cobra.Command, args []string) {
		var path string
		if len(args) > 0 {
			path = args[0]
		}
		err := internal.ShowDiff(path)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	},
}

// Log command
var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Show commit logs",
	Run: func(cmd *cobra.Command, args []string) {
		err := internal.ShowLog()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	},
}

// Help command (for detailed help)
var helpCmd = &cobra.Command{
	Use:   "help",
	Short: "Help about any command",
	Long: `Help provides help for any command in the application.
Simply type gitter help [path to command] for full details.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println(`These are common Gitter commands:

   init     Create an empty Gitter repository
   add      Add file contents to the index
   status   Show the working tree status
   commit   Record changes to the repository
   diff     Show changes between commits
   log      Show commit logs`)
		} else {
			// Handle specific command help
			switch args[0] {
			case "init":
				fmt.Println(`NAME:
   init - Create an empty Gitter repository

SYNOPSIS:
   gitter init

DESCRIPTION:
   Creates an empty Gitter repository locally. The default branch should be named 'main'.

OUTPUT:
   Initialized empty Git repository in <current working directory>/.gitter/`)

			case "add":
				fmt.Println(`NAME:
   add - Add file contents to the index

SYNOPSIS:
   gitter add <files>...

DESCRIPTION:
   Adds file contents to the index. This command can be used with individual files,
   patterns, or directories.

EXAMPLES:
   gitter add .                    # Adds all files changed in current working directory
   gitter add file1.txt           # Adds specific file
   gitter add *.py                # Adds all .py files from current working directory

OUTPUT:
   Empty (No direct output, but 'gitter status' will reflect the change.)`)

			case "status":
				fmt.Println(`NAME:
   status - Show the working tree status

SYNOPSIS:
   gitter status

DESCRIPTION:
   List the current state of the working branch. Each section (committed, not staged,
   and untracked) will appear only if the section has some file to show.

OUTPUT:
   Changes to be committed:
     modified: file1.txt
   
   Changes not staged for commit:
     modified: /test/file3.txt
   
   Untracked files:
     modified: /test/file4.txt`)

			case "commit":
				fmt.Println(`NAME:
   commit - Record changes to the repository

SYNOPSIS:
   gitter commit -m [-a] <msg>

DESCRIPTION:
   Create a new commit containing the current contents of the index and the given log message
   describing the changes. The new commit is a direct child of HEAD, usually the tip of the current branch,
   and the branch is updated to point to it.

OPTIONS:
   -a: Tell the command to automatically stage files that have been modified and deleted, but new files
       you have not told Git about are not affected.
   -m: Use the given <msg> as the commit message. If multiple -m options are given, their values are
       concatenated as separate paragraphs.

OUTPUT:
   [main 538bb9d] Your commit message`)

			case "diff":
				fmt.Println(`NAME:
   diff - Show changes between commits, commit and working tree, etc

SYNOPSIS:
   gitter diff [<path>...]

DESCRIPTION:
   Show diff between file content of current head and all unindexed files.
   
   If the given argument is a complete path then show diff only for that file.
   If the given argument is a directory then show diff all unindexed files within the directory.

OUTPUT:
   --- a/<file_path>
   +++ b/<file_path>
   @@ -X,Y +A,B @@
   <two line above the change from head>
   - This line was removed
   + This line was added
   <two line below the change from head>`)

			case "log":
				fmt.Println(`NAME:
   log - Show commit logs

SYNOPSIS:
   gitter log

DESCRIPTION:
   Show commit history of current head.

OUTPUT:
   commit 670a84c7cb01c8c90cf5516b2a919123d70a5a0b
   Author: user
   Date: Sat Jan 25 00:27:00 2025 +0530

       updates documentation and schema definition

   Note: The user is just a dummy name. We do not want to perform user management.`)

			default:
				fmt.Printf("No detailed help available for '%s'\n", args[0])
			}
		}
	},
}
