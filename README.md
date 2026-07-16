# Go Microservices Template

Monorepo em Go com containers independentes para Gateway, Auth, Catalog, Orders e Notification Worker.

## Incluído

- API Gateway com reverse proxy e validação JWT.
- Auth Service com cadastro, login, access token, refresh token rotativo e logout.
- Refresh token armazenado como hash no Redis.
- Catalog Service com CRUD de produtos.
- Order Service com criação e consulta de pedidos.
- RabbitMQ com exchanges `topic`, filas duráveis, mensagens persistentes, publisher confirms, ACK manual e dead-letter queue (DLQ).
- Transactional Outbox nos serviços Catalog e Orders.
- PostgreSQL com banco lógico por serviço.
- GORM, migrations com Goose, logs Zap, request/correlation ID e graceful shutdown.
- Um Dockerfile por aplicação.
- `go.work` para desenvolvimento do monorepo.

## Estrutura

```text
apps/
  gateway/
  auth-service/
  catalog-service/
  order-service/
  notification-worker/
pkg/platform/
deployments/
docker-compose.yml
go.work
```

## Primeira execução

```bash
cp .env.example .env
```

Troque os segredos JWT no `.env`. Em seguida:

```bash
make up
make migrate-up
```

Acompanhe os logs:

```bash
make logs
```

Serviços:

- Gateway: `http://localhost:8080`
- RabbitMQ Management: `http://localhost:15672` (`app` / `app_password`)
- PostgreSQL: `localhost:5432`
- Redis: `localhost:6379`

Apenas o Gateway deve ser publicado em produção. Os serviços internos permanecem na rede privada.

## Dependências Go

Comando unico para instalar as dependencias de todos os services:

```bash
Get-ChildItem -Path .\apps -Recurse -Filter go.mod | ForEach-Object {
    Write-Host "Instalando dependências em $($_.DirectoryName)"
    Push-Location $_.DirectoryName
    go mod tidy
    Pop-Location
}
```

Os módulos são armazenados no cache global do Go.

## Fluxo de autenticação

```text
Cliente -> Gateway -> Auth Service -> PostgreSQL + Redis
```

Registro:

```bash
curl -X POST http://localhost:8080/api/v1/auth/register   -H 'Content-Type: application/json'   -d '{"name":"Cairon","email":"cairon@example.com","password":"SenhaForte123"}'
```

Login:

```bash
curl -X POST http://localhost:8080/api/v1/auth/login   -H 'Content-Type: application/json'   -d '{"email":"cairon@example.com","password":"SenhaForte123"}'
```

Refresh:

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh   -H 'Content-Type: application/json'   -d '{"refreshToken":"SEU_REFRESH_TOKEN"}'
```

## CRUD de produtos

Todas as rotas abaixo exigem:

```http
Authorization: Bearer ACCESS_TOKEN
```

Criar:

```bash
curl -X POST http://localhost:8080/api/v1/products   -H 'Authorization: Bearer SEU_ACCESS_TOKEN'   -H 'Content-Type: application/json'   -d '{"sku":"SKU-001","name":"Produto A","description":"Exemplo","price":99.90,"stock":10}'
```

Rotas:

```text
POST   /api/v1/products
GET    /api/v1/products
GET    /api/v1/products/{id}
PUT    /api/v1/products/{id}
DELETE /api/v1/products/{id}
```

## Pedidos

```bash
curl -X POST http://localhost:8080/api/v1/orders   -H 'Authorization: Bearer SEU_ACCESS_TOKEN'   -H 'Content-Type: application/json'   -d '{"customerId":"UUID_DO_CLIENTE","items":[{"productId":"UUID_DO_PRODUTO","quantity":2}]}'
```

Ao criar um pedido, o Order Service salva `orders`, `order_items` e `outbox_events` na mesma transação. Um worker interno publica `order.created` no RabbitMQ. O Notification Worker consome esse evento.

## Criando outro serviço

1. Copie um serviço existente para `apps/seu-service`.
2. Troque o `module` no `go.mod`.
3. Adicione o módulo ao `go.work`.
4. Crie `cmd/api`, `internal` e `migrations`.
5. Adicione Dockerfile e serviço no Compose.
6. Registre a rota no Gateway.

Dentro de cada domínio, mantenha:

```text
internal/<dominio>/model.go
internal/<dominio>/service.go
internal/<dominio>/handler.go
```

Para sistemas maiores, você pode separar `repository.go`, `dto.go`, `errors.go` e `routes.go`.

## Mensageria

Exchanges existentes:

```text
catalog.events
order.events
```

Routing keys:

```text
product.created
product.updated
order.created
```

Cada consumidor deve possuir sua própria fila. Serviços diferentes não devem compartilhar a mesma fila quando todos precisam receber uma cópia do evento.

## Produção

Antes de produção:

- use secrets externos em vez de `.env` versionado;
- configure TLS no Gateway, PostgreSQL, Redis e RabbitMQ;
- adicione métricas OpenTelemetry/Prometheus;
- implemente reconexão supervisionada do RabbitMQ;
- adicione idempotência por `event.id` nos consumidores;
- use readiness/liveness probes no Kubernetes;
- restrinja os serviços internos por NetworkPolicy;
- acrescente testes unitários, integração e contrato;
- configure limites de CPU/memória e timeouts;
- não exponha PostgreSQL, Redis ou RabbitMQ publicamente.

## Observação sobre o nome do módulo

O caminho `github.com/cairon-henrique-60/goMicrosserviceTemplateAPI` é apenas o identificador do módulo. Troque pelo endereço real do seu repositório:

```bash
grep -RIl 'github.com/cairon-henrique-60/goMicrosserviceTemplateAPI' .   | xargs sed -i 's#github.com/cairon-henrique-60/goMicrosserviceTemplateAPI#SEU_MODULO#g'
```

Depois:

```bash
make tidy
```
