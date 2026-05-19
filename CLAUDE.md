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
- **pprof.pprofPort**: Profiling port (default: 5555)
- **pprof.pprofAddress**: Profiling address (default: localhost)

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

**Redis:**
- **redis.enableRedis**: Enable Redis caching
- **redis.host**: Redis host
- **redis.port**: Redis port
- **redis.password**: Redis password
- **redis.db**: Redis database number

**Other:**
- **cronjob.autoPresence.enable**: Enable auto-presence cron job
- **cronjob.autoPresence.cronJobSchedule**: Cron schedule for auto-presence
- **deleteAfterSend.enable**: Delete messages after sending

## Architecture Overview

### Directory Structure

```
backend/
├── activity/          # Activity logging service
├── auth/              # Authentication & authorization
├── boot/              # Application bootstrap
├── client/            # External clients (HTTP, Redis)
│   ├── http/          # HTTP client
│   └── redis/         # Redis client
├── commandhandler/    # Core business logic
├── config/            # Configuration management
├── cronjob/           # Background jobs
├── database/          # Database abstraction
├── handler/           # HTTP request handlers
├── listener/          # Event listeners
├── message/           # Message tracking service
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
2. **Auth Service**: User authentication and JWT token management
3. **Message Service**: Message tracking and statistics
4. **Activity Service**: Activity logging and audit trail
5. **Proxy Manager**: Loads proxy list from configured directory
6. **Command Handler**: Core business logic for WhatsApp operations
7. **Listener**: Event listeners for startup/shutdown events
8. **HTTP Handler**: Request handlers for REST API
9. **Router**: API route registration
10. **Cronjob**: Background jobs (e.g., auto-presence)

### Key Components

#### auth/
Authentication and authorization service with JWT-based token management:
- User registration and login with bcrypt password hashing
- JWT access tokens (15 min expiry) and refresh tokens (7 day expiry)
- Auth middleware for protecting routes
- Default admin user created on first run (username: admin, password: admin123)
- Password change functionality
- Supports both SQLite and PostgreSQL

#### message/
Message tracking service for monitoring sent messages:
- Records all sent messages with sender, recipient, content, and status
- Status tracking: pending, sent, delivered, read, failed
- Message statistics by sender (total, sent, delivered, read, failed counts)
- Aggregate statistics across all senders
- Supports both SQLite and PostgreSQL

#### activity/
Activity logging service for audit trail and monitoring:
- Logs user activities (connect, disconnect, send_message, bulk_send, etc.)
- Activity types: connect, disconnect, send_message, bulk_send, qr_generate, pair_code, logout
- Filter activities by sender, type, or time range
- Activity statistics and recent activity queries
- Supports both SQLite and PostgreSQL

#### client/
External client integrations:
- **http/**: HTTP client for making external API calls
- **redis/**: Redis client for caching and session management

#### commandhandler/
Core business logic for WhatsApp operations. Manages WhatsApp client lifecycle, message sending, QR code generation, device management, and bulk sending with anti-ban features. Uses concurrent map to store active clients keyed by JID.

#### handler/
HTTP request handlers that validate input, interact with CommandHandler, and return JSON responses. Implements all API endpoints with proper error handling. Split into multiple files:
- `handler.go`: Main WhatsApp operation handlers (33.7K)
- `auth_handler.go`: Authentication handlers (1.8K)
- `message_handler.go`: Message tracking handlers (2.4K)
- `activity_handler.go`: Activity logging handlers (2.3K)

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

### Authentication (Public Routes)
- `POST /login` - User login with username/password
- `POST /refresh-token` - Refresh JWT access token
- `GET /health-check` - Health check endpoint

### Connection Management (Protected)
- `GET /connect` - Connect a WhatsApp session
- `POST /connect-bulk` - Connect multiple sessions
- `GET /disconnect` - Disconnect a session
- `POST /disconnect-bulk` - Disconnect multiple sessions
- `GET /autologin` - Auto-login handler
- `GET /auto-disconnect` - Auto-disconnect handler
- `POST /logout` - Logout a session

### QR Code & Pairing (Protected)
- `GET /qr` - Generate QR code (image)
- `GET /qr-json` - Generate QR code (JSON response)
- `GET /pair-code` - Generate pairing code

### Messaging (Protected)
- `POST /send` - Send text message
- `POST /send-bulk` - Send bulk messages with anti-ban protection
- `GET /send-bulk/status` - Get bulk send status
- `POST /presence` - Send presence update
- `DELETE /message` - Delete messages

### User & Device Management (Protected)
- `POST /check-user` - Check if users exist on WhatsApp
- `POST /check-user-single` - Check single user
- `GET /devices` - List all devices
- `GET /devices/:jid` - Get device details
- `GET /device-proxies` - List device-proxy mappings
- `GET /status` - Get session status

### File Upload (Protected)
- `POST /upload` - Upload media file
- `POST /upload-single` - Upload single media file

### User Account (Protected)
- `POST /change-password` - Change user password

### Message Tracking (Protected)
- `GET /messages` - Get messages with filtering
- `GET /messages/stats` - Get message statistics by sender
- `GET /messages/stats/all` - Get aggregate message statistics
- `POST /messages/status` - Update message status

### Activity Logging (Protected)
- `POST /activities/log` - Log user activity
- `GET /activities` - Get recent activities
- `GET /activities/sender` - Get activities by sender
- `GET /activities/type` - Get activities by type
- `GET /activities/stats` - Get activity statistics

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

### Authentication System
The application uses JWT-based authentication:
- **Default Admin User**: Created automatically on first run
  - Username: `admin`
  - Password: `admin123`
  - **IMPORTANT**: Change the default password immediately after first login
- **Access Tokens**: Valid for 15 minutes
- **Refresh Tokens**: Valid for 7 days
- **Protected Routes**: All endpoints except `/login`, `/refresh-token`, and `/health-check` require authentication
- **Auth Header**: Include JWT token in `Authorization: Bearer <token>` header

### Database Schema
The whatsmeow library manages its own database schema for:
- Device information
- Message store
- Session data

Additional tables are created by the application services:
- **users**: User accounts and authentication (auth service)
- **messages**: Message tracking and status (message service)
- **activities**: Activity logs and audit trail (activity service)

All tables support both SQLite and PostgreSQL.

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
2. Test with Redis enabled/disabled
3. Test with proxy enabled/disabled
4. Verify graceful shutdown behavior
5. Check concurrent session handling
6. Test anti-ban features with bulk sending
7. Test authentication and authorization
8. Verify message tracking and activity logging

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
- `github.com/golang-jwt/jwt/v5` - JWT authentication
- `golang.org/x/crypto/bcrypt` - Password hashing

Key Node dependencies:
- `react` - UI framework
- `vite` - Build tool
- `typescript` - Type safety
