package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/postacksol/flux-relay-cli/internal/api"
	"github.com/postacksol/flux-relay-cli/internal/config"
	"github.com/spf13/cobra"
)

var sqlCmd = &cobra.Command{
	Use:   "sql <query>",
	Short: "Execute SQL query",
	Long: `Execute SQL query on the selected server and nameserver.

The query will automatically filter by server_id for data isolation.
If a nameserver is selected, you can query nameserver-specific tables.

Examples:
  flux-relay sql "SELECT * FROM conversations_db WHERE server_id = ? LIMIT 10"
  flux-relay sql "SELECT COUNT(*) FROM end_users_db WHERE server_id = ?"
  flux-relay sql "INSERT INTO conversations_db (server_id, ...) VALUES (?, ...)"`,
	Args: cobra.MinimumNArgs(1),
	RunE: runSql,
}

func init() {
	rootCmd.AddCommand(sqlCmd)
}

func runSql(cmd *cobra.Command, args []string) error {
	// Get API URL
	apiURL := getAPIURL()

	// Get access token
	cfg := config.New()
	accessToken := cfg.GetAccessToken()
	if accessToken == "" {
		return fmt.Errorf("not logged in. Run 'flux-relay login' first")
	}

	// Get selected project and server
	projectID := cfg.GetSelectedProject()
	if projectID == "" {
		return fmt.Errorf("no project selected. Use 'flux-relay pr <project-name-or-id>' to select a project")
	}

	serverID := cfg.GetSelectedServer()
	if serverID == "" {
		return fmt.Errorf("no server selected. Use 'flux-relay server <server-name-or-id>' to select a server")
	}

	// Join all args to handle queries with spaces
	query := strings.Join(args, " ")

	// Get selected nameserver (optional - for context)
	nameserverID := cfg.GetSelectedNameserver()

	// Create API client and execute query
	client := api.NewClient(apiURL)
	
	// Prepare query args - server_id will be automatically added by the API
	queryArgs := []interface{}{}
	
	// If nameserver is selected, we might want to use it in the query
	// But the API handles server_id automatically, so we just pass the query as-is
	queryResponse, err := client.ExecuteQuery(accessToken, projectID, serverID, query, queryArgs)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return fmt.Errorf("query failed: %s", apiErr.Error())
		}
		return fmt.Errorf("failed to execute query: %w", err)
	}

	if !queryResponse.Success {
		if queryResponse.ErrorMessage != "" {
			return fmt.Errorf("query error: %s", queryResponse.ErrorMessage)
		}
		return fmt.Errorf("query failed")
	}

	// Display results
	if len(queryResponse.Columns) > 0 {
		// SELECT query - display results in table
		fmt.Printf("Query executed successfully (%dms)\n\n", queryResponse.ExecutionTime)
		
		if len(queryResponse.Rows) == 0 {
			fmt.Println("No rows returned.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		
		// Print header
		fmt.Fprintln(w, strings.Join(queryResponse.Columns, "\t"))
		
		// Print separator
		separator := make([]string, len(queryResponse.Columns))
		for i := range separator {
			separator[i] = "──"
		}
		fmt.Fprintln(w, strings.Join(separator, "\t"))
		
		// Print rows
		for _, row := range queryResponse.Rows {
			rowStr := make([]string, len(row))
			for i, val := range row {
				if val == nil {
					rowStr[i] = "NULL"
				} else {
					// Convert to string, handling JSON encoding for complex types
					if str, ok := val.(string); ok {
						rowStr[i] = str
					} else {
						jsonBytes, _ := json.Marshal(val)
						rowStr[i] = string(jsonBytes)
					}
				}
			}
			fmt.Fprintln(w, strings.Join(rowStr, "\t"))
		}
		
		w.Flush()
		fmt.Println()
		fmt.Printf("Rows returned: %d\n", len(queryResponse.Rows))
	} else {
		// INSERT/UPDATE/DELETE query
		fmt.Printf("Query executed successfully (%dms)\n", queryResponse.ExecutionTime)
		fmt.Printf("Rows affected: %d\n", queryResponse.RowsAffected)
	}

	if nameserverID != "" {
		fmt.Println()
		fmt.Println("Note: Using selected nameserver context")
	}

	return nil
}
