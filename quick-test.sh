#!/bin/bash
# Quick Test Script for Flux Relay CLI
# Run this to quickly verify all features work

echo "═══════════════════════════════════════════════════════════════"
echo "Flux Relay CLI - Quick Test Script"
echo "═══════════════════════════════════════════════════════════════"
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
PASSED=0
FAILED=0

test_command() {
    local test_name="$1"
    local command="$2"
    local expected="$3"
    
    echo -n "Testing: $test_name... "
    
    if eval "$command" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ PASSED${NC}"
        ((PASSED++))
        return 0
    else
        echo -e "${RED}✗ FAILED${NC}"
        ((FAILED++))
        return 1
    fi
}

echo "1. Testing Authentication..."
test_command "Login status" "flux-relay login --help"

echo ""
echo "2. Testing Project Management..."
test_command "List projects" "flux-relay pr list"
test_command "Show project help" "flux-relay pr --help"

echo ""
echo "3. Testing Server Management..."
test_command "List servers" "flux-relay server list"
test_command "Server alias (srv)" "flux-relay srv list"
test_command "Show server help" "flux-relay server --help"

echo ""
echo "4. Testing Nameserver Management..."
test_command "List nameservers" "flux-relay ns list"
test_command "Show nameserver help" "flux-relay ns --help"

echo ""
echo "5. Testing SQL Command..."
test_command "SQL help" "flux-relay sql --help"

echo ""
echo "═══════════════════════════════════════════════════════════════"
echo "Test Results:"
echo "═══════════════════════════════════════════════════════════════"
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All basic tests passed!${NC}"
    echo ""
    echo "For interactive shell tests, see TEST_GUIDE.md"
    exit 0
else
    echo -e "${YELLOW}Some tests failed. Check your installation.${NC}"
    echo "Make sure:"
    echo "  1. CLI is built: go build -o flux-relay.exe ."
    echo "  2. You're logged in: flux-relay login"
    echo "  3. You have projects/servers set up"
    exit 1
fi
