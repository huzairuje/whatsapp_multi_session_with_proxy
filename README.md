# WhatsApp Multi-Session Bulk Sender

A comprehensive WhatsApp bulk messaging system with React frontend and Go backend, featuring advanced anti-ban protection, session management, and real-time monitoring.

## 🚀 Features

### Backend (Go)
- **Multi-Session Management** - Handle multiple WhatsApp accounts simultaneously
- **Anti-Ban Protection** - Advanced rate limiting, delays, and human-like behavior
- **Proxy Support** - Per-session proxy assignment with rotation
- **Health Monitoring** - Real-time session health tracking
- **Recipient Validation** - Automatic validation with caching
- **Time-of-Day Restrictions** - Send only during configured hours
- **Error-Based Backoff** - Automatic pause on rate limit detection
- **Bulk Sending** - Sequential sending with configurable delays

### Frontend (React)
- **Dashboard** - Overview of all sessions and statistics
- **Session Management** - Connect/disconnect devices, QR code scanning
- **Bulk Send** - Send messages to multiple recipients with templates
- **Recipient Management** - Upload, validate, and manage recipient lists
- **Message Templates** - Create and save reusable message templates
- **Settings** - Configure all anti-ban parameters
- **Real-time Updates** - Live session status and health monitoring

## 📋 Prerequisites

- **Go 1.22+** (backend)
- **Node.js 18+** (frontend)
- **gcc/build-essential** (for SQLite support)

## ⚡ Quick Start with Makefile

The easiest way to build and run the application is using the provided Makefile:

```bash
# Show all available commands
make help

# Quick start (install dependencies + build everything)
make quickstart

# Or step by step:
make install          # Install all dependencies
make build            # Build both backend and frontend

# Development mode (run in separate terminals)
make dev-backend      # Terminal 1: Run backend
make dev-frontend     # Terminal 2: Run frontend

# Build specific targets
make build-backend    # Build Go backend for Linux
make build-windows    # Build Go backend for Windows
make build-frontend   # Build React frontend

# Clean build artifacts
make clean
```

## 🛠️ Manual Installation

### Backend Setup

```bash
# Navigate to backend directory
cd backend

# Install Go dependencies
go mod download

# Build for Linux
env GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-w -s" -trimpath -o ../bin/whatsapp_multi_session-linux-amd64

# Build for Windows (from Linux/macOS with mingw)
env GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -ldflags="-w -s" -trimpath -o ../bin/whatsapp_multi_session-windows-amd64.exe
```

### Frontend Setup

```bash
cd frontend

# Install dependencies
npm install

# Development mode
npm run dev

# Production build
npm run build
```

## 🚀 Running the Application

### Development Mode

**Terminal 1 - Backend:**
```bash
# Copy example config
cp config.local.yaml.example backend/config.local.yaml

# Edit backend/config.local.yaml with your settings

# Run backend
cd backend
go run main.go
```

**Terminal 2 - Frontend:**
```bash
cd frontend
npm run dev
```

Access the application at: **http://localhost:3000**

### Production Mode

**Backend:**
```bash
# Linux
./bin/whatsapp_multi_session-linux-amd64 -env=prod

# Windows
bin\whatsapp_multi_session-windows-amd64.exe -env=prod
```

**Frontend:**
```bash
cd frontend
npm run build
npm run preview
```

## ⚙️ Configuration

### Backend Configuration (`config.prod.yaml`)

```yaml
port: 1234
proxy:
  enable: true
  directory: ./proxies.txt

bulkSend:
  minDelay: 20000              # 20s between messages
  maxDelay: 60000              # 60s max delay
  batchSize: 8                 # Messages per batch
  batchPauseMin: 360           # 6 min batch pause
  batchPauseMax: 720           # 12 min batch pause
  dailyLimit: 40               # Max messages/day
  typingDelayMin: 2500         # Typing simulation
  typingDelayMax: 6000
  enablePresenceSimulation: true
  
  # Anti-Ban Features
  allowedHourStart: 8          # 8 AM
  allowedHourEnd: 22           # 10 PM
  timezone: "Asia/Jakarta"
  enableTimeRestrictions: true
  errorBackoffMinutes: 30
  enableRecipientValidation: true
  validationCacheDuration: 24
  enableHealthCheck: true
  maxErrorRate: 0.3
```

### Frontend Configuration

The frontend automatically proxies API requests to the backend at `http://localhost:1234`.

To change the backend URL, edit `frontend/vite.config.ts`:

```typescript
server: {
  proxy: {
    '/api': {
      target: 'http://your-backend-url:1234',
      changeOrigin: true,
    },
  },
}
```

## 📖 Usage Guide

### 1. Connect a WhatsApp Session

1. Navigate to **Sessions** page
2. Click **Add Session**
3. Scan the QR code with WhatsApp mobile app
4. Wait for connection confirmation

### 2. Send Bulk Messages

1. Navigate to **Bulk Send** page
2. Select a connected session
3. Enter your message (use `{{variables}}` for personalization)
4. Add recipients (one phone number per line)
5. Click **Send to X Recipients**

### 3. Manage Recipients

1. Navigate to **Recipients** page
2. Add recipients manually or upload CSV/TXT file
3. Click **Validate All** to check if numbers are on WhatsApp
4. Export validated list as CSV

### 4. Create Message Templates

1. Navigate to **Templates** page
2. Click **New Template**
3. Enter template name and message
4. Use `{{name}}`, `{{phone}}`, or custom variables
5. Save and reuse in bulk sends

### 5. Configure Settings

1. Navigate to **Settings** page
2. Adjust anti-ban parameters
3. Configure time restrictions
4. Set daily limits
5. Click **Save Changes**

## 🔒 Anti-Ban Features

The system includes comprehensive anti-ban protection:

- ✅ **Sequential Sending** - One message at a time
- ✅ **Random Delays** - 20-60s between messages (configurable)
- ✅ **Batch Pauses** - 6-12 min breaks after every 8 messages
- ✅ **Daily Limits** - Maximum 40 messages per sender per day
- ✅ **Typing Simulation** - Sends "composing" presence before each message
- ✅ **Time Restrictions** - Only sends during business hours (8 AM - 10 PM)
- ✅ **Error Detection** - Automatically stops on rate limit errors
- ✅ **Health Monitoring** - Tracks session health and error rates
- ✅ **Recipient Validation** - Filters invalid numbers before sending
- ✅ **Proxy Support** - Different IP per session

## 📁 Project Structure

```
whatsapp_multi_session_with_proxy/
├── backend/                   # Go backend
│   ├── bulksender/           # Bulk send logic
│   ├── commandhandler/       # Business logic
│   ├── config/               # Configuration
│   ├── handler/              # HTTP handlers
│   ├── routers/              # API routes
│   ├── utils/                # Utilities
│   ├── main.go               # Backend entry point
│   └── config.prod.yaml      # Production config
├── frontend/                  # React frontend
│   ├── src/
│   │   ├── components/       # UI components
│   │   ├── pages/            # Page components
│   │   ├── services/         # API client
│   │   └── types/            # TypeScript types
│   └── package.json
├── bin/                       # Compiled binaries
└── README.md                  # This file
```

## 🐛 Troubleshooting

### Backend won't start
- Ensure port 1234 is not in use
- Check config file exists and is valid YAML
- Verify gcc/build-essential is installed (for SQLite)

### Frontend can't connect to backend
- Ensure backend is running on port 1234
- Check browser console for CORS errors
- Verify proxy configuration in `vite.config.ts`

### QR code won't display
- Ensure session is not already logged in
- Try disconnecting and reconnecting
- Check backend logs for errors

### Messages not sending
- Verify session is connected (green status)
- Check daily limit hasn't been reached
- Ensure current time is within allowed hours
- Review backend logs for rate limit errors

## 📝 API Documentation

See [CLAUDE.md](./CLAUDE.md) for complete API endpoint documentation.

## 🤝 Contributing

This is a private project. For issues or questions, contact the repository owner.

## 📄 License

Proprietary - All rights reserved

## ⚠️ Disclaimer

This tool is for authorized use only. Ensure compliance with WhatsApp's Terms of Service and applicable laws. The authors are not responsible for misuse or violations.
