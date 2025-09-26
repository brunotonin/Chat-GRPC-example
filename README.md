# Chat gRPC com RabbitMQ

Um sistema de chat em tempo real usando gRPC para comunicação cliente-servidor e RabbitMQ como message broker para distribuição das mensagens entre múltiplos clientes.

## 🏗️ Arquitetura

O projeto consiste em:

- **Server**: Servidor gRPC que implementa os serviços de Chat e Echo
- **Client**: Cliente que simula múltiplos usuários conversando simultaneamente  
- **RabbitMQ**: Message broker para distribuir mensagens entre clientes conectados
- **Benchmark**: Ferramenta para medir latência vs número de mensagens

### Fluxo de Mensagens

1. Cliente envia mensagem via gRPC stream para o servidor
2. Servidor serializa a mensagem e publica no RabbitMQ (exchange fanout)
3. RabbitMQ distribui para todas as filas dos clientes conectados
4. Cada cliente recebe e processa a mensagem via seu stream gRPC

## 🚀 Executando o Projeto

### Pré-requisitos

- Docker e Docker Compose
- Go 1.24+ (para desenvolvimento local)

### Usando Docker Compose (Recomendado)

```bash
# Inicia RabbitMQ, servidor e cliente de exemplo
docker-compose up

# Para executar apenas infraestrutura
docker-compose up rabbitmq chat-server

# Para executar cliente com N usuários
docker-compose run chat-client ./chat-client <numero_usuarios>
```

### Desenvolvimento Local

1. **Instale as dependências:**
   ```bash
   go mod download
   ```

2. **Inicie o RabbitMQ:**
   ```bash
   docker run -d --name rabbitmq \
     -p 5672:5672 -p 15672:15672 \
     -e RABBITMQ_DEFAULT_USER=guest \
     -e RABBITMQ_DEFAULT_PASS=guest \
     rabbitmq:3-management
   ```

3. **Configure as variáveis de ambiente:**
   ```bash
   # server/.env
   RABBITMQ_URL=amqp://guest:guest@localhost:5672/
   ENABLE_PPROF=true

   # client/.env  
   GRPC_SERVER_ADDRESS=localhost:50051
   ```

4. **Execute o servidor:**
   ```bash
   cd server
   go run .
   ```

5. **Execute o cliente:**
   ```bash
   cd client
   go run . <numero_usuarios>
   ```

## 📊 Benchmark

O projeto inclui uma ferramenta de benchmark que mede a latência conforme o número de usuários aumenta:

```bash
cd bechmark
go run .
```

O benchmark:
- Inicia com 50 usuários e aumenta progressivamente
- Mede latência média de entrega das mensagens
- Gera um gráfico PNG com os resultados
- Para automaticamente ao atingir 1000 usuários

## 🔧 Serviços gRPC

### ChatService
- **Método**: `Chat(stream ChatMessage) returns (stream ChatMessage)`
- **Funcionalidade**: Stream bidirecional para envio/recebimento de mensagens em tempo real

### EchoService  
- **Método**: `Echo(EchoRequest) returns (EchoResponse)`
- **Funcionalidade**: Serviço simples de echo para testes de conectividade

## 🐰 RabbitMQ

**Configuração:**
- **Exchange**: `chat_exchange` (fanout)
- **Filas**: Uma fila exclusiva e temporária por cliente conectado
- **Binding**: Todas as filas recebem todas as mensagens (broadcast)

**Interface Web**: http://localhost:15672 (guest/guest)

## 📁 Estrutura do Projeto

```
.
├── server/                 # Servidor gRPC
│   ├── main.go            # Ponto de entrada
│   ├── chat-server.go     # Implementação ChatService
│   ├── echo-server.go     # Implementação EchoService
│   └── Dockerfile
├── client/                 # Cliente simulador
│   ├── main.go
│   └── Dockerfile  
├── chat/                   # Definições protobuf do chat
│   ├── chat.proto
│   └── proto/             # Código gerado
├── echo/                   # Definições protobuf do echo
│   ├── echo.proto
│   └── proto/             # Código gerado
├── bechmark/              # Ferramenta de benchmark
│   └── main.go
├── docker-compose.yml      # Orquestração dos serviços
└── .github/workflows/      # CI/CD
```

## 🔍 Monitoramento

### pprof (Profiling)
Quando `ENABLE_PPROF=true`:
- **URL**: http://localhost:6060/debug/pprof/
- **Métricas**: CPU, Memory, Goroutines, etc.

### RabbitMQ Management
- **URL**: http://localhost:15672
- **Usuário**: guest / guest

### Variáveis de Ambiente

**Servidor:**
- `RABBITMQ_URL`: URL de conexão do RabbitMQ
- `ENABLE_PPROF`: Habilita profiling (true/false)

**Cliente:**  
- `GRPC_SERVER_ADDRESS`: Endereço do servidor gRPC
