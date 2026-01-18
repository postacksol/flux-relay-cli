package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"text/tabwriter"

	"github.com/postacksol/flux-relay-cli/internal/api"
	"github.com/postacksol/flux-relay-cli/internal/config"
)

// runServerShell starts an interactive shell for a server
func runServerShell(serverIdentifier string) error {
	// Get API URL
	apiURL := getAPIURL()

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

// Shell context to track current nameserver
type shellContext struct {
	projectID      string
	serverID       string
	serverName     string
	nameserverID   string
	nameserverName string
	client         *api.Client
	accessToken    string
	cfg            *config.ConfigManager
}

// startShell runs the interactive SQL shell
func startShell(cfg *config.ConfigManager, client *api.Client, accessToken, projectID, serverID, serverName, nameserverName string) error {
	// Get nameserver ID if nameserver name is provided
	var nameserverID string
	if nameserverName != "" {
		databasesResponse, err := client.ListDatabases(accessToken, projectID, serverID)
		if err == nil {
			for _, db := range databasesResponse.Databases {
				if db.DatabaseName == nameserverName {
					nameserverID = db.ID
					break
				}
			}
		}
	}

	ctx := &shellContext{
		projectID:      projectID,
		serverID:       serverID,
		serverName:     serverName,
		nameserverID:   nameserverID,
		nameserverName: nameserverName,
		client:         client,
		accessToken:    accessToken,
		cfg:            cfg,
	}

	return startShellWithContext(ctx)
}

func startShellWithContext(ctx *shellContext) error {
	// Print welcome message
	fmt.Printf("Connected to %s", ctx.serverName)
	if ctx.nameserverName != "" {
		fmt.Printf(" (nameserver: %s)", ctx.nameserverName)
	}
	fmt.Println()
	fmt.Println()
	fmt.Println("Welcome to Flux Relay SQL shell!")
	fmt.Println()
	fmt.Println("Type \".quit\" to exit the shell and \".help\" to list all available commands.")
	fmt.Println()
	fmt.Println("Enter SQL queries directly (no 'sql' prefix needed).")
	fmt.Println("End queries with semicolon (;) or press Enter twice to execute.")
	if ctx.nameserverName != "" {
		fmt.Println()
		fmt.Printf("📌 Current nameserver: %s\n", ctx.nameserverName)
		fmt.Printf("Note: Tables use nameserver suffix. Example: conversations_%s\n", ctx.nameserverName)
		fmt.Println("Use .tables to see all available tables.")
	} else {
		fmt.Println()
		fmt.Println("📌 Server context: All nameservers")
		fmt.Println("Use .nameservers to see available nameservers")
		fmt.Println("Use .use <nameserver> to switch to a specific nameserver")
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
					executeQuery(ctx.client, ctx.accessToken, ctx.projectID, ctx.serverID, query)
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
			case cmd == ".context" || cmd == ".ctx":
				// Show current context
				fmt.Printf("Current context:\n")
				fmt.Printf("  Server: %s (%s)\n", ctx.serverName, ctx.serverID)
				if ctx.nameserverName != "" {
					fmt.Printf("  Nameserver: %s (%s)\n", ctx.nameserverName, ctx.nameserverID)
					fmt.Printf("  Table suffix: conversations_%s\n", ctx.nameserverName)
				} else {
					fmt.Println("  Nameserver: (none - all nameservers)")
				}
			case cmd == ".tables":
				// Show all tables for all nameservers in this server
				// The API automatically filters to show only nameserver-specific tables
				databasesResponse, err := ctx.client.ListDatabases(ctx.accessToken, ctx.projectID, ctx.serverID)
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
							marker := "  "
							if ctx.nameserverName == ns {
								marker = "→ "
							}
							fmt.Printf("%s%s\n", marker, ns)
						}
						fmt.Println()
					}
				}
				
				// Query all tables - API will filter to show only nameserver-specific tables
				executeQuery(ctx.client, ctx.accessToken, ctx.projectID, ctx.serverID,
					"SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name")
				
				if err != nil {
					fmt.Printf("\nNote: Could not list nameservers: %v\n", err)
				} else if len(databasesResponse.Databases) == 0 {
					fmt.Println("\nNote: No nameservers found. Create one with: .create_ns <name>")
				}
				case cmd == ".nameservers" || cmd == ".ns":
				// List available nameservers for context
				databasesResponse, err := ctx.client.ListDatabases(ctx.accessToken, ctx.projectID, ctx.serverID)
				if err == nil {
					activeCount := 0
					inactiveCount := 0
					
					// Count first
					for _, db := range databasesResponse.Databases {
						if db.IsActive {
							activeCount++
						} else {
							inactiveCount++
						}
					}
					
					if activeCount > 0 {
						fmt.Println("Active nameservers:")
						for _, db := range databasesResponse.Databases {
							if db.IsActive {
								marker := "  "
								if ctx.nameserverID == db.ID {
									marker = "→ "
								}
								fmt.Printf("%s%s (ID: %s)\n", marker, db.DatabaseName, db.ID)
							}
						}
						fmt.Println()
					}
					
					if inactiveCount > 0 {
						fmt.Println("Inactive (soft-deleted) nameservers:")
						for _, db := range databasesResponse.Databases {
							if !db.IsActive {
								fmt.Printf("  %s (ID: %s) [inactive]\n", db.DatabaseName, db.ID)
							}
						}
						fmt.Println()
						fmt.Println("Note: Inactive nameservers can prevent creating new ones with the same name.")
						fmt.Println("      The system will reactivate them if you try to create a duplicate.")
						fmt.Println()
					}
					
					if activeCount == 0 && inactiveCount == 0 {
						fmt.Println("No nameservers found.")
						fmt.Println()
					}
					
					fmt.Println("Note: Tables are named like: conversations_{nameserver_name}")
					fmt.Println("Example: If nameserver is 'name1', use 'conversations_name1'")
					fmt.Println()
					fmt.Println("Commands:")
					fmt.Println("  .use <nameserver>  - Switch to a nameserver context")
					fmt.Println("  .create_ns <name>  - Create a new nameserver")
				} else {
					fmt.Printf("Error listing nameservers: %v\n", err)
				}
			case strings.HasPrefix(cmd, ".use"):
				parts := strings.Fields(cmd)
				if len(parts) > 1 {
					nameserverName := parts[1]
					// Find nameserver
					databasesResponse, err := ctx.client.ListDatabases(ctx.accessToken, ctx.projectID, ctx.serverID)
					if err != nil {
						fmt.Printf("Error: %v\n", err)
						break
					}
					
					var found *api.Database
					for i := range databasesResponse.Databases {
						db := &databasesResponse.Databases[i]
						if db.DatabaseName == nameserverName || db.ID == nameserverName {
							found = db
							break
						}
					}
					
					if found == nil {
						fmt.Printf("Nameserver '%s' not found. Use .nameservers to see available nameservers.\n", nameserverName)
					} else {
						ctx.nameserverID = found.ID
						ctx.nameserverName = found.DatabaseName
						fmt.Printf("✅ Switched to nameserver: %s\n", found.DatabaseName)
						fmt.Printf("   Tables will use suffix: conversations_%s\n", found.DatabaseName)
					}
				} else {
					if ctx.nameserverName != "" {
						fmt.Printf("Current nameserver: %s\n", ctx.nameserverName)
					} else {
						fmt.Println("No nameserver selected. Use .use <nameserver> to select one.")
					}
				}
			case strings.HasPrefix(cmd, ".create_ns") || strings.HasPrefix(cmd, ".create_nameserver"):
				parts := strings.Fields(cmd)
				if len(parts) > 1 {
					nameserverName := strings.Join(parts[1:], " ")
					if nameserverName == "" {
						fmt.Println("Usage: .create_ns <nameserver_name>")
						fmt.Println("Example: .create_ns db2")
						break
					}
					
					// First, check existing nameservers to help debug conflicts
					databasesResponse, listErr := ctx.client.ListDatabases(ctx.accessToken, ctx.projectID, ctx.serverID)
					if listErr == nil && len(databasesResponse.Databases) > 0 {
						// Check for case-insensitive match
						requestedLower := strings.ToLower(nameserverName)
						for _, db := range databasesResponse.Databases {
							if strings.ToLower(db.DatabaseName) == requestedLower {
								if db.DatabaseName == nameserverName {
									// Exact match
									if db.IsActive {
										fmt.Printf("⚠️  Nameserver '%s' already exists and is active.\n", db.DatabaseName)
										fmt.Printf("   ID: %s\n", db.ID)
										fmt.Println()
										fmt.Println("Use .use " + db.DatabaseName + " to switch to it.")
									} else {
										fmt.Printf("⚠️  Found inactive nameserver '%s' - will be reactivated.\n", db.DatabaseName)
										fmt.Printf("   ID: %s\n", db.ID)
									}
								} else {
									// Case-insensitive match but different case
									fmt.Printf("⚠️  Conflict: A nameserver with a similar name already exists:\n")
									fmt.Printf("   Requested: '%s'\n", nameserverName)
									fmt.Printf("   Existing:  '%s' (ID: %s)\n", db.DatabaseName, db.ID)
									fmt.Println()
									fmt.Println("Note: Nameserver names are case-insensitive in the database.")
									fmt.Println("      Use the existing nameserver or choose a different name.")
									break
								}
							}
						}
					}
					
					fmt.Printf("Creating nameserver '%s'...\n", nameserverName)
					response, err := ctx.client.CreateNameserver(ctx.accessToken, ctx.projectID, ctx.serverID, nameserverName)
					if err != nil {
						if apiErr, ok := err.(*api.APIError); ok {
							errorMsg := apiErr.Error()
							fmt.Printf("Error: %s\n", errorMsg)
							fmt.Println()
							
							// Check if error suggests an inactive nameserver exists
							if strings.Contains(errorMsg, "already exists") {
								fmt.Println("💡 This error usually means:")
								fmt.Println("   1. An active nameserver with this name exists, OR")
								fmt.Println("   2. An inactive (soft-deleted) nameserver exists and should be reactivated")
								fmt.Println()
								fmt.Println("The API should automatically reactivate inactive nameservers.")
								fmt.Println("If this keeps happening, the nameserver might be active but not visible.")
								fmt.Println()
								
								// Try to query directly for the nameserver using the API
								fmt.Println("💡 Troubleshooting tips:")
								fmt.Println("   - The API should automatically reactivate inactive nameservers")
								fmt.Println("   - If this error persists, there may be an active nameserver")
								fmt.Println("     with this name that's not visible in .nameservers")
								fmt.Println("   - Try using a different name, or contact support if needed")
								fmt.Println()
							}
							
							// Show existing nameservers to help user
							if listErr == nil && len(databasesResponse.Databases) > 0 {
								fmt.Println("Currently visible nameservers in this server:")
								for _, db := range databasesResponse.Databases {
									if db.IsActive {
										fmt.Printf("  - %s (ID: %s)\n", db.DatabaseName, db.ID)
									}
								}
								fmt.Println()
								fmt.Println("Note: Inactive nameservers may not be visible but can still block creation.")
								fmt.Println("      The API should reactivate them automatically when you try to create.")
							}
						} else {
							fmt.Printf("Error: %v\n", err)
						}
						break
					}
					
					// Check if it was reactivated
					if response.Database.ID != "" {
						// Check if this was a reactivation by looking at creation time
						fmt.Printf("✅ Nameserver '%s' created successfully!\n", response.Database.DatabaseName)
						fmt.Printf("   ID: %s\n", response.Database.ID)
						fmt.Println()
						fmt.Println("Next steps:")
						fmt.Println("  1. Initialize schema: .init_ns " + response.Database.DatabaseName)
						fmt.Println("  2. Switch to it: .use " + response.Database.DatabaseName)
						fmt.Println("  3. Create tables: CREATE TABLE conversations_" + response.Database.DatabaseName + " (...);")
					}
				} else {
					fmt.Println("Usage: .create_ns <nameserver_name>")
					fmt.Println("Example: .create_ns db2")
					fmt.Println()
					fmt.Println("This creates a new nameserver in the current server.")
					fmt.Println()
					fmt.Println("Use .nameservers to see existing nameservers first.")
				}
			case strings.HasPrefix(cmd, ".init_ns") || strings.HasPrefix(cmd, ".init_nameserver") || strings.HasPrefix(cmd, ".initialize"):
				parts := strings.Fields(cmd)
				var nameserverID string
				var nameserverName string
				
				if len(parts) > 1 {
					nameserverIdentifier := strings.Join(parts[1:], " ")
					// Find nameserver
					databasesResponse, err := ctx.client.ListDatabases(ctx.accessToken, ctx.projectID, ctx.serverID)
					if err != nil {
						fmt.Printf("Error: %v\n", err)
						break
					}
					
					var found *api.Database
					for i := range databasesResponse.Databases {
						db := &databasesResponse.Databases[i]
						if db.DatabaseName == nameserverIdentifier || db.ID == nameserverIdentifier {
							found = db
							break
						}
					}
					
					if found == nil {
						fmt.Printf("Nameserver '%s' not found.\n", nameserverIdentifier)
						break
					}
					
					nameserverID = found.ID
					nameserverName = found.DatabaseName
				} else if ctx.nameserverID != "" {
					nameserverID = ctx.nameserverID
					nameserverName = ctx.nameserverName
				} else {
					fmt.Println("Usage: .init_ns <nameserver_name>")
					fmt.Println("Example: .init_ns name1")
					fmt.Println()
					fmt.Println("Or switch to a nameserver first: .use name1")
					break
				}
				
				fmt.Printf("Initializing schema for nameserver '%s'...\n", nameserverName)
				response, err := ctx.client.InitializeNameserver(ctx.accessToken, ctx.projectID, ctx.serverID, nameserverID)
				if err != nil {
					if apiErr, ok := err.(*api.APIError); ok {
						fmt.Printf("Error: %s\n", apiErr.Error())
					} else {
						fmt.Printf("Error: %v\n", err)
					}
					break
				}
				
				fmt.Printf("✅ Schema initialized for '%s'!\n", nameserverName)
				if response.TablesCreated > 0 {
					fmt.Printf("   Created %d tables\n", response.TablesCreated)
				}
				if len(response.VerifiedTables) > 0 {
					fmt.Println("   Tables:")
					for _, table := range response.VerifiedTables {
						fmt.Printf("     - %s\n", table)
					}
				} else if response.TablesCreated > 0 {
					fmt.Println("   (Tables created but list not available)")
				}
				if response.Note != "" {
					fmt.Println()
					fmt.Println("   " + response.Note)
				}
				fmt.Println()
				fmt.Println("You can now:")
				fmt.Printf("  .use %s  - Switch to this nameserver\n", nameserverName)
				fmt.Printf("  .tables  - See all tables\n")
				fmt.Println()
				fmt.Println("Or create custom tables manually:")
				fmt.Printf("  CREATE TABLE custom_table_%s (id TEXT PRIMARY KEY, server_id TEXT, data TEXT);\n", nameserverName)
			case strings.HasPrefix(cmd, ".schema"):
				parts := strings.Fields(cmd)
				if len(parts) > 1 {
					tableName := parts[1]
					executeQuery(ctx.client, ctx.accessToken, ctx.projectID, ctx.serverID,
						fmt.Sprintf("SELECT sql FROM sqlite_master WHERE type='table' AND name = '%s'", tableName))
				} else {
					fmt.Println("Usage: .schema <table_name>")
				}
			case strings.HasPrefix(cmd, ".create_table") || strings.HasPrefix(cmd, ".create"):
				// Helper for creating tables - shows example
				parts := strings.Fields(cmd)
				if len(parts) > 1 {
					// User provided table name
					tableName := strings.Join(parts[1:], " ")
					if ctx.nameserverName != "" {
						fmt.Printf("To create table '%s' for nameserver '%s', use:\n", tableName, ctx.nameserverName)
						fmt.Printf("  CREATE TABLE %s_%s (id TEXT PRIMARY KEY, server_id TEXT, ...);\n", tableName, ctx.nameserverName)
						fmt.Println()
						fmt.Println("Or if you want a custom name:")
						fmt.Printf("  CREATE TABLE %s (id TEXT PRIMARY KEY, server_id TEXT, ...);\n", tableName)
						fmt.Println()
						fmt.Println("Note: Table names must follow the pattern: {baseName}_{nameserverName}")
						fmt.Println("      Or use any name - the API will validate it's for your nameserver.")
					} else {
						fmt.Printf("To create table '%s', first switch to a nameserver:\n", tableName)
						fmt.Println("  .use <nameserver>")
						fmt.Println()
						fmt.Println("Then create the table:")
						fmt.Printf("  CREATE TABLE %s_<nameserver> (id TEXT PRIMARY KEY, server_id TEXT, ...);\n", tableName)
					}
				} else {
					// Show general help
					fmt.Println("To create a table, use SQL directly:")
					if ctx.nameserverName != "" {
						fmt.Printf("  CREATE TABLE my_table_%s (id TEXT PRIMARY KEY, server_id TEXT, data TEXT);\n", ctx.nameserverName)
						fmt.Println()
						fmt.Printf("Current nameserver: %s\n", ctx.nameserverName)
					} else {
						fmt.Println("  CREATE TABLE my_table_<nameserver> (id TEXT PRIMARY KEY, server_id TEXT, ...);")
						fmt.Println()
						fmt.Println("First switch to a nameserver: .use <nameserver>")
					}
					fmt.Println()
					fmt.Println("Note: Table names must follow the pattern: {baseName}_{nameserverName}")
					fmt.Println("Example: conversations_name1, messages_name1, custom_table_db2, etc.")
					fmt.Println()
					fmt.Println("Use .nameservers to see available nameserver names.")
					fmt.Println("Use .use <nameserver> to set the context.")
				}
			case strings.HasPrefix(cmd, ".drop_table") || strings.HasPrefix(cmd, ".drop"):
				parts := strings.Fields(cmd)
				if len(parts) > 1 {
					tableName := parts[1]
					fmt.Printf("To drop table '%s', use:\n", tableName)
					fmt.Printf("  DROP TABLE %s;\n", tableName)
					fmt.Println()
					fmt.Println("Or execute directly:")
					executeQuery(ctx.client, ctx.accessToken, ctx.projectID, ctx.serverID,
						fmt.Sprintf("DROP TABLE %s", tableName))
				} else {
					fmt.Println("Usage: .drop_table <table_name>")
					if ctx.nameserverName != "" {
						fmt.Printf("Example: .drop_table conversations_%s\n", ctx.nameserverName)
					} else {
						fmt.Println("Example: .drop_table conversations_name1")
					}
				}
			case strings.HasPrefix(cmd, ".alter_table") || strings.HasPrefix(cmd, ".alter"):
				// Helper for altering tables - shows example
				if ctx.nameserverName != "" {
					fmt.Printf("Current nameserver: %s\n", ctx.nameserverName)
					fmt.Println()
					fmt.Println("Common schema customizations:")
					fmt.Println()
					fmt.Println("1. Add a column to conversations:")
					fmt.Printf("   ALTER TABLE conversations_%s ADD COLUMN priority INTEGER DEFAULT 0;\n", ctx.nameserverName)
					fmt.Printf("   ALTER TABLE conversations_%s ADD COLUMN tags TEXT;\n", ctx.nameserverName)
					fmt.Println()
					fmt.Println("2. Add a column to messages:")
					fmt.Printf("   ALTER TABLE messages_%s ADD COLUMN reactions TEXT DEFAULT '[]';\n", ctx.nameserverName)
					fmt.Printf("   ALTER TABLE messages_%s ADD COLUMN edited_at TEXT;\n", ctx.nameserverName)
					fmt.Println()
					fmt.Println("3. Add a column to end_users:")
					fmt.Printf("   ALTER TABLE end_users_%s ADD COLUMN avatar_url TEXT;\n", ctx.nameserverName)
					fmt.Printf("   ALTER TABLE end_users_%s ADD COLUMN status TEXT DEFAULT 'offline';\n", ctx.nameserverName)
					fmt.Println()
					fmt.Println("4. Rename a column (SQLite 3.25.0+):")
					fmt.Printf("   ALTER TABLE conversations_%s RENAME COLUMN name TO title;\n", ctx.nameserverName)
					fmt.Println()
					fmt.Println("5. Change data type (requires table recreation):")
					fmt.Println("   -- Step 1: Create new table with desired schema")
					fmt.Printf("   CREATE TABLE conversations_%s_new (\n", ctx.nameserverName)
					fmt.Printf("     id TEXT PRIMARY KEY,\n")
					fmt.Printf("     server_id TEXT NOT NULL,\n")
					fmt.Printf("     priority INTEGER,  -- Changed from TEXT to INTEGER\n")
					fmt.Printf("     created_at TEXT NOT NULL\n")
					fmt.Printf("   );\n")
					fmt.Println("   -- Step 2: Copy data (with type conversion)")
					fmt.Printf("   INSERT INTO conversations_%s_new SELECT id, server_id, CAST(priority AS INTEGER), created_at\n", ctx.nameserverName)
					fmt.Printf("   FROM conversations_%s WHERE server_id = ?;\n", ctx.nameserverName)
					fmt.Println("   -- Step 3: Drop old table")
					fmt.Printf("   DROP TABLE conversations_%s;\n", ctx.nameserverName)
					fmt.Println("   -- Step 4: Rename new table")
					fmt.Printf("   ALTER TABLE conversations_%s_new RENAME TO conversations_%s;\n", ctx.nameserverName, ctx.nameserverName)
					fmt.Println()
					fmt.Println("6. Create an index:")
					fmt.Printf("   CREATE INDEX idx_conversations_%s_priority ON conversations_%s(priority);\n", ctx.nameserverName, ctx.nameserverName)
					fmt.Println()
					fmt.Println("⚠️  Note: SQLite doesn't support direct column type changes.")
					fmt.Println("   To change a column type, you need to recreate the table.")
					fmt.Println("   See example #5 above for the process.")
				} else {
					fmt.Println("To alter a table, use SQL directly:")
					fmt.Println("  ALTER TABLE conversations_name1 ADD COLUMN new_field TEXT;")
					fmt.Println("  ALTER TABLE conversations_name1 RENAME COLUMN old_field TO new_field;")
					fmt.Println()
					fmt.Println("First switch to a nameserver: .use <nameserver>")
				}
				fmt.Println()
				fmt.Println("Note: You can only alter tables that belong to your server's nameservers.")
				fmt.Println("      Use .schema <table> to see current table structure.")
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

			// Check for incomplete queries (common patterns)
			querySoFar := strings.TrimSpace(currentQuery.String() + " " + line)
			queryUpper := strings.ToUpper(querySoFar)
			
			// Check for incomplete LIMIT clause
			if strings.Contains(queryUpper, " LIMIT") && !strings.Contains(queryUpper, " LIMIT ") && !strings.HasSuffix(queryUpper, " LIMIT") {
				// LIMIT with no number - check if it ends with just "LIMIT"
				if strings.HasSuffix(strings.TrimSpace(queryUpper), "LIMIT") {
					fmt.Println("⚠️  Incomplete query: LIMIT requires a number (e.g., LIMIT 10)")
					fmt.Println("   Complete your query or type .clear to start over")
					continue
				}
			}
			
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

				// Validate query completeness before executing
				queryUpperCheck := strings.ToUpper(query)
				if strings.Contains(queryUpperCheck, " LIMIT") {
					// Check if LIMIT has a number after it
					limitPattern := regexp.MustCompile(`LIMIT\s+(\d+)`)
					if !limitPattern.MatchString(queryUpperCheck) && strings.HasSuffix(strings.TrimSpace(queryUpperCheck), "LIMIT") {
						fmt.Println("⚠️  Error: LIMIT requires a number (e.g., LIMIT 10)")
						fmt.Println("   Your query: " + query)
						currentQuery.Reset()
						continue
					}
				}

				if query != "" {
					executeQuery(ctx.client, ctx.accessToken, ctx.projectID, ctx.serverID, query)
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
			errorMsg := apiErr.Error()
			fmt.Printf("Error: %s\n", errorMsg)
			
			// Provide helpful hints for common errors
			if strings.Contains(errorMsg, "SQL_PARSE_ERROR") || strings.Contains(errorMsg, "unexpected end of input") {
				fmt.Println()
				fmt.Println("💡 Common causes:")
				fmt.Println("  - Incomplete query (e.g., LIMIT without a number)")
				fmt.Println("  - Missing semicolon or closing parenthesis")
				fmt.Println("  - Typo in SQL syntax")
				fmt.Println()
				fmt.Println("Example: SELECT * FROM table WHERE server_id = ? LIMIT 10;")
			} else if strings.Contains(errorMsg, "no such table") {
				fmt.Println()
				fmt.Println("💡 Make sure:")
				fmt.Println("  - Table name includes nameserver suffix (e.g., conversations_name1)")
				fmt.Println("  - Use .tables to see available tables")
				fmt.Println("  - Use .nameservers to see nameserver names")
			} else if strings.Contains(errorMsg, "server_id") {
				fmt.Println()
				fmt.Println("💡 Remember: All queries must include WHERE server_id = ?")
			}
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
	fmt.Println("  .help, .h              Show this help message")
	fmt.Println("  .examples, .ex         Show example queries and operations")
	fmt.Println("  .quit, .exit, .q      Exit the shell")
	fmt.Println("  .clear, .c            Clear the current query")
	fmt.Println("  .context, .ctx        Show current context (server/nameserver)")
	fmt.Println("  .tables               List all tables")
	fmt.Println("  .schema <table>       Show schema for a table")
	fmt.Println("  .nameservers, .ns     List available nameservers")
	fmt.Println("  .use <nameserver>     Switch to a nameserver context")
	fmt.Println("  .create_ns <name>     Create a new nameserver")
	fmt.Println("  .init_ns <name>       Initialize schema for a nameserver")
	fmt.Println("  .drop_table <name>    Drop a table")
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
	fmt.Println("Nameserver management:")
	fmt.Println("  .create_ns <name>  - Create a new nameserver")
	fmt.Println("  .init_ns <name>    - Initialize default schema (creates standard tables)")
	fmt.Println("  .use <name>        - Switch context to a nameserver")
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
	fmt.Println("    CREATE TABLE custom_products_name1 (")
	fmt.Println("      id TEXT PRIMARY KEY,")
	fmt.Println("      server_id TEXT NOT NULL,")
	fmt.Println("      name TEXT,")
	fmt.Println("      price REAL,")
	fmt.Println("      created_at TEXT DEFAULT (datetime('now'))")
	fmt.Println("    );")
	fmt.Println()
	fmt.Println("22. Add a column to existing table:")
	fmt.Println("    ALTER TABLE conversations_name1 ADD COLUMN status TEXT DEFAULT 'active';")
	fmt.Println()
	fmt.Println("23. Rename a column:")
	fmt.Println("    ALTER TABLE conversations_name1 RENAME COLUMN old_name TO new_name;")
	fmt.Println()
	fmt.Println("🎨 CUSTOMIZING MESSAGING SCHEMA:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	fmt.Println()
	fmt.Println("24. Add priority to conversations:")
	fmt.Println("    ALTER TABLE conversations_name1 ADD COLUMN priority INTEGER DEFAULT 0;")
	fmt.Println("    CREATE INDEX idx_conversations_name1_priority ON conversations_name1(priority);")
	fmt.Println()
	fmt.Println("25. Add tags/categories to conversations:")
	fmt.Println("    ALTER TABLE conversations_name1 ADD COLUMN tags TEXT DEFAULT '[]';")
	fmt.Println("    ALTER TABLE conversations_name1 ADD COLUMN category TEXT;")
	fmt.Println()
	fmt.Println("26. Add reactions to messages:")
	fmt.Println("    ALTER TABLE messages_name1 ADD COLUMN reactions TEXT DEFAULT '[]';")
	fmt.Println("    ALTER TABLE messages_name1 ADD COLUMN edited_at TEXT;")
	fmt.Println()
	fmt.Println("27. Add user profile fields:")
	fmt.Println("    ALTER TABLE end_users_name1 ADD COLUMN avatar_url TEXT;")
	fmt.Println("    ALTER TABLE end_users_name1 ADD COLUMN status TEXT DEFAULT 'offline';")
	fmt.Println("    ALTER TABLE end_users_name1 ADD COLUMN bio TEXT;")
	fmt.Println()
	fmt.Println("28. Add message metadata:")
	fmt.Println("    ALTER TABLE messages_name1 ADD COLUMN metadata TEXT;")
	fmt.Println("    ALTER TABLE messages_name1 ADD COLUMN reply_to_id TEXT;")
	fmt.Println()
	fmt.Println("29. Add conversation settings:")
	fmt.Println("    ALTER TABLE conversations_name1 ADD COLUMN settings TEXT DEFAULT '{}';")
	fmt.Println("    ALTER TABLE conversations_name1 ADD COLUMN archived INTEGER DEFAULT 0;")
	fmt.Println()
	fmt.Println("30. Rename a column (SQLite 3.25.0+):")
	fmt.Println("    ALTER TABLE conversations_name1 RENAME COLUMN name TO title;")
	fmt.Println()
	fmt.Println("31. Change column data type (requires table recreation):")
	fmt.Println("    -- Example: Change TEXT column to INTEGER")
	fmt.Println("    -- Step 1: Create new table with desired schema")
	fmt.Println("    CREATE TABLE conversations_name1_new (")
	fmt.Println("      id TEXT PRIMARY KEY,")
	fmt.Println("      server_id TEXT NOT NULL,")
	fmt.Println("      priority INTEGER,  -- Changed from TEXT to INTEGER")
	fmt.Println("      created_at TEXT NOT NULL")
	fmt.Println("    );")
	fmt.Println("    -- Step 2: Copy data (with type conversion)")
	fmt.Println("    INSERT INTO conversations_name1_new")
	fmt.Println("    SELECT id, server_id, CAST(priority AS INTEGER), created_at")
	fmt.Println("    FROM conversations_name1 WHERE server_id = ?;")
	fmt.Println("    -- Step 3: Drop old table")
	fmt.Println("    DROP TABLE conversations_name1;")
	fmt.Println("    -- Step 4: Rename new table")
	fmt.Println("    ALTER TABLE conversations_name1_new RENAME TO conversations_name1;")
	fmt.Println()
	fmt.Println("    Note: Temporary tables ending with _new, _old, _temp, _backup are allowed")
	fmt.Println("          for schema migrations.")
	fmt.Println()
	fmt.Println("32. Create indexes for custom columns:")
	fmt.Println("    CREATE INDEX idx_conversations_name1_archived ON conversations_name1(archived);")
	fmt.Println("    CREATE INDEX idx_messages_name1_reply_to ON messages_name1(reply_to_id);")
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
