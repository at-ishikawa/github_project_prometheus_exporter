package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/at-ishikawa/github-project-prometheus-exporter/internal/github"
	"github.com/spf13/cobra"
)

const (
	exitCodeOK int = 0
)

func main() {
	exitCode, err := runMain()
	if err != nil {
		log.Fatalln(err)
	}
	os.Exit(exitCode)
}

func runMain() (int, error) {
	exitCode := 1

	rootCommand := cobra.Command{
		Use:  "exporter [userId] [projectNumber]",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// TODO: Support both of env vars and arguments
			userId := args[0]
			projectNumber, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("projectNumber must be integer, but %s", args[1])
			}

			githubToken := os.Getenv("GITHUB_TOKEN")
			if githubToken == "" {
				return fmt.Errorf("GITHUB_TOKEN environment variable is required")
			}

			client, err := github.NewClient(githubToken)
			if err != nil {
				return fmt.Errorf("github.NewClient: %w", err)
			}

			projectID, err := client.FetchUserProject(ctx, userId, projectNumber)
			if err != nil {
				return fmt.Errorf("FetchUserProject: %w", err)
			}

			fmt.Println(projectID)
			return nil
		},
	}

	if err := rootCommand.Execute(); err != nil {
		return exitCode, err
	}
	return exitCodeOK, nil
}
