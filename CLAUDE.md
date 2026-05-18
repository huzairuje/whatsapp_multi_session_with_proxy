# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

WhatsApp multi-session engine built with Go that emulates WhatsApp Web using the [whatsmeow](https://github.com/tulir/whatsmeow) library. Provides a REST API for managing multiple WhatsApp sessions with proxy support, message sending, QR code generation, and device management.

## Prerequisites

- **Go 1.22+** (currently using Go 1.24)
- **gcc/build-essential** (required for CGO and SQLite)
  - Linux: `sudo apt install build-essential`
  - Windows: Install mingw or gcc via choco
  - macOS: Install Xcode Command Line Tools or mingw for cross-compilation

## Build Commands

**IMPORTANT:** CGO_ENABLED=1 is required for SQLite support.

### Linux (native or cross-compile)
```bash
env GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-w -s" -trimpath -o bin/whatsapp_multi_session_with_proxies-linux-amd64
```

### Windows (from Linux/macOS)
```bash
env GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-w -s" -trimpath -o bin/whatsapp_multi_session_with_proxies-windows-amd64.exe
```

### Windows (native)
```cmd
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=1
go build -o bin/whatsapp_multi_session-windows-amd64.exe
```

### macOS to Windows (using mingw)
```bash
env GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -ldflags="-w -s" -trimpath -o bin/whatsapp_multi_session_with_proxies-windows-amd64.exe
```

## Running the Application

### Development
```bash
# Copy example config
cp config.local.yaml.example config.local.yaml

# Run with default (local) environment
go run main.go

# Run with specific environment
go run main.go -env=prod
```

### Production
```bash
# Linux
./whatsapp_multi_session_with_proxies-linux-amd64

# Windows
whatsapp_multi_session_with_proxies-windows-amd64.exe
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
- **port**: HTTP server port (default: 1234)
- **pprof**: Enable profiling endpoint (default port: 5555)
- **proxy.enable**: Enable proxy support for WhatsApp connections
- **proxy.directory**: Path to proxy list file
- **startUp.enableAutoLogin**: Auto-login on startup
- **shutDown.enableAutoShutDown**: Auto-disconnect on shutdown
- **autoLogout**: Enable automatic logout
- **autoDisconnect**: Enable automatic disconnect
- **cronjob.autoPresence**: Configure auto-presence cron job
- **deleteAfterSend.enable**: Delete messages after sending
- **postgres**: PostgreSQL configuration (alternative to SQLite)
- **redis**: Redis configuration (optional)

## Architecture Overview

### Application Bootstrap (boot/setup.go)
The `Setup()` function initializes all components in this order:
1. **Database**: SQLite (default) or PostgreSQL based on config
2. **Proxy Manager**: Loads proxy list from configured directory
3. **Command Handler**: Core business logic for WhatsApp operations
4. **Listener**: Event listeners for startup/shutdown events
5. **HTTP Handler**: Request handlers for REST API
6. **Router**: API route registration
7. **Cronjob**: Background jobs (e.g., auto-presence)

### Key Components

#### commandhandler/
Core business logic for WhatsApp operations. Manages WhatsApp client lifecycle, message sending, QR code generation, and device management. Uses a concurrent map to store active WhatsApp clients keyed by JID (user identifier).

#### handler/
HTTP request handlers that validate input, interact with CommandHandler, and return JSON responses. Each handler corresponds to an API endpoint.

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

All endpoints are registered in `routers/routers.go`:

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
- `POST /send-bulk` - Send bulk messages
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

### Testing
When adding new endpoints or modifying existing ones:
1. Test with both SQLite and PostgreSQL backends
2. Test with proxy enabled/disabled
3. Verify graceful shutdown behavior
4. Check concurrent session handling
