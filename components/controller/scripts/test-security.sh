#!/bin/bash

# Security Features Test Script for IDEKube Controller
# This script tests various security features implemented in the controller

set -e

BASE_URL="${BASE_URL:-http://localhost:8080}"
COLOR_GREEN='\033[0;32m'
COLOR_RED='\033[0;31m'
COLOR_YELLOW='\033[1;33m'
COLOR_NC='\033[0m' # No Color

echo "=========================================="
echo "IDEKube Controller Security Tests"
echo "=========================================="
echo "Base URL: $BASE_URL"
echo ""

# Test 1: Security Headers
echo -e "${COLOR_YELLOW}Test 1: Security Headers${COLOR_NC}"
HEADERS=$(curl -s -I "$BASE_URL/health" 2>&1)
echo "Checking security headers..."

check_header() {
    local header=$1
    if echo "$HEADERS" | grep -qi "$header"; then
        echo -e "  ${COLOR_GREEN}✓${COLOR_NC} $header found"
        return 0
    else
        echo -e "  ${COLOR_RED}✗${COLOR_NC} $header missing"
        return 1
    fi
}

check_header "X-Frame-Options"
check_header "X-XSS-Protection"
check_header "X-Content-Type-Options"
check_header "Strict-Transport-Security"
check_header "Content-Security-Policy"
echo ""

# Test 2: Malicious Route Interception
echo -e "${COLOR_YELLOW}Test 2: Malicious Route Interception${COLOR_NC}"
echo "Testing blocked routes..."

test_blocked_route() {
    local route=$1
    local response=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL$route")
    if [ "$response" = "404" ]; then
        echo -e "  ${COLOR_GREEN}✓${COLOR_NC} $route blocked (404)"
        return 0
    else
        echo -e "  ${COLOR_RED}✗${COLOR_NC} $route returned $response (expected 404)"
        return 1
    fi
}

test_blocked_route "/.env"
test_blocked_route "/wp-admin"
test_blocked_route "/.git/config"
test_blocked_route "/phpmyadmin"
test_blocked_route "/admin"
echo ""

# Test 3: Rate Limiting
echo -e "${COLOR_YELLOW}Test 3: Rate Limiting${COLOR_NC}"
echo "Sending rapid requests to trigger rate limiting..."
echo "This may take a minute..."

success_count=0
rate_limited_count=0

for i in {1..110}; do
    response=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health")
    if [ "$response" = "200" ]; then
        ((success_count++))
    elif [ "$response" = "429" ]; then
        ((rate_limited_count++))
    fi
    sleep 0.05
done

echo "  Successful requests: $success_count"
echo "  Rate limited (429): $rate_limited_count"

if [ $rate_limited_count -gt 0 ]; then
    echo -e "  ${COLOR_GREEN}✓${COLOR_NC} Rate limiting is working"
else
    echo -e "  ${COLOR_RED}✗${COLOR_NC} Rate limiting may not be working properly"
fi
echo ""

# Test 4: Request Size Limit
echo -e "${COLOR_YELLOW}Test 4: Request Size Limit${COLOR_NC}"
echo "Testing request size limit (this may fail if endpoint requires auth)..."

# Create a large payload (11MB)
large_payload=$(dd if=/dev/zero bs=1M count=11 2>/dev/null | base64)

response=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST "$BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"data\":\"$large_payload\"}" 2>/dev/null || echo "error")

if [ "$response" = "error" ] || [ "$response" = "413" ] || [ "$response" = "400" ]; then
    echo -e "  ${COLOR_GREEN}✓${COLOR_NC} Large request blocked (response: $response)"
else
    echo -e "  ${COLOR_YELLOW}⚠${COLOR_NC} Request returned $response (may be handled by other middleware)"
fi
echo ""

# Test 5: Authentication Timing
echo -e "${COLOR_YELLOW}Test 5: Authentication Failure Delay${COLOR_NC}"
echo "Testing authentication failure timing..."

# Measure response time for failed auth
start=$(date +%s%N)
curl -s -o /dev/null -w "%{http_code}" \
    -X POST "$BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"invalid","password":"invalid"}' > /dev/null
end=$(date +%s%N)

elapsed_ms=$(( (end - start) / 1000000 ))

echo "  Failed auth response time: ${elapsed_ms}ms"

if [ $elapsed_ms -ge 100 ] && [ $elapsed_ms -le 600 ]; then
    echo -e "  ${COLOR_GREEN}✓${COLOR_NC} Random delay appears to be working (100-600ms range)"
else
    echo -e "  ${COLOR_YELLOW}⚠${COLOR_NC} Delay is $elapsed_ms ms (expected 100-600ms range)"
fi
echo ""

# Test 6: Server Timeout Configuration
echo -e "${COLOR_YELLOW}Test 6: Server Configuration${COLOR_NC}"
echo "Checking if server is responding..."

response=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health")
if [ "$response" = "200" ]; then
    echo -e "  ${COLOR_GREEN}✓${COLOR_NC} Server is responding correctly"
else
    echo -e "  ${COLOR_RED}✗${COLOR_NC} Server returned $response"
fi
echo ""

# Summary
echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo "All basic security features have been tested."
echo "Review the results above for any failures."
echo ""
echo "Note: Some tests may show warnings if the server"
echo "is not running or if endpoints require authentication."
echo "=========================================="
