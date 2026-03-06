package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/pixincreate/gito/internal/config"
	"github.com/pixincreate/gito/internal/github"
	"github.com/spf13/cobra"
)

var (
	verbose bool
	backend string
)

var rootCmd = &cobra.Command{
	Use:   "gito",
	Short: "GitHub automation CLI tool",
	Long: `Gito is a command-line tool for interacting with GitHub PRs and review comments.
Supports both gh CLI and curl (GitHub REST API) backends.

Backend selection:
  - Uses 'gh' CLI if installed and authenticated
  - Falls back to 'curl' with GITHUB_PAT environment variable
  - Override with --backend=gh or --backend=curl`,
	Version: "dev",
}

func SetVersion(v string) {
	rootCmd.Version = v
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVar(&backend, "backend", "auto", "backend to use: gh, curl, or auto")
}

// resolveClient determines which backend to use and returns a ready-to-use Client.
//
// Flow:
//  1. If --backend=gh, use gh CLI (error if not available)
//  2. If --backend=curl, use curl with GITHUB_PAT (error if not set)
//  3. If --backend=auto (default):
//     a. Try gh CLI — if available and authenticated, use it
//     b. Fall back to curl — print yellow warning, require GITHUB_PAT
func resolveClient() (github.Client, error) {
	switch backend {
	case "gh":
		if !isGHAvailable() {
			return nil, fmt.Errorf("gh CLI is not installed or not authenticated. Install from https://cli.github.com/")
		}
		return &github.GHClient{}, nil

	case "curl":
		return newCurlClient()

	default: // "auto"
		if isGHAvailable() {
			return &github.GHClient{}, nil
		}
		warnYellow("gh CLI not found or not authenticated, falling back to curl")
		return newCurlClient()
	}
}

func isGHAvailable() bool {
	cmd := exec.Command("gh", "auth", "status")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

func newCurlClient() (*github.CurlClient, error) {
	token := os.Getenv("GITHUB_PAT")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_PAT environment variable is not set.\n\nSet it with:\n  export GITHUB_PAT=ghp_your_token_here\n\nOr install gh CLI:\n  https://cli.github.com/")
	}
	return &github.CurlClient{Token: token}, nil
}

func detectRepo() (string, error) {
	return config.DetectRepo()
}

// warnYellow prints a warning message in yellow to stderr.
func warnYellow(msg string) {
	fmt.Fprintf(os.Stderr, "\033[33mWarning: %s\033[0m\n", msg)
}
