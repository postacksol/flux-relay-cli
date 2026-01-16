package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
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

	// Set up signal handler for Ctrl+C (like Turso - never exits, only .quit does)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)

	// Handle Ctrl+C in a goroutine - never exits, only cancels queries
	go func() {
		for {
			<-sigChan
			if currentQuery.Len() > 0 {
				// Clear current query if one is in progress
				currentQuery.Reset()
				fmt.Println()
				fmt.Println("^C")
				fmt.Println("Query cancelled.")
			} else {
				// Just show a message, never exit
				fmt.Println()
				fmt.Println("^C")
				fmt.Println("Use '.quit' to exit the shell.")
			}
		}
	}()

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
			case cmd == ".examples" || cmd == ".ex":
				printExamples()
			case cmd == ".clear" || cmd == ".c":
				currentQuery.Reset()
				fmt.Println("Query cleared.")
			case cmd == ".tables":
				// Show all tables for all nameservers in this server
				// The API automatically filters to show only nameserver-specific tables
				databasesResponse, err := client.ListDatabases(accessToken, projectID, serverID)
				if err == nil && len(databasesResponse.Databases) > 0 {
					// Get all nameservers
					activeNameservers := make([]string, 0)
					for _, db := range databasesResponse.Databases {
						if db.IsActive {
							activeNameservers = append(activeNameservers, db.DatabaseName)
						}
					}
					
					if len(activeNameservers) > 0 {
						fmt.Printf("Showing tables for %d nameserver(s) in this server:\n", len(activeNameservers))
						for _, ns := range activeNameservers {
							fmt.Printf("  - %s\n", ns)
						}
						fmt.Println()
					}
				}
				
				// Query all tables - API will filter to show only nameserver-specific tables
				executeQuery(client, accessToken, projectID, serverID,
					"SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name")
				
				if err != nil {
					fmt.Printf("\nNote: Could not list nameservers: %v\n", err)
				} else if len(databasesResponse.Databases) == 0 {
					fmt.Println("\nNote: No nameservers found. Initialize a nameserver to create tables.")
				}
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
			case strings.HasPrefix(cmd, ".create_table") || strings.HasPrefix(cmd, ".create"):
				// Helper for creating tables - shows example
				fmt.Println("To create a table, use SQL directly:")
				fmt.Println("  CREATE TABLE conversations_name1 (id TEXT PRIMARY KEY, server_id TEXT, ...);")
				fmt.Println()
				fmt.Println("Note: Table names must follow the pattern: {baseName}_{nameserverName}")
				fmt.Println("Example: conversations_name1, messages_name1, etc.")
				fmt.Println()
				fmt.Println("Use .nameservers to see available nameserver names.")
			case strings.HasPrefix(cmd, ".drop_table") || strings.HasPrefix(cmd, ".drop"):
				parts := strings.Fields(cmd)
				if len(parts) > 1 {
					tableName := parts[1]
					fmt.Printf("To drop table '%s', use:\n", tableName)
					fmt.Printf("  DROP TABLE %s;\n", tableName)
					fmt.Println()
					fmt.Println("Or execute directly:")
					executeQuery(client, accessToken, projectID, serverID,
						fmt.Sprintf("DROP TABLE %s", tableName))
				} else {
					fmt.Println("Usage: .drop_table <table_name>")
					fmt.Println("Example: .drop_table conversations_name1")
				}
			case strings.HasPrefix(cmd, ".alter_table") || strings.HasPrefix(cmd, ".alter"):
				// Helper for altering tables - shows example
				fmt.Println("To alter a table, use SQL directly:")
				fmt.Println("  ALTER TABLE conversations_name1 ADD COLUMN new_field TEXT;")
				fmt.Println("  ALTER TABLE conversations_name1 RENAME COLUMN old_field TO new_field;")
				fmt.Println()
				fmt.Println("Note: You can only alter tables that belong to your server's nameservers.")
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

	// Check for errors - but also handle cases where Success might not be set but we have data
	if !queryResponse.Success && queryResponse.ErrorMessage != "" {
		fmt.Printf("Error: %s\n", queryResponse.ErrorMessage)
		return
	}
	
	// If Success is false but no error message, and we have no data, it might be an empty result
	if !queryResponse.Success && queryResponse.ErrorMessage == "" && len(queryResponse.Columns) == 0 && len(queryResponse.Rows) == 0 {
		fmt.Println("Query returned no results.")
		fmt.Println()
		fmt.Println("This could mean:")
		fmt.Println("  - No tables exist yet (initialize your nameserver schema)")
		fmt.Println("  - Tables don't match the expected pattern")
		fmt.Println("  - Use .nameservers to see available nameservers")
		return
	}

	// Display results
	if len(queryResponse.Columns) > 0 {
		// SELECT query - display results in table
		if len(queryResponse.Rows) == 0 {
			fmt.Println("No rows returned.")
			fmt.Println()
			fmt.Println("Note: If you expected to see tables, make sure:")
			fmt.Println("  1. Your nameserver has been initialized")
			fmt.Println("  2. Tables follow the pattern: {baseName}_{nameserverName}")
			fmt.Println("  3. Use .nameservers to see available nameservers")
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
	fmt.Println("  .examples, .ex     Show example queries and operations")
	fmt.Println("  .quit, .exit, .q  Exit the shell")
	fmt.Println("  .clear, .c        Clear the current query")
	fmt.Println("  .tables           List all tables")
	fmt.Println("  .schema <table>   Show schema for a table")
	fmt.Println("  .nameservers, .ns List available nameservers")
	fmt.Println("  .drop_table <name> Drop a table (with confirmation)")
	fmt.Println()
	fmt.Println("SQL queries:")
	fmt.Println("  Enter SQL queries directly. End with semicolon (;) or empty line to execute.")
	fmt.Println("  Multi-line queries are supported.")
	fmt.Println()
	fmt.Println("Table management:")
	fmt.Println("  CREATE TABLE - Create new tables (must follow pattern: {baseName}_{nameserverName})")
	fmt.Println("  ALTER TABLE  - Modify table structure (add/rename columns, etc.)")
	fmt.Println("  DROP TABLE   - Delete tables (use .drop_table for helper)")
	fmt.Println()
	fmt.Println("Table naming:")
	fmt.Println("  Tables are named with nameserver suffix: conversations_{nameserver_name}")
	fmt.Println("  Example: If nameserver is 'name1', use 'conversations_name1'")
	fmt.Println("  Use .tables to see all available tables")
	fmt.Println()
	fmt.Println("Security:")
	fmt.Println("  You can only create/modify tables for your server's nameservers")
	fmt.Println("  System/platform tables are not accessible")
	fmt.Println()
	fmt.Println("Type '.examples' for example queries and operations")
}

// printExamples displays example queries and operations
func printExamples() {
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("Example Queries and Operations")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()
	
	fmt.Println("📋 LISTING DATA:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	fmt.Println()
	fmt.Println("1. List all tables:")
	fmt.Println("   .tables")
	fmt.Println()
	fmt.Println("2. Show table schema:")
	fmt.Println("   .schema conversations_name1")
	fmt.Println("   SELECT sql FROM sqlite_master WHERE type='table' AND name='conversations_name1';")
	fmt.Println()
	fmt.Println("3. List all conversations (replace 'name1' with your nameserver):")
	fmt.Println("   SELECT * FROM conversations_name1 WHERE server_id = ? LIMIT 10;")
	fmt.Println()
	fmt.Println("4. Count records in a table:")
	fmt.Println("   SELECT COUNT(*) as total FROM conversations_name1 WHERE server_id = ?;")
	fmt.Println()
	fmt.Println("5. List with specific columns:")
	fmt.Println("   SELECT id, created_at, title FROM conversations_name1 WHERE server_id = ? ORDER BY created_at DESC LIMIT 5;")
	fmt.Println()
	fmt.Println("6. Filter and search:")
	fmt.Println("   SELECT * FROM conversations_name1 WHERE server_id = ? AND title LIKE '%search%' LIMIT 10;")
	fmt.Println()
	fmt.Println("7. Join tables (if you have related tables):")
	fmt.Println("   SELECT c.id, c.title, COUNT(m.id) as message_count")
	fmt.Println("   FROM conversations_name1 c")
	fmt.Println("   LEFT JOIN messages_name1 m ON c.id = m.conversation_id")
	fmt.Println("   WHERE c.server_id = ?")
	fmt.Println("   GROUP BY c.id LIMIT 10;")
	fmt.Println()
	
	fmt.Println("📊 VIEWING DATA:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	fmt.Println()
	fmt.Println("8. View recent conversations:")
	fmt.Println("   SELECT * FROM conversations_name1 WHERE server_id = ? ORDER BY created_at DESC LIMIT 20;")
	fmt.Println()
	fmt.Println("9. View messages in a conversation:")
	fmt.Println("   SELECT * FROM messages_name1 WHERE server_id = ? AND conversation_id = 'conv_123' ORDER BY created_at;")
	fmt.Println()
	fmt.Println("10. View end users:")
	fmt.Println("    SELECT * FROM end_users_name1 WHERE server_id = ? LIMIT 10;")
	fmt.Println()
	fmt.Println("11. View with date range:")
	fmt.Println("    SELECT * FROM conversations_name1")
	fmt.Println("    WHERE server_id = ? AND created_at >= datetime('now', '-7 days')")
	fmt.Println("    ORDER BY created_at DESC;")
	fmt.Println()
	
	fmt.Println("➕ INSERTING DATA:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	fmt.Println()
	fmt.Println("12. Insert a new conversation:")
	fmt.Println("    INSERT INTO conversations_name1 (id, server_id, title, created_at)")
	fmt.Println("    VALUES ('conv_123', ?, 'My Conversation', datetime('now'));")
	fmt.Println()
	fmt.Println("13. Insert a message:")
	fmt.Println("    INSERT INTO messages_name1 (id, server_id, conversation_id, content, created_at)")
	fmt.Println("    VALUES ('msg_456', ?, 'conv_123', 'Hello!', datetime('now'));")
	fmt.Println()
	fmt.Println("14. Insert multiple records:")
	fmt.Println("    INSERT INTO conversations_name1 (id, server_id, title, created_at)")
	fmt.Println("    VALUES")
	fmt.Println("      ('conv_1', ?, 'First', datetime('now')),")
	fmt.Println("      ('conv_2', ?, 'Second', datetime('now')),")
	fmt.Println("      ('conv_3', ?, 'Third', datetime('now'));")
	fmt.Println()
	
	fmt.Println("✏️  UPDATING DATA:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	fmt.Println()
	fmt.Println("15. Update a conversation:")
	fmt.Println("    UPDATE conversations_name1")
	fmt.Println("    SET title = 'Updated Title'")
	fmt.Println("    WHERE id = 'conv_123' AND server_id = ?;")
	fmt.Println()
	fmt.Println("16. Update multiple fields:")
	fmt.Println("    UPDATE conversations_name1")
	fmt.Println("    SET title = 'New Title', updated_at = datetime('now')")
	fmt.Println("    WHERE id = 'conv_123' AND server_id = ?;")
	fmt.Println()
	fmt.Println("17. Update with condition:")
	fmt.Println("    UPDATE messages_name1")
	fmt.Println("    SET content = 'Edited message'")
	fmt.Println("    WHERE id = 'msg_456' AND server_id = ?;")
	fmt.Println()
	
	fmt.Println("🗑️  DELETING DATA:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	fmt.Println()
	fmt.Println("18. Delete a specific record:")
	fmt.Println("    DELETE FROM conversations_name1 WHERE id = 'conv_123' AND server_id = ?;")
	fmt.Println()
	fmt.Println("19. Delete with condition:")
	fmt.Println("    DELETE FROM messages_name1 WHERE conversation_id = 'conv_123' AND server_id = ?;")
	fmt.Println()
	fmt.Println("20. Delete old records:")
	fmt.Println("    DELETE FROM conversations_name1")
	fmt.Println("    WHERE server_id = ? AND created_at < datetime('now', '-30 days');")
	fmt.Println()
	
	fmt.Println("🏗️  TABLE MANAGEMENT:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	fmt.Println()
	fmt.Println("21. Create a custom table:")
	fmt.Println("    CREATE TABLE custom_data_name1 (")
	fmt.Println("      id TEXT PRIMARY KEY,")
	fmt.Println("      server_id TEXT NOT NULL,")
	fmt.Println("      data TEXT,")
	fmt.Println("      created_at TEXT DEFAULT (datetime('now'))")
	fmt.Println("    );")
	fmt.Println()
	fmt.Println("22. Add a column to existing table:")
	fmt.Println("    ALTER TABLE conversations_name1 ADD COLUMN status TEXT DEFAULT 'active';")
	fmt.Println()
	fmt.Println("23. Rename a column:")
	fmt.Println("    ALTER TABLE conversations_name1 RENAME COLUMN old_name TO new_name;")
	fmt.Println()
	fmt.Println("24. Drop a table:")
	fmt.Println("    DROP TABLE custom_data_name1;")
	fmt.Println("    -- Or use: .drop_table custom_data_name1")
	fmt.Println()
	
	fmt.Println("📈 AGGREGATIONS & STATISTICS:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	fmt.Println()
	fmt.Println("25. Count by group:")
	fmt.Println("    SELECT status, COUNT(*) as count")
	fmt.Println("    FROM conversations_name1")
	fmt.Println("    WHERE server_id = ?")
	fmt.Println("    GROUP BY status;")
	fmt.Println()
	fmt.Println("26. Get statistics:")
	fmt.Println("    SELECT")
	fmt.Println("      COUNT(*) as total_conversations,")
	fmt.Println("      COUNT(DISTINCT conversation_id) as unique_conversations,")
	fmt.Println("      MIN(created_at) as oldest,")
	fmt.Println("      MAX(created_at) as newest")
	fmt.Println("    FROM conversations_name1 WHERE server_id = ?;")
	fmt.Println()
	fmt.Println("27. Top conversations by message count:")
	fmt.Println("    SELECT c.id, c.title, COUNT(m.id) as message_count")
	fmt.Println("    FROM conversations_name1 c")
	fmt.Println("    LEFT JOIN messages_name1 m ON c.id = m.conversation_id")
	fmt.Println("    WHERE c.server_id = ?")
	fmt.Println("    GROUP BY c.id")
	fmt.Println("    ORDER BY message_count DESC LIMIT 10;")
	fmt.Println()
	
	fmt.Println("⚠️  IMPORTANT NOTES:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	fmt.Println("• Replace 'name1' with your actual nameserver name")
	fmt.Println("• All queries must include 'WHERE server_id = ?' for data isolation")
	fmt.Println("• Table names must follow pattern: {baseName}_{nameserverName}")
	fmt.Println("• Use .nameservers to see available nameserver names")
	fmt.Println("• Use .tables to see all available tables")
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
}
