# Go API

API REST de produtos desenvolvida em Go com Gin, PostgreSQL e Docker
Compose.

## Tecnologias

- Go 1.25
- Gin
- PostgreSQL 15.3
- Docker e Docker Compose
- `github.com/lib/pq`
- `github.com/Masterminds/squirrel`
- `swaggo/gin-swagger`

## Requisitos

- Go instalado para executar localmente
- Docker Desktop em execução para usar o ambiente completo

## Executando com Docker

Na raiz do projeto, execute:

```powershell
docker compose up -d --build
```

O `--build` é necessário após alterar o código da API; sem ele, o
container pode continuar executando uma imagem antiga sem a rota do Swagger.

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
Não use `go run main.go` na raiz: o ponto de entrada está em `cmd\main.go`.

Se a API já estiver em execução em um container, recrie-o para carregar o
binário atualizado:

```powershell
docker compose up -d --build --force-recreate go-app
```

## Endpoints

### Swagger

Com a API em execução, acesse a documentação interativa em:

`http://localhost:8000/swagger/index.html`

Para regenerar os arquivos da documentação após alterar as anotações:

```powershell
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g .\cmd\main.go -o .\docs
```

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

O preço é enviado como número e retornado como texto com duas casas
decimais, por exemplo: `"price": "149.90"`.

### Atualizar produto

```http
PUT /products/:productId
Content-Type: application/json
```

Corpo:

```json
{
  "name": "Teclado mecânico",
  "price": 249.90
}
```

Resposta: `200 OK` com o produto atualizado.

Para atualização parcial, use `PATCH`. É possível enviar somente os campos
que devem ser alterados:

```http
PATCH /products/:productId
Content-Type: application/json
```

```json
{
  "price": 80.50
}
```

O campo `name` também pode ser enviado isoladamente. O corpo não pode ser
vazio.

Resposta:

```json
{
  "id_product": 1,
  "name": "Teclado mecânico",
  "price": "80.50"
}
```

O preço é retornado como texto com duas casas decimais para preservar o
formato monetário. Na requisição, ele continua sendo enviado como número
JSON, usando ponto como separador decimal.

### Excluir produto

```http
DELETE /products/:productId
```

Resposta: `204 No Content`.

## Próximas atividades

- [x] Finalizar o CRUD de produtos, adicionando operações de atualização
  (`PUT`/`PATCH`) e exclusão (`DELETE`).
- [x] Adicionar testes unitários, de integração e dos endpoints da API.
- [x] Refatorar a arquitetura para melhorar a separação de responsabilidades,
  a testabilidade e a manutenção do projeto.
- [x] Documentar e disponibilizar as rotas na web usando Swagger/OpenAPI,
  incluindo exemplos de requisições e respostas.

## Testes

Execute os testes e verifique a compilação com:

```powershell
go test ./...
```

Para gerar o relatório de cobertura dos pacotes com testes:

```powershell
go test -coverprofile=coverage.out -covermode=atomic ./cmd ./controller ./repository ./response ./service ./usecase ./db
go tool cover -func=coverage.out
```

O perfil atual cobre 100% das instruções dos pacotes da aplicação:
`cmd`, `controller`, `db`, `repository`, `response`, `service` e `usecase`.

Os endpoints recebem DTOs validados e retornam DTOs específicos da API. O ID
do produto é controlado pela rota e não pode ser definido pelo corpo da
requisição.

## Estrutura do projeto

```text
.
├── cmd/             # Inicialização da aplicação e rotas
├── controller/      # Handlers HTTP
├── db/              # Conexão com o PostgreSQL
├── docs/            # Documentação Swagger gerada
├── dto/             # DTOs de entrada e validação dos payloads
├── model/           # Modelos de domínio
├── repository/      # Acesso aos dados
├── response/        # DTOs das respostas HTTP
├── service/         # Orquestração dos serviços da aplicação
├── usecase/         # Regras de negócio
├── Dockerfile       # Build da imagem da API
└── docker-compose.yml
```
