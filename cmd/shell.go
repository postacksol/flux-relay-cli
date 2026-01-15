package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/postacksol/flux-relay-cli/internal/api"
	"github.com/postacksol/flux-relay-cli/internal/config"
	"github.com/spf13/viper"
)

// runServerShell starts an interactive shell for a server
func runServerShell(serverIdentifier string) error {
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

	// Get selected project
	projectID := cfg.GetSelectedProject()
	if projectID == "" {
		return fmt.Errorf("no project selected. Use 'flux-relay pr <project-name-or-id>' to select a project")
	}

	// Find server by ID or name
	client := api.NewClient(apiURL)
	serversResponse, err := client.ListServers(accessToken, projectID)
	if err != nil {
		return fmt.Errorf("failed to list servers: %w", err)
	}

	var selectedServer *api.Server
	for i := range serversResponse.Servers {
		server := &serversResponse.Servers[i]
		if server.ID == serverIdentifier || strings.EqualFold(server.Name, serverIdentifier) {
			selectedServer = server
			break
		}
	}

	if selectedServer == nil {
		return fmt.Errorf("server '%s' not found", serverIdentifier)
	}

	// Save selected server
	if err := cfg.SetSelectedServer(selectedServer.ID); err != nil {
		return fmt.Errorf("failed to save server selection: %w", err)
	}

	// Start interactive shell
	return startShell(cfg, client, accessToken, projectID, selectedServer.ID, selectedServer.Name, "")
}

// runNameserverShell starts an interactive shell for a nameserver
func runNameserverShell(nameserverIdentifier string) error {
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

	// Get selected project and server
	projectID := cfg.GetSelectedProject()
	if projectID == "" {
		return fmt.Errorf("no project selected. Use 'flux-relay pr <project-name-or-id>' to select a project")
	}

	serverID := cfg.GetSelectedServer()
	if serverID == "" {
		return fmt.Errorf("no server selected. Use 'flux-relay server <server-name-or-id>' to select a server")
	}

	// Find nameserver by ID or name
	client := api.NewClient(apiURL)
	databasesResponse, err := client.ListDatabases(accessToken, projectID, serverID)
	if err != nil {
		return fmt.Errorf("failed to list nameservers: %w", err)
	}

	var selectedNameserver *api.Database
	for i := range databasesResponse.Databases {
		ns := &databasesResponse.Databases[i]
		if ns.ID == nameserverIdentifier || strings.EqualFold(ns.DatabaseName, nameserverIdentifier) {
			selectedNameserver = ns
			break
		}
	}

	if selectedNameserver == nil {
		return fmt.Errorf("nameserver '%s' not found", nameserverIdentifier)
	}

	// Save selected nameserver
	if err := cfg.SetSelectedNameserver(selectedNameserver.ID); err != nil {
		return fmt.Errorf("failed to save nameserver selection: %w", err)
	}

	// Get server name for display
	serversResponse, err := client.ListServers(accessToken, projectID)
	if err != nil {
		return fmt.Errorf("failed to get server info: %w", err)
	}

	var serverName string
	for _, srv := range serversResponse.Servers {
		if srv.ID == serverID {
			serverName = srv.Name
			break
		}
	}

	// Start interactive shell
	return startShell(cfg, client, accessToken, projectID, serverID, serverName, selectedNameserver.DatabaseName)
}

// startShell runs the interactive SQL shell
func startShell(cfg *config.ConfigManager, client *api.Client, accessToken, projectID, serverID, serverName, nameserverName string) error {
	// Print welcome message
	fmt.Printf("Connected to %s", serverName)
	if nameserverName != "" {
		fmt.Printf(" (nameserver: %s)", nameserverName)
	}
	fmt.Println()
	fmt.Println()
	fmt.Println("Welcome to Flux Relay SQL shell!")
	fmt.Println()
	fmt.Println("Type \".quit\" to exit the shell and \".help\" to list all available commands.")
	fmt.Println()
	fmt.Println("Enter SQL queries directly (no 'sql' prefix needed).")
	fmt.Println("End queries with semicolon (;) or press Enter twice to execute.")
	if nameserverName != "" {
		fmt.Println()
		fmt.Printf("Note: Tables use nameserver suffix. Example: conversations_%s\n", nameserverName)
		fmt.Println("Use .tables to see all available tables.")
	}
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	var currentQuery strings.Builder

	for {
		// Show prompt
		if currentQuery.Len() == 0 {
			fmt.Print("→ ")
		} else {
			fmt.Print("  ")
		}

		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())

		// Handle empty lines
		if line == "" {
			if currentQuery.Len() > 0 {
				// Empty line after query - execute it
				query := strings.TrimSpace(currentQuery.String())
				if query != "" {
					executeQuery(client, accessToken, projectID, serverID, query)
				}
				currentQuery.Reset()
			}
			continue
		}

		// Detect and strip "sql" prefix if user accidentally includes it
		lineLower := strings.ToLower(line)
		if strings.HasPrefix(lineLower, "sql ") {
			line = strings.TrimSpace(line[4:]) // Remove "sql " prefix
			fmt.Println("Note: You don't need 'sql' prefix in the shell. Just type the query directly.")
		}

		// Handle special commands (start with .)
		if strings.HasPrefix(line, ".") {
			cmd := strings.ToLower(strings.TrimSpace(line))
			switch {
			case cmd == ".quit" || cmd == ".exit" || cmd == ".q":
				fmt.Println("Goodbye!")
				return nil
			case cmd == ".help" || cmd == ".h":
				printHelp()
			case cmd == ".clear" || cmd == ".c":
				currentQuery.Reset()
				fmt.Println("Query cleared.")
			case cmd == ".tables":
				// List all tables, including nameserver-specific ones
				executeQuery(client, accessToken, projectID, serverID, 
					"SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name")
			case cmd == ".nameservers" || cmd == ".ns":
				// List available nameservers for context
				databasesResponse, err := client.ListDatabases(accessToken, projectID, serverID)
				if err == nil {
					fmt.Println("Available nameservers:")
					for _, db := range databasesResponse.Databases {
						if db.IsActive {
							fmt.Printf("  - %s (ID: %s)\n", db.DatabaseName, db.ID)
						}
					}
					fmt.Println()
					fmt.Println("Note: Tables are named like: conversations_{nameserver_name}")
					fmt.Println("Example: If nameserver is 'name1', use 'conversations_name1'")
				} else {
					fmt.Printf("Error listing nameservers: %v\n", err)
				}
			case strings.HasPrefix(cmd, ".schema"):
				parts := strings.Fields(cmd)
				if len(parts) > 1 {
					tableName := parts[1]
					executeQuery(client, accessToken, projectID, serverID, 
						fmt.Sprintf("SELECT sql FROM sqlite_master WHERE type='table' AND name = '%s'", tableName))
				} else {
					fmt.Println("Usage: .schema <table_name>")
				}
			default:
				fmt.Printf("Unknown command: %s\n", line)
				fmt.Println("Type \".help\" for available commands.")
			}
			currentQuery.Reset()
			continue
		}

		// Add line to current query
		if currentQuery.Len() > 0 {
			currentQuery.WriteString(" ")
		}
		currentQuery.WriteString(line)

		// Check if line ends with semicolon (end of query)
		// Also handle queries that are wrapped in quotes (remove quotes)
		trimmedLine := strings.TrimRight(line, " \t")
		if strings.HasSuffix(trimmedLine, ";") {
			query := strings.TrimSpace(currentQuery.String())
			// Remove trailing semicolon
			query = strings.TrimSuffix(query, ";")
			query = strings.TrimSpace(query)
			
			// Remove surrounding quotes if present
			if len(query) >= 2 {
				if (query[0] == '"' && query[len(query)-1] == '"') ||
				   (query[0] == '\'' && query[len(query)-1] == '\'') {
					query = query[1 : len(query)-1]
				}
			}
			
			if query != "" {
				executeQuery(client, accessToken, projectID, serverID, query)
			}
			currentQuery.Reset()
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	return nil
}

// executeQuery executes a SQL query and displays results
func executeQuery(client *api.Client, accessToken, projectID, serverID, query string) {
	queryArgs := []interface{}{}
	
	queryResponse, err := client.ExecuteQuery(accessToken, projectID, serverID, query, queryArgs)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			fmt.Printf("Error: %s\n", apiErr.Error())
		} else {
			fmt.Printf("Error: %v\n", err)
		}
		return
	}

	if !queryResponse.Success {
		if queryResponse.ErrorMessage != "" {
			fmt.Printf("Error: %s\n", queryResponse.ErrorMessage)
		} else {
			fmt.Println("Query failed")
		}
		return
	}

	// Display results
	if len(queryResponse.Columns) > 0 {
		// SELECT query - display results in table
		if len(queryResponse.Rows) == 0 {
			fmt.Println("No rows returned.")
			return
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
		fmt.Printf("Rows returned: %d (%dms)\n", len(queryResponse.Rows), queryResponse.ExecutionTime)
	} else {
		// INSERT/UPDATE/DELETE query
		fmt.Printf("Query executed successfully (%dms)\n", queryResponse.ExecutionTime)
		fmt.Printf("Rows affected: %d\n", queryResponse.RowsAffected)
	}
}

// printHelp displays available shell commands
func printHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  .help, .h          Show this help message")
	fmt.Println("  .quit, .exit, .q  Exit the shell")
	fmt.Println("  .clear, .c        Clear the current query")
	fmt.Println("  .tables           List all tables")
	fmt.Println("  .schema <table>   Show schema for a table")
	fmt.Println("  .nameservers, .ns List available nameservers")
	fmt.Println()
	fmt.Println("SQL queries:")
	fmt.Println("  Enter SQL queries directly. End with semicolon (;) or empty line to execute.")
	fmt.Println("  Multi-line queries are supported.")
	fmt.Println()
	fmt.Println("Table naming:")
	fmt.Println("  Tables are named with nameserver suffix: conversations_{nameserver_name}")
	fmt.Println("  Example: If nameserver is 'name1', use 'conversations_name1'")
	fmt.Println("  Use .tables to see all available tables")
}
