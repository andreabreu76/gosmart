
# GoSmart

GoSmart é uma aplicação escrita em Go que utiliza o framework Fiber para construir APIs RESTful. O projeto é integrado com a OpenAI para geração de texto e usa Redis para armazenamento de logs. O projeto é modularizado e segue boas práticas de desenvolvimento.

---

## Estrutura do Projeto

```
gosmart/
├── config/
│   └── environments.go    # Gerenciamento de variáveis de ambiente
├── docs/                  # Documentação gerada pelo Swagger
├── entities/
│   └── request.proto      # Definições Protobuf para os dados
├── handlers/
│   └── openai.go          # Handlers para as rotas da OpenAI
├── router/
│   └── router.go          # Definição das rotas do projeto
├── services/
│   ├── openai.go          # Serviço para integração com a OpenAI
│   ├── redis.go           # Serviço para integração com o Redis
│   └── openai_models.go   # Modelos usados pelo serviço OpenAI
├── utils/
│   └── logger.go          # Implementação de logger JSON
├── .env                   # Variáveis de ambiente (exemplo fornecido)
├── main.go                # Ponto de entrada da aplicação
```

---

## Pré-requisitos

- [Go 1.20+](https://golang.org/dl/)
- [Redis](https://redis.io/download)
- [Protoc](https://grpc.io/docs/protoc-installation/) (para gerar código Protobuf)
- [Swagger CLI](https://github.com/swaggo/swag) (opcional, para regenerar a documentação)

---

## Configuração

### 1. Clonar o Repositório
```bash
git clone https://github.com/seu-usuario/gosmart.git
cd gosmart
```

### 2. Configurar Variáveis de Ambiente
Crie um arquivo `.env` na raiz do projeto com o seguinte conteúdo:
```
OPENAI_API_KEY=your_openai_api_key
OPENAI_API_URL=https://api.openai.com/v1/chat/completions
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
```

### 3. Instalar Dependências
```bash
go mod tidy
```

### 4. Gerar Código Protobuf
Certifique-se de que o `protoc` está instalado, e execute:
```bash
protoc --go_out=. --go_opt=paths=source_relative ./entities/request.proto
```

### 5. Gerar Documentação Swagger
Se precisar atualizar a documentação, execute:
```bash
swag init
```

---

## Uso

1. Inicie o Redis:
   ```bash
   redis-server
   ```

2. Execute a aplicação:
   ```bash
   go run main.go
   ```

3. Acesse a API via Swagger:
   ```
   http://localhost:3000/swagger/index.html
   ```

---

## Scripts

### `main.go`
Ponto de entrada da aplicação, inicializa serviços, configura rotas e inicia o servidor Fiber.

### `config/environments.go`
Gerencia o carregamento das variáveis de ambiente do arquivo `.env`.

### `services/openai.go`
Integração direta com a OpenAI para geração de texto usando a API.

### `services/redis.go`
Gerencia a conexão com o Redis para armazenamento de logs.

### `utils/logger.go`
Implementa um logger estruturado em formato JSON com saída no `stdout`.

### `handlers/openai.go`
Rota que processa as requisições para gerar texto a partir de prompts.

### `router/router.go`
Define e organiza as rotas da aplicação.

---

## Dependências

- [Fiber](https://gofiber.io/): Framework web para Go.
- [Redis](https://redis.io/): Banco de dados em memória para armazenamento de logs.
- [Protobuf](https://developers.google.com/protocol-buffers): Para definição de dados estruturados.
- [Swaggo](https://github.com/swaggo/swag): Para documentação Swagger.

---

## Estrutura Modular

- **Config**: Gerencia configurações e variáveis de ambiente.
- **Entities**: Definições de dados com Protobuf.
- **Handlers**: Implementação das rotas da API.
- **Router**: Configuração das rotas.
- **Services**: Lógica de negócios e integrações externas (Redis, OpenAI).
- **Utils**: Funções auxiliares (ex.: logging).

---

## Contribuições

Contribuições são bem-vindas! Sinta-se à vontade para abrir um pull request ou relatar problemas na aba de issues.

---

## Licença

Este projeto está licenciado sob a [MIT License](LICENSE).

