package cmd

import (
	"fmt"
	"strings"

	"github.com/postacksol/flux-relay-cli/internal/api"
	"github.com/postacksol/flux-relay-cli/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var prCmd = &cobra.Command{
	Use:   "pr [project-name-or-id]",
	Short: "Select a project (enter project context)",
	Long: `Select a project to work with. You can use either the project name or ID.
After selecting a project, you can use commands like 'flux-relay server list'.

Examples:
  flux-relay pr MyProject        # Select by name
  flux-relay pr 56OSXXQH        # Select by ID
  flux-relay pr                  # Show current project`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPr,
}

func init() {
	rootCmd.AddCommand(prCmd)
}

func runPr(cmd *cobra.Command, args []string) error {
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
			fmt.Println("  flux-relay projects list")
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

	projectIdentifier := args[0]

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
		return fmt.Errorf("project '%s' not found. Use 'flux-relay projects list' to see available projects", projectIdentifier)
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
