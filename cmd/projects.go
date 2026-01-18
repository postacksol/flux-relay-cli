package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/postacksol/flux-relay-cli/internal/api"
	"github.com/postacksol/flux-relay-cli/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Manage projects",
	Long:  "List and manage your Flux Relay projects",
}

var projectsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	Long:  "List all projects in your account",
	RunE:  runProjectsList,
}

func init() {
	projectsCmd.AddCommand(projectsListCmd)
	rootCmd.AddCommand(projectsCmd)
}

func runProjectsList(cmd *cobra.Command, args []string) error {
	// Get API URL
	apiURL := getAPIURL()

	// Get access token
	cfg := config.New()
	accessToken := cfg.GetAccessToken()
	if accessToken == "" {
		return fmt.Errorf("not logged in. Run 'flux-relay login' first")
	}

	// Create API client and list projects
	client := api.NewClient(apiURL)
	projectsResponse, err := client.ListProjects(accessToken)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			if apiErr.Code() == "Unauthorized" || apiErr.Code() == "unauthorized" {
				return fmt.Errorf("authentication failed. Please run 'flux-relay login' again")
			}
			return fmt.Errorf("API error: %w", apiErr)
		}
		return fmt.Errorf("failed to list projects: %w", err)
	}

	projects := projectsResponse.Projects

	if len(projects) == 0 {
		fmt.Println("No projects found.")
		fmt.Println()
		fmt.Println("Create a project using the web dashboard or API.")
		return nil
	}

	// Display projects in a table
	fmt.Printf("Found %d project(s):\n\n", len(projects))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tDESCRIPTION\tCREATED\tSTATUS")
	fmt.Fprintln(w, "──\t────\t───────────\t───────\t──────")

	for _, project := range projects {
		// Format created date
		createdAt, err := time.Parse(time.RFC3339, project.CreatedAt)
		createdStr := project.CreatedAt
		if err == nil {
			createdStr = createdAt.Format("2006-01-02")
		}

		// Truncate description if too long
		description := project.Description
		if len(description) > 40 {
			description = description[:37] + "..."
		}
		if description == "" {
			description = "-"
		}

		// Status
		status := "Active"
		if !project.IsActive {
			status = "Inactive"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			project.ID,
			project.Name,
			description,
			createdStr,
			status,
		)
	}

	w.Flush()
	fmt.Println()

	return nil
}
