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
   concatenated as separate paragraphs.`)
			default:
				fmt.Printf("No detailed help available for '%s'\n", args[0])
			}
		}
	},
}
