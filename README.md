# RabbitMQ Stream Viewer

A modern web-based UI for browsing and inspecting RabbitMQ Streams. This tool provides a clean interface to view stream statistics, browse messages, and inspect AMQP properties without consuming messages.

## Features

- 🔍 **Stream Discovery**: Automatically discover all streams across multiple RabbitMQ connections
- 📊 **Real-time Statistics**: View message counts, stream sizes, and offset information
- 📨 **Message Browsing**: Non-destructive message reading with offset-based navigation
- 🏷️ **AMQP Properties**: Full visibility into message properties and metadata
- 🎨 **Modern UI**: Clean, responsive interface built with React and TailwindCSS
- 🐳 **Containerised**: Easy deployment with Docker
- ⚙️ **Multi-Connection**: Support for multiple RabbitMQ hosts and vhosts

## Architecture

- **Backend**: Go REST API with RabbitMQ Stream client
- **Frontend**: React + Vite + TailwindCSS
- **Deployment**: Docker container with multi-stage build

## Prerequisites

- Go 1.21+ (for local development)
- Node.js 20+ (for frontend development)
- Docker & Docker Compose (for containerised deployment)
- RabbitMQ server with streams plugin enabled

### Network Requirements

The following ports must be accessible on your RabbitMQ server:

| Port  | Protocol | Purpose                      | Required |
| ----- | -------- | ---------------------------- | -------- |
| 5672  | AMQP     | Standard AMQP connections    | Yes      |
| 5671  | AMQPS    | AMQP over TLS/SSL            | Optional |
| 5552  | Stream   | RabbitMQ Stream Protocol     | Yes      |
| 15672 | HTTP     | Management API & Web Console | Yes      |

**Note:** If your RabbitMQ server uses custom ports, adjust the configuration accordingly in `config.yaml`.

## Quick Start

### Using Docker Compose

1. **Configure your connections**:
   ```bash
   cp config.example.yaml config.yaml
   # Edit config.yaml with your RabbitMQ connection details
   ```

2. **Build and run**:
   ```bash
   docker-compose up -d
   ```

3. **Access the UI**:
   Open http://localhost:8080 in your browser

### Local Development

1. **Backend Setup**:
   ```bash
   # Install dependencies
   go mod download
   
   # Run the server
   go run cmd/server/main.go -config config.yaml
   ```

2. **Frontend Setup**:
   ```bash
   cd web
   npm install
   npm run dev
   ```

3. **Access the development UI**:
   Open http://localhost:5173 in your browser (Vite dev server)

## Configuration

Configuration is done via a YAML file. See `config.example.yaml` for a complete example.

```yaml
server:
  port: 8080

connections:
  - id: dev
    name: Development
    host: localhost
    port: 5672
    vhost: /
    username: guest
    password: guest
    http_port: 15672

  - id: prod
    name: Production
    host: rabbitmq.example.com
    port: 5672
    vhost: /production
    username: admin
    password: secret
    http_port: 15672
```

### Configuration Options

- `server.port`: Port for the web server (default: 8080)
- `connections[]`: Array of RabbitMQ connections
  - `id`: Unique identifier for the connection
  - `name`: Display name
  - `host`: RabbitMQ server hostname
  - `port`: AMQP port (default: 5672, or 5671 for TLS)
  - `vhost`: Virtual host (use "/" for default)
  - `username`: RabbitMQ username
  - `password`: RabbitMQ password
  - `http_port`: Management API port (default: 15672)
  - `stream_port`: RabbitMQ Stream Protocol port (default: 5552)

## Testing

### Running Tests

```bash
# Run all Go tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/config
go test ./internal/api
```

### Test Publisher

A test publisher is included to generate sample messages for testing:

```bash
# Run the publisher (creates stream and publishes messages)
go run cmd/publisher/main.go \
  -stream test-stream \
  -uri amqp://root:magic@cdvm:5672 \
  -rate 1 \
  -count 100

# Publish continuously at 5 messages per second
go run cmd/publisher/main.go \
  -stream test-stream \
  -rate 5

# Use with Docker
docker run --rm \
  --network host \
  rmq-stream-viewer:latest \
  ./publisher -stream test-stream -uri amqp://guest:guest@localhost:5672
```

#### Publisher Options

- `-stream`: Stream name to publish to (default: "test-stream")
- `-uri`: RabbitMQ connection URI (default: "amqp://root:magic@cdvm:5672")
- `-rate`: Messages per second (default: 1)
- `-count`: Total messages to send (0 = infinite, default: 0)

The publisher generates JSON messages with the format:
```json
{
  "timestamp": "2025-10-16T03:00:00Z",
  "counter": 123
}
```

Each message includes AMQP properties:
- `message_id`: Unique UUID
- `content_type`: "application/json"
- `creation_time`: Message timestamp
- `subject`: "test-message-{counter}"
- Application properties: `app_id`, `counter`, `environment`

## API Endpoints

### REST API

- `GET /api/connections` - List all configured connections
- `GET /api/streams` - List all streams across all connections
- `GET /api/streams/:connection_id/:stream_name/stats` - Get stream statistics
- `GET /api/streams/:connection_id/:stream_name/messages?offset=X&limit=Y` - Read messages

### Health Check

- `GET /health` - Server health status

## UI Usage

### Navigating Streams

1. **Select a Connection**: Click on a connection in the sidebar to expand it
2. **Select a Stream**: Click on a stream name to view its details
3. **View Statistics**: The default view shows stream statistics (message count, size, offsets)
4. **Browse Messages**: Click the "Messages" button to switch to message browsing mode

### Message Browsing

- **Navigate**: Use Previous/Next buttons or arrow keys (← →)
- **Jump to Offset**: Enter a specific offset and click "Go"
- **Jump to First/Last**: Quick navigation to stream boundaries
- **Change Page Size**: Adjust the limit dropdown (5, 10, 25, 50, 100)
- **Select Message**: Click on a message in the list to view full details
- **Copy Content**: Use copy buttons to copy properties or message content

### Keyboard Shortcuts

- `←` (Left Arrow): Previous page
- `→` (Right Arrow): Next page

## Building from Source

### Build Backend

```bash
go build -o server cmd/server/main.go
go build -o publisher cmd/publisher/main.go
```

### Build Frontend

```bash
cd web
npm install
npm run build
```

The built files will be in `web/dist/`.

### Build Docker Image

```bash
docker build -t rmq-stream-viewer:latest .
```

## Development

### Project Structure

```
├── cmd/
│   ├── server/          # Main server application
│   └── publisher/       # Test message publisher
├── internal/
│   ├── config/          # Configuration management
│   ├── rabbitmq/        # RabbitMQ client
│   └── api/             # HTTP handlers
├── web/
│   ├── src/
│   │   ├── components/  # React components
│   │   ├── services/    # API client
│   │   └── App.jsx      # Main application
│   └── dist/            # Built frontend (generated)
├── config.yaml          # Configuration file
├── Dockerfile           # Container build
└── docker-compose.yml   # Docker Compose config
```

### Adding New Features

1. **Backend**: Add handlers in `internal/api/`
2. **Frontend**: Add components in `web/src/components/`
3. **Tests**: Add tests alongside code with `_test.go` suffix
4. **Documentation**: Update this README

## Troubleshooting

### Connection Issues

- Verify RabbitMQ is running and accessible
- Check that the management plugin is enabled: `rabbitmq-plugins enable rabbitmq_management`
- Ensure the streams plugin is enabled: `rabbitmq-plugins enable rabbitmq_stream`
- Verify credentials and ports in `config.yaml`

### No Streams Showing

- Make sure streams exist on the RabbitMQ server
- Check connection credentials and vhost configuration
- Verify the user has permissions to list streams

### Message Reading Issues

- Ensure the stream has messages
- Check that the offset is within the valid range
- Verify network connectivity to RabbitMQ

### Frontend Not Loading

- Check that the backend server is running
- Verify the frontend was built: `cd web && npm run build`
- Check browser console for errors
- Ensure API proxy is configured correctly in development

## Performance

- **Non-Destructive Reads**: Messages are read from streams without consuming them
- **Offset-Based**: Direct offset access for fast navigation
- **Pagination**: Configurable page sizes to manage large streams
- **Auto-Refresh**: Optional auto-refresh for real-time statistics

## Security Considerations

- Store credentials securely (use environment variables in production)
- Use HTTPS in production environments
- Restrict access to the management API
- Consider authentication for the web UI (not implemented in base version)

## License

MIT License - See LICENSE file for details

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

