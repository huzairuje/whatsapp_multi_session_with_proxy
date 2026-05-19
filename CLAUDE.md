# CLAUDE.md

This file provides guidance when working with code in this repository.

## Project Overview

WhatsApp multi-session bulk messaging system with React frontend and Go backend. Features advanced anti-ban protection, session management, real-time monitoring, and proxy support. Built with [whatsmeow](https://github.com/tulir/whatsmeow) library for WhatsApp Web emulation.

## Prerequisites

- **Go 1.25.0+** (backend)
- **Node.js 18+** (frontend)
- **gcc/build-essential** (required for CGO and SQLite)
  - Linux: `sudo apt install build-essential`
  - Windows: Install mingw or gcc via choco
  - macOS: Install Xcode Command Line Tools or mingw for cross-compilation

## Quick Start

### Using Makefile (Recommended)

```bash
# Show all available commands
make help

# Quick start (install dependencies + build everything)
make quickstart

# Development mode (run in separate terminals)
make dev-backend      # Terminal 1: Run backend
make dev-frontend     # Terminal 2: Run frontend
```

### Manual Setup

**Backend:**
```bash
cd backend
go mod download
env GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-w -s" -trimpath -o ../bin/whatsapp_multi_session-linux-amd64
```

**Frontend:**
```bash
cd frontend
npm install
npm run dev
```

## Build Commands

**IMPORTANT:** CGO_ENABLED=1 is required for SQLite support.

### Linux
```bash
env GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-w -s" -trimpath -o ../bin/whatsapp_multi_session-linux-amd64
```

### Windows (from Linux/macOS with mingw)
```bash
env GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -ldflags="-w -s" -trimpath -o ../bin/whatsapp_multi_session-windows-amd64.exe
```

### Windows (native)
```cmd
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=1
go build -o ../bin/whatsapp_multi_session-windows-amd64.exe
```

## Running the Application

### Development

**Terminal 1 - Backend:**
```bash
cd backend
cp config.local.yaml.example config.local.yaml
go run main.go
```

**Terminal 2 - Frontend:**
```bash
cd frontend
npm run dev
```

### Production

```bash
# Linux
./bin/whatsapp_multi_session-linux-amd64

# Windows
whatsapp_multi_session-windows-amd64.exe
```

## Configuration

Uses [Viper](https://github.com/spf13/viper) for configuration management.

### Environment Files
- `config.local.yaml` - Local development
- `config.dev.yaml` - Development environment
- `config.uat.yaml` - UAT environment
- `config.prod.yaml` - Production environment

### Config Search Paths
1. `/etc/whatsapp_multi_session_with_proxies`
2. `$HOME/.whatsapp_multi_session_with_proxies`
3. Current directory (`.`)

### Key Configuration Options

**Server:**
- **env**: Environment name (local, dev, uat, prod)
- **port**: HTTP server port (default: 1234)
- **pprof.enable**: Enable profiling endpoint
- **pprof.portPprof**: Profiling port (default: 6060)

**Proxy:**
- **proxy.enable**: Enable proxy support
- **proxy.directory**: Path to proxy list file

**Auto-Login/Logout:**
- **startUp.enableAutoLogin**: Auto-login on startup
- **shutDown.enableAutoShutDown**: Auto-disconnect on shutdown
- **autoLogout**: Enable automatic logout
- **autoDisconnect**: Enable automatic disconnect

**Bulk Send Anti-Ban Features:**
- **bulkSend.minDelay**: Minimum delay between messages (ms)
- **bulkSend.maxDelay**: Maximum delay between messages (ms)
- **bulkSend.batchSize**: Messages per batch before pause
- **bulkSend.batchPauseMin**: Minimum batch pause (seconds)
- **bulkSend.batchPauseMax**: Maximum batch pause (seconds)
- **bulkSend.dailyLimit**: Max messages per sender per day
- **bulkSend.typingDelayMin**: Min "composing" presence duration (ms)
- **bulkSend.typingDelayMax**: Max "composing" presence duration (ms)
- **bulkSend.enablePresenceSimulation**: Send "composing" before each message
- **bulkSend.allowedHourStart**: Earliest hour to send (0-23)
- **bulkSend.allowedHourEnd**: Latest hour to send (0-23)
- **bulkSend.timezone**: Timezone for time restrictions
- **bulkSend.enableTimeRestrictions**: Enable time-of-day restrictions
- **bulkSend.errorBackoffMinutes**: Pause duration after rate limit error
- **bulkSend.enableRecipientValidation**: Validate recipients before sending
- **bulkSend.validationCacheDuration**: Cache validation results (hours)
- **bulkSend.enableHealthCheck**: Check session health before bulk send
- **bulkSend.maxErrorRate**: Max acceptable error rate (0.0-1.0)

**Database:**
- **postgres.enablePostgres**: Use PostgreSQL instead of SQLite
- **postgres.host**: PostgreSQL host
- **postgres.port**: PostgreSQL port
- **postgres.user**: PostgreSQL user
- **postgres.password**: PostgreSQL password
- **postgres.dbName**: Database name
- **postgres.schema**: Database schema

**Other:**
- **messageStatus.enable**: Enable message status tracking
- **cronjob.autoPresence.enable**: Enable auto-presence cron job
- **deleteAfterSend.enable**: Delete messages after sending

## Architecture Overview

### Directory Structure

```
backend/
├── boot/              # Application bootstrap
├── commandhandler/    # Core business logic
├── config/            # Configuration management
├── cronjob/           # Background jobs
├── database/          # Database abstraction
├── handler/           # HTTP request handlers
├── listener/          # Event listeners
├── primitive/         # Shared types and constants
├── proxy/             # Proxy management
├── routers/           # Route registration
├── utils/             # Utility functions
├── validator/         # Input validation
├── main.go            # Application entry point
└── go.mod             # Go dependencies

frontend/
├── src/               # React source code
├── public/            # Static assets
├── package.json       # Node dependencies
└── vite.config.ts     # Vite configuration
```

### Application Bootstrap (backend/boot/setup.go)

The `Setup()` function initializes components in this order:
1. **Database**: PostgreSQL or SQLite based on config
2. **Proxy Manager**: Loads proxy list from configured directory
3. **Command Handler**: Core business logic for WhatsApp operations
4. **Listener**: Event listeners for startup/shutdown events
5. **HTTP Handler**: Request handlers for REST API
6. **Router**: API route registration
7. **Cronjob**: Background jobs (e.g., auto-presence)

### Key Components

#### commandhandler/
Core business logic for WhatsApp operations. Manages WhatsApp client lifecycle, message sending, QR code generation, device management, and bulk sending with anti-ban features. Uses concurrent map to store active clients keyed by JID.

#### handler/
HTTP request handlers that validate input, interact with CommandHandler, and return JSON responses. Implements all API endpoints with proper error handling.

#### listener/
Event-driven lifecycle management using [gookit/event](https://github.com/gookit/event):
- `TriggerStartUp()`: Fires startup event for auto-login
- `ListenForShutdownEvent()`: Listens for graceful shutdown signal

#### proxy/
Proxy management for WhatsApp connections. Loads proxy list from file and assigns proxies to WhatsApp sessions. Supports rotation and per-session proxy assignment.

#### database/
Database abstraction layer supporting:
- **SQLite** (default): Uses `go.mau.fi/whatsmeow/store/sqlstore`
- **PostgreSQL**: Alternative storage backend

#### primitive/
Shared types and constants:
- `constant.go`: Application constants and error messages
- `model.go`: Domain models
- `request.go`: API request structures
- `response.go`: API response structures

#### cronjob/
Background job scheduler for recurring tasks like auto-presence updates.

#### bulksender/
Bulk message sending with anti-ban protection:
- Rate limiting and delays
- Time-of-day restrictions
- Recipient validation with caching
- Health monitoring
- Error-based backoff

### Event System

Uses `github.com/gookit/event` for lifecycle events:
- `primitive.ShutDownEvent`: Fired on graceful shutdown (SIGINT/SIGTERM)
- Startup events: Triggered by listener for auto-login

### Graceful Shutdown

The application handles SIGINT/SIGTERM signals:
1. Receives shutdown signal
2. Fires `primitive.ShutDownEvent`
3. Waits 1 second for cleanup
4. Shuts down HTTP server gracefully

## API Endpoints

All endpoints are registered in `backend/routers/routers.go`:

### Connection Management
- `GET /connect` - Connect a WhatsApp session
- `POST /connect-bulk` - Connect multiple sessions
- `GET /disconnect` - Disconnect a session
- `POST /disconnect-bulk` - Disconnect multiple sessions
- `GET /autologin` - Auto-login handler
- `GET /auto-disconnect` - Auto-disconnect handler
- `POST /logout` - Logout a session

### QR Code & Pairing
- `GET /qr` - Generate QR code (image)
- `GET /qr-json` - Generate QR code (JSON response)
- `GET /pair-code` - Generate pairing code

### Messaging
- `POST /send` - Send text message
- `POST /send-bulk` - Send bulk messages with anti-ban protection
- `GET /send-bulk/status` - Get bulk send status
- `POST /presence` - Send presence update
- `DELETE /message` - Delete messages

### User & Device Management
- `POST /check-user` - Check if users exist on WhatsApp
- `POST /check-user-single` - Check single user
- `GET /devices` - List all devices
- `GET /devices/:jid` - Get device details
- `GET /device-proxies` - List device-proxy mappings
- `GET /status` - Get session status

### File Upload
- `POST /upload` - Upload media file
- `POST /upload-single` - Upload single media file

### Health Check
- `GET /health-check` - Health check endpoint

## Development Notes

### Query Parameters
Most endpoints use query parameters for `sender` (WhatsApp JID). Example:
```
GET /connect?sender=6281234567890
```

### JID Format
WhatsApp JIDs (Jabber IDs) use the format: `{phone_number}@s.whatsapp.net`
- The application automatically appends `@s.whatsapp.net` using `types.DefaultUserServer`
- Phone numbers should include country code without `+` (e.g., `6281234567890`)

### Client Storage
Active WhatsApp clients are stored in a concurrent map in `commandhandler` package:
- Key: User part of JID (phone number)
- Value: WhatsApp client instance
- Access via `LoadClientConcurrent()` function

### Proxy Assignment
When proxy support is enabled (`config.Conf.Proxy.Enable`):
- Each WhatsApp session can be assigned a specific proxy
- Proxy list is loaded from file specified in `config.Conf.Proxy.Directory`
- Proxy manager handles assignment and rotation

### Database Schema
The whatsmeow library manages its own database schema for:
- Device information
- Message store
- Session data

No custom migrations are required for basic functionality.

### Logging
Uses `github.com/sirupsen/logrus` for structured logging. Log level and format are configurable via config file.

### Anti-Ban Features
The bulk sender implements multiple anti-ban mechanisms:
- **Random delays** between messages (configurable min/max)
- **Batch pauses** after sending N messages
- **Time-of-day restrictions** to avoid sending outside business hours
- **Presence simulation** (typing indicators before messages)
- **Recipient validation** with caching to avoid invalid numbers
- **Health checks** before bulk operations
- **Error-based backoff** when rate limits are detected
- **Daily limits** per sender to prevent abuse

### Testing
When adding new endpoints or modifying existing ones:
1. Test with both SQLite and PostgreSQL backends
2. Test with proxy enabled/disabled
3. Verify graceful shutdown behavior
4. Check concurrent session handling
5. Test anti-ban features with bulk sending

## Dependencies

Key Go dependencies:
- `github.com/gin-gonic/gin` - HTTP framework
- `go.mau.fi/whatsmeow` - WhatsApp Web client
- `github.com/spf13/viper` - Configuration management
- `github.com/sirupsen/logrus` - Logging
- `github.com/gookit/event` - Event system
- `github.com/lib/pq` - PostgreSQL driver
- `github.com/mattn/go-sqlite3` - SQLite driver
- `github.com/skip2/go-qrcode` - QR code generation
- `google.golang.org/protobuf` - Protocol buffers

Key Node dependencies:
- `react` - UI framework
- `vite` - Build tool
- `typescript` - Type safety
