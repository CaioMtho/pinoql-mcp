# PinoQL MCP

Your agent wants to be a real DBA!

---

## What is this?

PinoQL is a group of open-source software tools focused on providing reliable, flexible, and efficient data access.

This is the main PinoQL tool, called **PinoQL-MCP**. It is a server that implements the Model Context Protocol to allow AI Agents to interact with SQL Databases in safe flows.

The main goal of PinoQL-MCP is to establish an N-N relation between agents and databases, with JWT tokens managed by the clients. With this approach, you only need one instance running.

---

## Key Features

- **Model Context Protocol (MCP)**: Provides a standardized way for agents to communicate with SQL databases.
- **Secure Authentication**: Uses JWT tokens for client-managed authentication and authorization, and optionally API keys for the API routes.
- **Multi-Tenancy**: Supports multiple tenants, each with their own connections and audit logs.
- **Scalability**: Supports multiple agents and multiple databases simultaneously.
- **Flexibility**: Works with different SQL database engines (PostgreSQL, SQLite) without requiring custom integrations.
- **Audit Logging**: Comprehensive logging of all database operations for security and compliance.
- **Single Instance Architecture**: One server instance can handle multiple connections, reducing complexity.

---

## Use Cases

- Allowing AI agents to query and manage SQL databases safely.
- Acting as a middleware between applications and databases.
- Providing a unified interface for multiple agents to access multiple databases.
- Enabling secure, token-based access control for distributed environments.
- Multi-tenant database access management.

---

## Requirements

To build and develop in PinoQL-MCP you will need:

- Go 1.25.5 or later
- SQLite3 (for the internal database)

---

## Getting Started

1. **Clone and Install Dependencies**

   ```bash
   git clone https://github.com/CaioMtho/pinoql-mcp.git
   cd pinoql-mcp
   go mod tidy
   ```

2. **Configure Environment**

   Copy the example environment file and set your secrets:

   ```bash
   cp .env.example .env
   ```

   Edit `.env` with your values:
   - `MASTER_KEY`: A base64-encoded 32-byte key for encryption
   - `JWT_SECRET`: A secret key for JWT token signing

3. **Run the Server**

   ```bash
   go run cmd/main.go --enable-auth
   ```

   > Omit `--enable-auth` to disable API key authentication for API routes.

4. **Set up Tenants and Connections**

   Use the API endpoints to create tenants and database connections.

5. **Generate JWT Tokens**

   Issue JWT tokens for specific connections within tenants.

6. **Connect MCP Clients**

   Use the MCP endpoint with JWT authentication.

---

## Configuration

The server uses the following environment variables:

- `MASTER_KEY` (required): Base64-encoded 32-byte master key for encrypting sensitive data
- `JWT_SECRET` (required): Secret key used for signing JWT tokens
- `DB_PATH` (optional): Path to the SQLite database file (default: `./db/pinoql.sqlite`)
- `PORT` (optional): Port number to run the server on (default: `8080`)

Create a `.env` file in the root directory with these values.

---

## Database Setup

PinoQL-MCP uses SQLite for its internal metadata storage. Database migrations are applied automatically on startup using the migration files in `db/migrations/`.

The internal database stores:
- Tenants
- Connection configurations (encrypted)
- JWT tokens
- Audit logs

---

## API Endpoints

The server provides REST API endpoints for management, and an MCP endpoint for agent interactions.

### Authentication

API routes can be protected with API keys when `--enable-auth` is used. The MCP endpoint requires JWT authentication.

### Tenants

- `POST /api/v1/tenants` - Create a new tenant
- `GET /api/v1/tenants` - List all tenants
- `GET /api/v1/tenants/:id` - Get tenant details
- `PUT /api/v1/tenants/:id` - Update tenant
- `DELETE /api/v1/tenants/:id` - Delete tenant

### Connections

- `POST /api/v1/connections` - Create a database connection
- `GET /api/v1/connections` - List connections
- `GET /api/v1/connections/:id` - Get connection details
- `PUT /api/v1/connections/:id` - Update connection
- `DELETE /api/v1/connections/:id` - Delete connection

### JWT Tokens

- `POST /api/v1/jwt/issue` - Issue a new JWT token for connections
- `POST /api/v1/jwt/revoke` - Revoke a JWT token
- `GET /api/v1/jwt/list` - List active JWT tokens

### Audit Logs

- `GET /api/v1/audit/logs` - Retrieve audit logs
- `GET /api/v1/audit/stats` - Get audit statistics

### MCP

- `POST /mcp` - Model Context Protocol endpoint (requires JWT auth)

### Health Check

- `GET /api/v1/health` - Server health status
- `GET /ping` - Simple ping endpoint

---

## Building

To build a standalone binary:

```bash
go build -o pinoql-mcp cmd/main.go
```

Run the binary:

```bash
./pinoql-mcp --enable-auth
```

---

## Architecture

PinoQL-MCP is built with a modular architecture:

- **Adapters**: Database adapters for different SQL engines (PostgreSQL, SQLite)
- **Connection Management**: Handles database connections with pooling and limits
- **Credentials**: Manages tenants, connections, tokens, and audit logs
- **Crypto**: Encryption/decryption of sensitive connection data
- **MCP**: Implements the Model Context Protocol for agent interactions
- **Routes**: REST API endpoints for management

---

## Security

- All connection credentials are encrypted using AES-256
- JWT tokens have configurable expiration
- Audit logging tracks all database operations
- Optional API key authentication for management endpoints
- Connection limits and read-only modes supported

---

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

---

## License

This project is licensed under the MIT License. See the LICENSE file for details.

---

