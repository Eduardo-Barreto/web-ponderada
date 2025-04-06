# Sistema de Gestão Web

Um sistema web modularizado para gestão de produtos e usuários.

## Tecnologias Utilizadas

- Frontend: HTML, CSS e HTMX
- Backend: Go com Gin Framework
- Banco de Dados: PostgreSQL
- Servidor Web: Nginx
- Containers: Docker e Docker Compose

## Como Executar

### Pré-requisitos

- Docker
- Docker Compose

### Passos para Execução

1. Clone o repositório
2. No diretório raiz do projeto, execute:

```bash
docker compose up --build
```

Isso irá:
- Construir e iniciar todos os containers
- Criar o banco de dados e suas tabelas
- Configurar o Nginx como proxy reverso
- Iniciar a aplicação web

3. Acesse a aplicação:
- Frontend: http://localhost:80
- Backend API: http://localhost:8000/api/v1

## Estrutura do Projeto

```
/
├── compose.yaml          # Configuração Docker Compose
├── nginx/               # Configurações do Nginx
│   └── default.conf    
└── src/
    ├── backend/        # API em Go
    └── frontend/       # Interface web
```

## Funcionalidades

- Cadastro e autenticação de usuários
- Gerenciamento de produtos (CRUD)
- Upload de imagens
- Interface responsiva e intuitiva

## Segurança

- Autenticação JWT
- CORS configurado
- Proxy reverso com Nginx
- Rede Docker isolada
- Senhas criptografadas

## API Endpoints

### Autenticação
- POST /api/v1/auth/register - Registro de usuário
- POST /api/v1/auth/login - Login de usuário

### Produtos
- GET /api/v1/products - Lista todos os produtos
- GET /api/v1/products/:id - Obtém um produto específico
- POST /api/v1/products - Cria um novo produto
- PUT /api/v1/products/:id - Atualiza um produto
- DELETE /api/v1/products/:id - Remove um produto

### Usuários (requer autenticação)
- GET /api/v1/users - Lista todos os usuários
- GET /api/v1/users/:id - Obtém um usuário específico
- PUT /api/v1/users/:id - Atualiza um usuário
- DELETE /api/v1/users/:id - Remove um usuário
