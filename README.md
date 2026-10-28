# Go API

API REST de produtos desenvolvida em Go com Gin, PostgreSQL e Docker
Compose.

## Tecnologias

- Go 1.25
- Gin
- PostgreSQL 15.3
- Docker e Docker Compose
- `github.com/lib/pq`

## Requisitos

- Go instalado para executar localmente
- Docker Desktop em execução para usar o ambiente completo

## Executando com Docker

Na raiz do projeto, execute:

```powershell
docker compose up -d --build
```

Verifique o status dos serviços:

```powershell
docker compose ps
```

Para acompanhar os logs:

```powershell
docker compose logs -f
```

Para parar os serviços:

```powershell
docker compose down
```

O comando `docker compose down -v` também remove o volume do PostgreSQL e
apaga os dados persistidos.

### Endereços do ambiente Docker

| Serviço | Endereço |
| --- | --- |
| API | `http://localhost:8000` |
| PostgreSQL pelo host | `localhost:5433` |
| PostgreSQL entre containers | `go_db:5432` |

Credenciais padrão do banco:

```text
Usuário: postgres
Senha: 1234
Banco: postgres
```

## Executando localmente

Inicie o PostgreSQL com Docker:

```powershell
docker compose up -d go_db
```

Depois execute a API:

```powershell
go run .\cmd\main.go
```

Ao executar a API diretamente no Windows, a conexão usa `localhost:5433`.

## Endpoints

### Verificar o serviço

```http
GET /ping
```

Resposta:

```json
{
  "message": "pong"
}
```

### Listar produtos

```http
GET /products
```

### Buscar produto por ID

```http
GET /products/:productId
```

Exemplo:

```http
GET /products/1
```

Respostas possíveis:

- `200 OK`: produto encontrado
- `400 Bad Request`: ID inválido
- `404 Not Found`: produto não encontrado

### Criar produto

```http
POST /products
Content-Type: application/json
```

Corpo:

```json
{
  "name": "Teclado",
  "price": 149.90
}
```

Resposta: `201 Created`.

## Próximas atividades

- [ ] Finalizar o CRUD de produtos, adicionando operações de atualização
  (`PUT`/`PATCH`) e exclusão (`DELETE`).
- [x] Adicionar testes unitários, de integração e dos endpoints da API.
- [ ] Refatorar a arquitetura para melhorar a separação de responsabilidades,
  a testabilidade e a manutenção do projeto.
- [ ] Documentar e disponibilizar as rotas na web usando Swagger/OpenAPI,
  incluindo exemplos de requisições e respostas.

## Testes

Execute os testes e verifique a compilação com:

```powershell
go test ./...
```

Para gerar o relatório de cobertura dos pacotes com testes:

```powershell
go test -coverprofile=coverage.out -covermode=atomic ./cmd ./controller ./repository ./usecase ./db
go tool cover -func=coverage.out
```

O perfil atual cobre 100% das instruções dos pacotes da aplicação:
`cmd`, `controller`, `db`, `repository` e `usecase`.

## Estrutura do projeto

```text
.
├── cmd/             # Inicialização da aplicação e rotas
├── controller/      # Handlers HTTP
├── db/              # Conexão com o PostgreSQL
├── model/           # Modelos e respostas da API
├── repository/      # Acesso aos dados
├── usecase/         # Regras de negócio
├── Dockerfile       # Build da imagem da API
└── docker-compose.yml
```
