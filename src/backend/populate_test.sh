#!/bin/bash

# Cores para saída no terminal
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# URL base
BASE_URL="http://localhost:8000"

# Create a dummy file for image uploads
DUMMY_IMAGE="dummy_image.png"
echo "This is a dummy image file for testing." > $DUMMY_IMAGE
echo -e "${YELLOW}Created dummy file: $DUMMY_IMAGE${NC}" >&2

# Cleanup dummy file on exit
trap 'rm -f $DUMMY_IMAGE' EXIT

# Function to display message (stderr)
print_message() {
  local color=$1
  local symbol=$2
  local message=$3
  local status_code=$4
  if [ -z "$status_code" ]; then
      echo -e "${color}${symbol} $message${NC}" >&2
  else
      echo -e "${color}${symbol} $message - ($status_code)${NC}" >&2
  fi
}

# Function to check HTTP status and log body on error (stderr)
check_status_and_log_error() {
  local status=$1
  local body=$2 # Pass body for logging
  local message=$3
  if [ $status -ge 200 ] && [ $status -lt 300 ]; then
    print_message "$GREEN" "✅" "$message" "$status"
    return 0
  else
    print_message "$RED" "❌" "$message" "$status"
    # Log the error body if it's not empty
    if [ ! -z "$body" ]; then
        echo -e "${RED}Error Body:${NC} $body" >&2
    fi
    return 1
  fi
}

# Function to test JSON routes (e.g., auth, users)
# Prints informational messages to stderr
# Prints ONLY the response body to stdout
# Logs error body to stderr
test_json_route() {
  local method=$1
  local endpoint=$2
  local data=$3
  local auth_header=$4
  local message=$5

  print_message "$YELLOW" "Testing:" "$message"

  # Build headers
  local headers="-H 'Content-Type: application/json'"
  if [ ! -z "$auth_header" ]; then
    headers="$headers -H 'Authorization: Bearer $auth_header'"
  fi

  # Execute request and save full output (body + status code)
  local full_response
  if [ "$method" == "GET" ]; then
    full_response=$(eval "curl -s -w '\\n%{http_code}' $headers $BASE_URL$endpoint")
  else
    full_response=$(eval "curl -s -w '\\n%{http_code}' -X $method $headers -d '$data' $BASE_URL$endpoint")
  fi

  # Extract status code (last line)
  local status=$(echo "$full_response" | tail -n1)
  # Extract body (everything except the last line)
  local body=$(echo "$full_response" | sed '$d')

  # Check response status and log error body (prints to stderr)
  check_status_and_log_error "$status" "$body" "$message"
  local check_res=$? # Capture return status

  # Print the body to stdout for capture by caller (even on error, might be useful)
  echo "$body"

  # Return the success/failure status
  return $check_res
}


# --- Main Script Logic ---

echo -e "${YELLOW}===== Teste de Saúde da API =====${NC}" >&2
test_json_route "GET" "/health" "" "" "Health Check"

echo -e "\n${YELLOW}===== Criando Usuários =====${NC}" >&2
# Criando primeiro usuário (admin)
admin_response=$(test_json_route "POST" "/api/v1/auth/register" '{"name":"Admin User","email":"admin@example.com","password":"password123"}' "" "Register Admin User")
admin_id=$(echo "$admin_response" | jq -r '.user.id // empty')

# Login como admin
login_response=$(test_json_route "POST" "/api/v1/auth/login" '{"email":"admin@example.com","password":"password123"}' "" "Login Admin User")
login_status=$?
admin_token=""

if [ $login_status -eq 0 ]; then
  admin_token=$(echo "$login_response" | jq -r '.token // empty')
fi

if [ -z "$admin_token" ]; then
  print_message "$RED" "❌" "Failed to get admin token"
  exit 1
else
   print_message "$GREEN" "✅" "Successfully obtained admin token"
fi


echo -e "\n${YELLOW}===== Criando Produtos (Multipart Form) =====${NC}" >&2
# --- Create Product 1 ---
print_message "$YELLOW" "Testing:" "Create Product 1"
prod1_full_response=$(curl -s -w '\n%{http_code}' \
  -X POST \
  -H "Authorization: Bearer $admin_token" \
  -F "description=Latest smartphone with amazing features" \
  -F "value=999.99" \
  -F "quantity=50" \
  -F "image=@$DUMMY_IMAGE" \
  "$BASE_URL/api/v1/products")
prod1_status=$(echo "$prod1_full_response" | tail -n1)
prod1_body=$(echo "$prod1_full_response" | sed '$d')
check_status_and_log_error "$prod1_status" "$prod1_body" "Create Product 1"
product1_id=$(echo "$prod1_body" | jq -r '.id // empty') # Attempt extract even on error

# --- Create Product 2 ---
print_message "$YELLOW" "Testing:" "Create Product 2"
prod2_full_response=$(curl -s -w '\n%{http_code}' \
  -X POST \
  -H "Authorization: Bearer $admin_token" \
  -F "description=High-performance laptop for professionals" \
  -F "value=1499.99" \
  -F "quantity=25" \
  -F "image=@$DUMMY_IMAGE" \
  "$BASE_URL/api/v1/products")
prod2_status=$(echo "$prod2_full_response" | tail -n1)
prod2_body=$(echo "$prod2_full_response" | sed '$d')
check_status_and_log_error "$prod2_status" "$prod2_body" "Create Product 2"
product2_id=$(echo "$prod2_body" | jq -r '.id // empty')

# --- Create Product 3 ---
print_message "$YELLOW" "Testing:" "Create Product 3"
prod3_full_response=$(curl -s -w '\n%{http_code}' \
  -X POST \
  -H "Authorization: Bearer $admin_token" \
  -F "description=Premium sound quality with noise cancellation" \
  -F "value=199.99" \
  -F "quantity=100" \
  -F "image=@$DUMMY_IMAGE" \
  "$BASE_URL/api/v1/products")
prod3_status=$(echo "$prod3_full_response" | tail -n1)
prod3_body=$(echo "$prod3_full_response" | sed '$d')
check_status_and_log_error "$prod3_status" "$prod3_body" "Create Product 3"
product3_id=$(echo "$prod3_body" | jq -r '.id // empty')


echo -e "\n${YELLOW}===== Listando Produtos (Público) =====${NC}" >&2
# Listar todos os produtos (público) - Uses JSON GET
test_json_route "GET" "/api/v1/products" "" "" "List All Products"

# Se temos ID do produto 1, testar busca por ID
if [ ! -z "$product1_id" ]; then
  echo -e "\n${YELLOW}===== Buscando Produto por ID =====${NC}" >&2
  test_json_route "GET" "/api/v1/products/$product1_id" "" "" "Get Product by ID"
else
   print_message "$YELLOW" "Skipping" "'Get Product by ID' test because Product 1 ID was not extracted."
fi


echo -e "\n${YELLOW}===== Modificando Produto (Multipart Form) =====${NC}" >&2
# Atualizar produto 1 se temos ID
if [ ! -z "$product1_id" ]; then
  print_message "$YELLOW" "Testing:" "Update Product 1"
  update_full_response=$(curl -s -w '\n%{http_code}' \
    -X PUT \
    -H "Authorization: Bearer $admin_token" \
    -F "description=Latest smartphone with amazing features (Updated)" \
    -F "value=899.99" \
    -F "quantity=45" \
    -F "image=@$DUMMY_IMAGE" \
    "$BASE_URL/api/v1/products/$product1_id")
  update_status=$(echo "$update_full_response" | tail -n1)
  update_body=$(echo "$update_full_response" | sed '$d')
  check_status_and_log_error "$update_status" "$update_body" "Update Product 1"
else
   print_message "$YELLOW" "Skipping" "'Update Product' test because Product 1 ID was not extracted."
fi


echo -e "\n${YELLOW}===== Listando Usuários (Requer autenticação) =====${NC}" >&2
# Listar todos os usuários (com autenticação) - Uses JSON GET
test_json_route "GET" "/api/v1/users" "" "$admin_token" "List All Users"


echo -e "\n${YELLOW}===== Testes Concluídos =====${NC}" >&2

# Note: Dummy file is removed by trap EXIT
