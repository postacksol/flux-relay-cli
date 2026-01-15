package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/postacksol/flux-relay-cli/internal/api"
	"github.com/postacksol/flux-relay-cli/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var prCmd = &cobra.Command{
	Use:   "pr [project-name-or-id]",
	Short: "Manage projects",
	Long: `List and select projects to work with.

Examples:
  flux-relay pr list              # List all projects
  flux-relay pr MyProject         # Select by name
  flux-relay pr 56OSXXQH          # Select by ID
  flux-relay pr                   # Show current project`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPrShowOrSelect,
}

var prListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	Long:  "List all projects in your account",
	RunE:  runPrList,
}

func init() {
	prCmd.AddCommand(prListCmd)
	rootCmd.AddCommand(prCmd)
}

func runPrShowOrSelect(cmd *cobra.Command, args []string) error {
	// Get API URL
	apiURL := apiBaseURL
	if apiURL == "" {
		apiURL = viper.GetString("api_url")
		if apiURL == "" {
			apiURL = "http://localhost:3000"
		}
	}

	// Get access token
	cfg := config.New()
	accessToken := cfg.GetAccessToken()
	if accessToken == "" {
		return fmt.Errorf("not logged in. Run 'flux-relay login' first")
	}

	// If no argument, show current project
	if len(args) == 0 {
		selectedProjectID := cfg.GetSelectedProject()
		if selectedProjectID == "" {
			fmt.Println("No project selected.")
			fmt.Println()
			fmt.Println("Select a project using:")
			fmt.Println("  flux-relay pr <project-name-or-id>")
			fmt.Println()
			fmt.Println("Or list available projects:")
			fmt.Println("  flux-relay pr list")
			return nil
		}

		// Get project details
		client := api.NewClient(apiURL)
		projectsResponse, err := client.ListProjects(accessToken)
		if err != nil {
			return fmt.Errorf("failed to get project info: %w", err)
		}

		// Find the selected project
		var selectedProject *api.Project
		for i := range projectsResponse.Projects {
			if projectsResponse.Projects[i].ID == selectedProjectID {
				selectedProject = &projectsResponse.Projects[i]
				break
			}
		}

		if selectedProject == nil {
			fmt.Printf("⚠️  Selected project (ID: %s) not found.\n", selectedProjectID)
			fmt.Println("Please select a different project.")
			return nil
		}

		fmt.Printf("Current project: %s (%s)\n", selectedProject.Name, selectedProject.ID)
		if selectedProject.Description != "" {
			fmt.Printf("Description: %s\n", selectedProject.Description)
		}
		return nil
	}

	// If argument provided, treat as project selection
	// Join all args to handle names with spaces (e.g., "My Project Name")
	projectIdentifier := strings.Join(args, " ")

	// Get all projects
	client := api.NewClient(apiURL)
	projectsResponse, err := client.ListProjects(accessToken)
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	// Find project by ID or name (case-insensitive)
	var selectedProject *api.Project
	for i := range projectsResponse.Projects {
		project := &projectsResponse.Projects[i]
		if project.ID == projectIdentifier || 
		   strings.EqualFold(project.Name, projectIdentifier) {
			selectedProject = project
			break
		}
	}

	if selectedProject == nil {
		return fmt.Errorf("project '%s' not found. Use 'flux-relay pr list' to see available projects", projectIdentifier)
	}

	// Save selected project
	if err := cfg.SetSelectedProject(selectedProject.ID); err != nil {
		return fmt.Errorf("failed to save project selection: %w", err)
	}

	fmt.Printf("✅ Selected project: %s (%s)\n", selectedProject.Name, selectedProject.ID)
	if selectedProject.Description != "" {
		fmt.Printf("   Description: %s\n", selectedProject.Description)
	}
	fmt.Println()
	fmt.Println("You can now use:")
	fmt.Println("  flux-relay server list")

	return nil
}

func runPrList(cmd *cobra.Command, args []string) error {
	// Get API URL
	apiURL := apiBaseURL
	if apiURL == "" {
		apiURL = viper.GetString("api_url")
		if apiURL == "" {
			apiURL = "http://localhost:3000"
		}
	}

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
