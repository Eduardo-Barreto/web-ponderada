#!/bin/bash

# Cores para saída no terminal
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# URL base
BASE_URL="http://localhost:8000"

# Função para exibir mensagem de sucesso/erro
check_response() {
  local status=$1
  local message=$2
  if [ $status -ge 200 ] && [ $status -lt 300 ]; then
    echo -e "${GREEN}✅ $message - Success ($status)${NC}"
    return 0
  else
    echo -e "${RED}❌ $message - Failed ($status)${NC}"
    return 1
  fi
}

# Função para testar rotas
test_route() {
  local method=$1
  local endpoint=$2
  local data=$3
  local auth_header=$4
  local message=$5
  
  echo -e "\n${YELLOW}Testing: $message${NC}"
  
  # Construindo os headers
  local headers=""
  if [ ! -z "$auth_header" ]; then
    headers="-H 'Authorization: Bearer $auth_header'"
  fi
  
  # Executando a requisição e salvando status e resposta
  local response
  if [ "$method" == "GET" ]; then
    response=$(eval "curl -s -w '\n%{http_code}' $headers $BASE_URL$endpoint")
  else
    # Ensure data is properly quoted for the curl command
    response=$(eval "curl -s -w '\n%{http_code}' -X $method $headers -H 'Content-Type: application/json' -d '$data' $BASE_URL$endpoint")
  fi
  
  # Extraindo status code e body
  status=$(echo "$response" | tail -n1)
  body=$(echo "$response" | sed '$d')
  
  # Debug para verificar o formato da resposta (opcional)
  # echo "DEBUG - Raw response: $response"
  # echo "DEBUG - Status: $status"
  # echo "DEBUG - Body: $body"
  
  # Verificando resposta
  check_response $status "$message"
  
  # Retornando o body para possível uso posterior
  echo "$body"
}

echo -e "${YELLOW}===== Teste de Saúde da API =====${NC}"
test_route "GET" "/health" "" "" "Health Check"

echo -e "\n${YELLOW}===== Criando Usuários =====${NC}"
# Criando primeiro usuário (admin)
admin_response=$(test_route "POST" "/api/v1/auth/register" '{"name":"Admin User","email":"admin@example.com","password":"password123"}' "" "Register Admin User")
admin_id=$(echo "$admin_response" | jq -r '.user.id // empty')

# Login como admin
login_response=$(test_route "POST" "/api/v1/auth/login" '{"email":"admin@example.com","password":"password123"}' "" "Login Admin User")
# Debugging the login response for troubleshooting
echo "Login response JSON: $login_response"
admin_token=$(echo "$login_response" | jq -r '.token // empty')

# Verificar se token foi obtido
if [ -z "$admin_token" ]; then
  echo -e "${RED}❌ Failed to get admin token${NC}"
  exit 1
fi

echo -e "\n${YELLOW}===== Criando Produtos =====${NC}"
# Criar produtos (necessita autenticação)
product1_response=$(test_route "POST" "/api/v1/products" '{"description":"Latest smartphone with amazing features","value":999.99,"quantity":50}' "$admin_token" "Create Product 1")
product1_id=$(echo "$product1_response" | jq -r '.id // empty')

product2_response=$(test_route "POST" "/api/v1/products" '{"description":"High-performance laptop for professionals","value":1499.99,"quantity":25}' "$admin_token" "Create Product 2")
product2_id=$(echo "$product2_response" | jq -r '.id // empty')

product3_response=$(test_route "POST" "/api/v1/products" '{"description":"Premium sound quality with noise cancellation","value":199.99,"quantity":100}' "$admin_token" "Create Product 3")
product3_id=$(echo "$product3_response" | jq -r '.id // empty')

echo -e "\n${YELLOW}===== Listando Produtos (Público) =====${NC}"
# Listar todos os produtos (público)
test_route "GET" "/api/v1/products" "" "" "List All Products"

# Se temos ID do produto 1, testar busca por ID
if [ ! -z "$product1_id" ]; then
  echo -e "\n${YELLOW}===== Buscando Produto por ID =====${NC}"
  test_route "GET" "/api/v1/products/$product1_id" "" "" "Get Product by ID"
fi

echo -e "\n${YELLOW}===== Modificando Produto =====${NC}"
# Atualizar produto 1 se temos ID
if [ ! -z "$product1_id" ]; then
  test_route "PUT" "/api/v1/products/$product1_id" '{"description":"Latest smartphone with amazing features (Updated)","value":899.99,"quantity":45}' "$admin_token" "Update Product"
fi

echo -e "\n${YELLOW}===== Listando Usuários (Requer autenticação) =====${NC}"
# Listar todos os usuários (com autenticação)
test_route "GET" "/api/v1/users" "" "$admin_token" "List All Users"

echo -e "\n${YELLOW}===== Testes Concluídos =====${NC}"
