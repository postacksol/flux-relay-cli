# Flux Relay CLI - Complete Test Guide

This guide provides step-by-step tests for all CLI features. Run each command in order to verify everything works.

## Prerequisites

1. **Build the CLI:**
   ```bash
   cd flux-relay-cli
   go build -o flux-relay.exe .
   ```

2. **Make sure you're logged in:**
   ```bash
   flux-relay login
   ```

---

## Test 1: Authentication ✅

```bash
# Test login status
flux-relay login

# Expected: Shows your user info or prompts to login
```

---

## Test 2: Project Management ✅

```bash
# List all projects
flux-relay pr list

# Expected: Shows list of projects with IDs and names

# Select a project by name (replace with your project name)
flux-relay pr "My Project"

# Or select by ID
flux-relay pr 56OSXXQH

# Show current project
flux-relay pr

# Expected: Shows selected project details
```

---

## Test 3: Server Management ✅

```bash
# List servers in selected project
flux-relay server list
# or
flux-relay srv list

# Expected: Shows servers with nameserver counts

# Select a server by name
flux-relay server "MyServer"
# or
flux-relay srv "MyServer"

# Or select by ID
flux-relay server server_123

# Show current server
flux-relay server

# Expected: Shows selected server details
```

---

## Test 4: One-Off SQL Query ✅

```bash
# Execute a simple query (replace nameserver name)
flux-relay sql "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' LIMIT 5"

# Expected: Returns table names
```

---

## Test 5: Interactive SQL Shell - Basic Commands ✅

```bash
# Enter server shell
flux-relay server shell "MyServer"
# or
flux-relay srv shell "MyServer"

# In the shell, test these commands:

# Test .help
.help

# Test .context
.context

# Test .nameservers
.nameservers

# Test .tables (if nameservers exist)
.tables

# Test .examples
.examples

# Test .clear
SELECT * FROM test;
.clear

# Test .quit
.quit
```

---

## Test 6: Nameserver Management ✅

```bash
# Enter server shell
flux-relay server shell "MyServer"

# Test .create_ns
.create_ns test_db

# Expected: Creates nameserver successfully

# Test .nameservers (should show new nameserver)
.nameservers

# Test .init_ns
.init_ns test_db

# Expected: Initializes schema, creates default tables

# Test .use
.use test_db

# Expected: Switches context to test_db

# Test .context (should show test_db)
.context

# Test .tables (should show tables with _test_db suffix)
.tables

# Exit shell
.quit
```

---

## Test 7: Custom Table Creation ✅

```bash
# Enter server shell
flux-relay server shell "MyServer"

# Switch to a nameserver
.use test_db

# Test .create_table helper
.create_table products

# Test creating a custom table
CREATE TABLE custom_products_test_db (
    id TEXT PRIMARY KEY,
    server_id TEXT NOT NULL,
    name TEXT NOT NULL,
    price REAL,
    created_at TEXT DEFAULT (datetime('now'))
);

# Expected: Table created successfully

# Test .tables (should show custom_products_test_db)
.tables

# Test .schema
.schema custom_products_test_db

# Expected: Shows table schema

# Exit shell
.quit
```

---

## Test 8: Schema Customization (ALTER TABLE) ✅

```bash
# Enter server shell
flux-relay server shell "MyServer"

# Switch to a nameserver
.use test_db

# Test .alter_table helper
.alter_table

# Expected: Shows customization examples

# Test adding a column to conversations
ALTER TABLE conversations_test_db ADD COLUMN priority INTEGER DEFAULT 0;

# Expected: Column added successfully

# Test adding a column to messages
ALTER TABLE messages_test_db ADD COLUMN reactions TEXT DEFAULT '[]';

# Expected: Column added successfully

# Test adding a column to end_users
ALTER TABLE end_users_test_db ADD COLUMN avatar_url TEXT;

# Expected: Column added successfully

# Test creating an index
CREATE INDEX idx_conversations_test_db_priority ON conversations_test_db(priority);

# Expected: Index created successfully

# Verify with .schema
.schema conversations_test_db

# Expected: Shows updated schema with new column

# Exit shell
.quit
```

---

## Test 9: Data Operations ✅

```bash
# Enter server shell
flux-relay server shell "MyServer"

# Switch to a nameserver
.use test_db

# Test INSERT
INSERT INTO conversations_test_db (id, server_id, project_id, participants, is_group, created_at)
VALUES ('test_conv_1', 'your_server_id', 'your_project_id', '["user1", "user2"]', 0, datetime('now'));

# Expected: Row inserted

# Test SELECT
SELECT * FROM conversations_test_db WHERE server_id = 'your_server_id' LIMIT 5;

# Expected: Returns inserted row

# Test UPDATE
UPDATE conversations_test_db
SET name = 'Updated Name'
WHERE id = 'test_conv_1' AND server_id = 'your_server_id';

# Expected: Row updated

# Test DELETE
DELETE FROM conversations_test_db WHERE id = 'test_conv_1' AND server_id = 'your_server_id';

# Expected: Row deleted

# Exit shell
.quit
```

---

## Test 10: Multi-line Queries ✅

```bash
# Enter server shell
flux-relay server shell "MyServer"

# Switch to a nameserver
.use test_db

# Test multi-line query (type this, press Enter after each line, end with semicolon)
SELECT 
    id,
    name,
    created_at
FROM conversations_test_db
WHERE server_id = 'your_server_id'
ORDER BY created_at DESC
LIMIT 10;

# Expected: Executes multi-line query successfully

# Test empty line execution (type query, press Enter twice)
SELECT COUNT(*) FROM conversations_test_db WHERE server_id = 'your_server_id'

# (Press Enter again on empty line)
# Expected: Executes query

# Exit shell
.quit
```

---

## Test 11: Error Handling ✅

```bash
# Enter server shell
flux-relay server shell "MyServer"

# Switch to a nameserver
.use test_db

# Test incomplete LIMIT
SELECT * FROM conversations_test_db WHERE server_id = ? LIMIT

# Expected: Shows error about incomplete LIMIT

# Test missing server_id
SELECT * FROM conversations_test_db;

# Expected: Shows error or hint about server_id

# Test invalid table name
SELECT * FROM nonexistent_table WHERE server_id = ?;

# Expected: Shows "no such table" error

# Test Ctrl+C (should NOT exit shell, only cancel query)
# Press Ctrl+C while typing a query
# Expected: Shows "Query cancelled" or "Use '.quit' to exit"

# Test .quit (only way to exit)
.quit

# Expected: Exits shell gracefully
```

---

## Test 12: Nameserver Shell (Direct) ✅

```bash
# Enter nameserver shell directly
flux-relay ns shell test_db

# Expected: Opens shell with test_db context

# Test .context (should show test_db)
.context

# Test .tables
.tables

# Exit shell
.quit
```

---

## Test 13: Edge Cases ✅

```bash
# Enter server shell
flux-relay server shell "MyServer"

# Test creating nameserver with spaces
.create_ns "my test db"

# Expected: Creates nameserver with spaces handled

# Test .use with nameserver that has spaces
.use "my test db"

# Expected: Switches to nameserver

# Test creating table with nameserver that has spaces
CREATE TABLE test_my_test_db (id TEXT PRIMARY KEY, server_id TEXT);

# Expected: Table created (spaces converted to underscores)

# Test .drop_table helper
.drop_table test_my_test_db

# Expected: Shows how to drop table or executes it

# Exit shell
.quit
```

---

## Test 14: Query with Quotes ✅

```bash
# Enter server shell
flux-relay server shell "MyServer"

# Switch to a nameserver
.use test_db

# Test query wrapped in quotes (should auto-strip)
sql "SELECT * FROM conversations_test_db WHERE server_id = ? LIMIT 5"

# Expected: Strips "sql" prefix and quotes, executes query

# Exit shell
.quit
```

---

## Test 15: Complete Workflow ✅

```bash
# Full end-to-end test

# 1. Login
flux-relay login

# 2. Select project
flux-relay pr "My Project"

# 3. Select server
flux-relay server "MyServer"

# 4. Enter shell
flux-relay server shell "MyServer"

# 5. Create nameserver
.create_ns e2e_test

# 6. Initialize schema
.init_ns e2e_test

# 7. Switch to it
.use e2e_test

# 8. Customize schema
ALTER TABLE conversations_e2e_test ADD COLUMN priority INTEGER DEFAULT 0;

# 9. Create custom table
CREATE TABLE custom_data_e2e_test (
    id TEXT PRIMARY KEY,
    server_id TEXT NOT NULL,
    data TEXT,
    created_at TEXT DEFAULT (datetime('now'))
);

# 10. Verify tables
.tables

# Expected: Shows both default and custom tables

# 11. Insert data
INSERT INTO conversations_e2e_test (id, server_id, project_id, participants, is_group, created_at, priority)
VALUES ('e2e_conv_1', 'your_server_id', 'your_project_id', '["user1"]', 0, datetime('now'), 1);

# 12. Query data
SELECT * FROM conversations_e2e_test WHERE server_id = 'your_server_id';

# 13. Verify custom column
SELECT id, priority FROM conversations_e2e_test WHERE server_id = 'your_server_id';

# Expected: Shows priority column

# 14. Exit
.quit
```

---

## Expected Results Summary

✅ **All tests should:**
- Execute without errors
- Show appropriate output
- Handle edge cases gracefully
- Provide helpful error messages when needed

❌ **If any test fails:**
1. Check you're logged in (`flux-relay login`)
2. Verify project/server names are correct
3. Make sure nameservers exist before testing operations on them
4. Replace placeholder values (server_id, project_id) with actual values
5. Check the error message for hints

---

## Quick Verification Checklist

- [ ] Authentication works
- [ ] Project selection works
- [ ] Server selection works
- [ ] Nameserver creation works
- [ ] Schema initialization works
- [ ] Custom table creation works
- [ ] Schema customization (ALTER TABLE) works
- [ ] Data operations (INSERT/SELECT/UPDATE/DELETE) work
- [ ] Multi-line queries work
- [ ] Error handling works
- [ ] Ctrl+C doesn't exit shell
- [ ] .quit exits shell
- [ ] All helper commands work (.help, .tables, .schema, etc.)

---

## Notes

- Replace `"MyServer"`, `"My Project"`, `test_db`, etc. with your actual names
- Replace `'your_server_id'` and `'your_project_id'` with actual IDs
- Some operations require existing data/tables to test fully
- The shell supports both semicolon-terminated and empty-line-terminated queries
