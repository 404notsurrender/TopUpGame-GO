#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color
BLUE='\033[0;34m'

# Base URL
BASE_URL="http://localhost:8080"
TOKEN=""

echo -e "${BLUE}Starting API Tests...${NC}\n"

# Function to test an endpoint
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local auth=$4
    local expected_status=$5

    echo -e "\n${BLUE}Testing $method $endpoint${NC}"
    
    headers="-H 'Content-Type: application/json'"
    if [ "$auth" = "true" ] && [ ! -z "$TOKEN" ]; then
        headers="$headers -H 'Authorization: Bearer $TOKEN'"
    fi

    if [ ! -z "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X $method $headers -d "$data" $BASE_URL$endpoint)
    else
        response=$(curl -s -w "\n%{http_code}" -X $method $headers $BASE_URL$endpoint)
    fi

    status_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed \$d)

    if [ "$status_code" -eq "$expected_status" ]; then
        echo -e "${GREEN}✓ Success ($status_code)${NC}"
    else
        echo -e "${RED}✗ Failed (Expected: $expected_status, Got: $status_code)${NC}"
        echo "Response: $body"
    fi
}

# 1. Test Authentication
echo -e "\n${BLUE}=== Authentication Tests ===${NC}"

# Register Admin
test_endpoint "POST" "/api/auth/register" \
    '{"email":"admin@test.com","password":"admin123","role":"admin"}' \
    "false" 201

# Register Reseller
test_endpoint "POST" "/api/auth/register" \
    '{"email":"reseller@test.com","password":"reseller123","role":"reseller"}' \
    "false" 201

# Login Admin
response=$(curl -s -X POST -H "Content-Type: application/json" \
    -d '{"email":"admin@test.com","password":"admin123"}' \
    $BASE_URL/api/auth/login)
TOKEN=$(echo $response | jq -r '.token')

if [ ! -z "$TOKEN" ]; then
    echo -e "${GREEN}✓ Login successful${NC}"
else
    echo -e "${RED}✗ Login failed${NC}"
    exit 1
fi

# 2. Test Product Management
echo -e "\n${BLUE}=== Product Management Tests ===${NC}"

# Create Product
test_endpoint "POST" "/api/admin/products" \
    '{"name":"Mobile Legends Diamonds","category":"MLBB","price":50000,"sku":"MLBB-100","description":"100 Diamonds"}' \
    "true" 201

# List Products
test_endpoint "GET" "/api/products" "" "false" 200

# Get Single Product
test_endpoint "GET" "/api/products/1" "" "false" 200

# Update Product
test_endpoint "PUT" "/api/admin/products/1" \
    '{"name":"Mobile Legends Diamonds","category":"MLBB","price":55000,"sku":"MLBB-100","description":"100 Diamonds"}' \
    "true" 200

# 3. Test Transaction Flow
echo -e "\n${BLUE}=== Transaction Tests ===${NC}"

# Create Transaction
test_endpoint "POST" "/api/checkout" \
    '{"product_id":1,"game_id":"12345","game_server":"1001","method":"bank_transfer"}' \
    "false" 201

# Check Transaction Status
test_endpoint "GET" "/api/transaction/INV-1" "" "false" 200

# List Transactions (Admin)
test_endpoint "GET" "/api/admin/transactions" "" "true" 200

# 4. Test Error Cases
echo -e "\n${BLUE}=== Error Cases Tests ===${NC}"

# Invalid Login
test_endpoint "POST" "/api/auth/login" \
    '{"email":"wrong@test.com","password":"wrong123"}' \
    "false" 401

# Access Protected Route Without Token
test_endpoint "GET" "/api/admin/products" "" "false" 401

# Invalid Product Data
test_endpoint "POST" "/api/admin/products" \
    '{"name":"","category":"","price":-1}' \
    "true" 400

# 5. Test VIP Reseller Integration
echo -e "\n${BLUE}=== VIP Reseller Integration Tests ===${NC}"

# Sync Products
test_endpoint "POST" "/api/admin/products/sync" "" "true" 200

echo -e "\n${BLUE}Testing Complete!${NC}"
