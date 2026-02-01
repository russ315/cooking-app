#!/bin/bash

# Recipe Backend API Test Script
# This script tests all authentication endpoints

BASE_URL="http://localhost:8080"
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "Recipe Backend API Test Suite"
echo "=========================================="
echo ""

# Test 1: Health Check
echo -e "${YELLOW}Test 1: Health Check${NC}"
response=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health")
if [ $response -eq 200 ]; then
    echo -e "${GREEN}✓ Health check passed${NC}"
else
    echo -e "${RED}✗ Health check failed (HTTP $response)${NC}"
fi
echo ""

# Test 2: API Info
echo -e "${YELLOW}Test 2: Get API Info${NC}"
response=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/")
if [ $response -eq 200 ]; then
    echo -e "${GREEN}✓ API info endpoint working${NC}"
else
    echo -e "${RED}✗ API info failed (HTTP $response)${NC}"
fi
echo ""

# Test 3: Register User
echo -e "${YELLOW}Test 3: Register New User${NC}"
register_response=$(curl -s -X POST "$BASE_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "TestPass123"
  }')

echo "Response: $register_response"
token=$(echo $register_response | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ ! -z "$token" ]; then
    echo -e "${GREEN}✓ User registered successfully${NC}"
    echo "Token: ${token:0:20}..."
else
    echo -e "${RED}✗ Registration failed${NC}"
fi
echo ""

# Test 4: Login User
echo -e "${YELLOW}Test 4: Login User${NC}"
login_response=$(curl -s -X POST "$BASE_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "TestPass123"
  }')

echo "Response: $login_response"
token=$(echo $login_response | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ ! -z "$token" ]; then
    echo -e "${GREEN}✓ Login successful${NC}"
    echo "Token: ${token:0:20}..."
else
    echo -e "${RED}✗ Login failed${NC}"
fi
echo ""

# Test 5: Validate Token
echo -e "${YELLOW}Test 5: Validate Token${NC}"
if [ ! -z "$token" ]; then
    validate_response=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/auth/validate" \
      -H "Authorization: Bearer $token")
    
    http_code=$(echo "$validate_response" | tail -n1)
    body=$(echo "$validate_response" | head -n-1)
    
    echo "Response: $body"
    
    if [ $http_code -eq 200 ]; then
        echo -e "${GREEN}✓ Token validation successful${NC}"
    else
        echo -e "${RED}✗ Token validation failed (HTTP $http_code)${NC}"
    fi
else
    echo -e "${RED}✗ No token available to validate${NC}"
fi
echo ""

# Test 6: Get User Profile (Protected)
echo -e "${YELLOW}Test 6: Get User Profile (Protected Endpoint)${NC}"
if [ ! -z "$token" ]; then
    profile_response=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/auth/profile" \
      -H "Authorization: Bearer $token")
    
    http_code=$(echo "$profile_response" | tail -n1)
    body=$(echo "$profile_response" | head -n-1)
    
    echo "Response: $body"
    
    if [ $http_code -eq 200 ]; then
        echo -e "${GREEN}✓ Profile retrieval successful${NC}"
    else
        echo -e "${RED}✗ Profile retrieval failed (HTTP $http_code)${NC}"
    fi
else
    echo -e "${RED}✗ No token available${NC}"
fi
echo ""

# Test 7: Invalid Email Validation
echo -e "${YELLOW}Test 7: Test Email Validation${NC}"
validation_response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "invaliduser",
    "email": "invalid-email",
    "password": "TestPass123"
  }')

http_code=$(echo "$validation_response" | tail -n1)

if [ $http_code -eq 400 ]; then
    echo -e "${GREEN}✓ Email validation working correctly${NC}"
else
    echo -e "${RED}✗ Email validation not working (HTTP $http_code)${NC}"
fi
echo ""

# Test 8: Weak Password Validation
echo -e "${YELLOW}Test 8: Test Password Strength Validation${NC}"
password_response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "weakpass",
    "email": "weak@example.com",
    "password": "123"
  }')

http_code=$(echo "$password_response" | tail -n1)

if [ $http_code -eq 400 ]; then
    echo -e "${GREEN}✓ Password validation working correctly${NC}"
else
    echo -e "${RED}✗ Password validation not working (HTTP $http_code)${NC}"
fi
echo ""

# Test 9: Wrong Password Login
echo -e "${YELLOW}Test 9: Test Wrong Password${NC}"
wrong_response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "WrongPassword123"
  }')

http_code=$(echo "$wrong_response" | tail -n1)

if [ $http_code -eq 401 ]; then
    echo -e "${GREEN}✓ Wrong password correctly rejected${NC}"
else
    echo -e "${RED}✗ Wrong password handling not working (HTTP $http_code)${NC}"
fi
echo ""

# Test 10: Protected Endpoint Without Token
echo -e "${YELLOW}Test 10: Test Protected Endpoint Without Token${NC}"
no_token_response=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/auth/profile")

if [ $no_token_response -eq 401 ]; then
    echo -e "${GREEN}✓ Protected endpoint correctly requires authentication${NC}"
else
    echo -e "${RED}✗ Protected endpoint security not working (HTTP $no_token_response)${NC}"
fi
echo ""

echo "=========================================="
echo "Test Suite Complete"
echo "=========================================="
