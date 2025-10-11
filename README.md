# Cursor Voice Chat

Voice-enabled interface for asking cursor-agent questions about your codebase while driving or away from your desk.

## 🎯 Features

- 🎤 **Voice Input** - Push-to-talk interface using Web Speech API
- 🔊 **Text-to-Speech** - Automatic audio responses
- 📱 **Mobile-Optimized** - Designed for phone use with large touch targets
- 🤖 **Cursor Integration** - Full cursor-agent integration with codebase context
- 💾 **Conversation History** - Automatic session summaries and context preservation
- 🔒 **Secure** - Access via Tailscale for private network access

## 📋 Prerequisites

- **Go** 1.25.0+ (latest stable)
- **Node.js** 22+ (LTS)
- **pnpm** (package manager)
- **air** (Go hot reload, install globally)
- **Git** for version control

## 🚀 Quick Start

### 1. Clone the Repository

```bash
git clone <repository-url>
cd janus
```

### 2. Configure Environment Variables

```bash
# Copy example environment file
cp .env.example .env

# Edit .env and set required values
# - PORT (default: 3000)
# - CURSOR_API_KEY (get from https://cursor.sh/settings - needed for PBI-2)
# - CODEBASE_PATH (path to your codebase - needed for PBI-2)
```

### 3. Start the Backend

```bash
cd api
go mod download
air
```

The backend will start on `http://localhost:3000`

### 4. Start the Frontend

```bash
# In a new terminal
cd web
pnpm install
pnpm dev
```

The frontend will start on `http://localhost:3001`

### 5. Open in Browser

Navigate to `http://localhost:3001` to use the application.

## 📁 Project Structure

```
janus/
├── api/                    # Go backend
│   ├── cmd/server/        # Main entry point
│   ├── internal/          # Internal packages
│   │   ├── api/          # HTTP handlers and router
│   │   ├── config/       # Configuration management
│   │   └── session/      # Session management
│   ├── go.mod            # Go dependencies
│   └── .air.toml         # Hot reload configuration
│
├── web/                    # Next.js frontend
│   ├── app/              # Next.js 15 App Router
│   ├── components/       # React components + shadcn/ui
│   ├── lib/              # Utilities and API client
│   └── package.json      # Node dependencies
│
├── docs/                   # Project documentation
│   └── delivery/         # PBI and task documentation
│
├── .cursor-voice/         # Context files for cursor integration
│   ├── conversation-summaries/  # Auto-generated summaries
│   └── README.md         # Context directory explanation
│
├── .env.example          # Environment variable template
├── .gitignore            # Git ignore rules
└── README.md             # This file
```

## 🛠️ Development

### Backend Development

```bash
cd api

# Install dependencies
go mod download

# Run with hot reload
air

# Build production binary
go build -o bin/janus ./cmd/server

# Run tests (when implemented)
go test ./...
```

### Frontend Development

```bash
cd web

# Install dependencies
pnpm install

# Run development server
pnpm dev

# Build for production
pnpm build

# Run production build
pnpm start

# Lint
pnpm lint
```

## 🧪 Testing

Testing will be implemented in PBI-0 task 0-5.

## 📚 Technology Stack

### Backend
- **Go 1.25.0** - Backend language
- **Gin** - Web framework
- **godotenv** - Environment variable loading
- **CORS** - Cross-origin resource sharing

### Frontend
- **Next.js 15.5.4** - React framework
- **React 19.1.0** - UI library
- **TypeScript 5.9.3** - Type safety
- **Tailwind CSS 4.1.14** - Styling
- **shadcn/ui** - Component library (54 components)
- **Web Speech API** - Voice input/output (will be implemented in PBI-4)

### Infrastructure
- **pnpm** - Fast package manager
- **air** - Go hot reload
- **Tailscale** - Secure network access (for remote use)

## 🗺️ Roadmap

Current implementation status (PBI-0 - Project Structure):

- ✅ **PBI-0 Task 0-1**: Go backend with Gin framework
- ✅ **PBI-0 Task 0-2**: Next.js frontend with shadcn/ui
- ⏳ **PBI-0 Task 0-3**: Development and build scripts
- 🔄 **PBI-0 Task 0-4**: Configuration files and documentation
- ⏳ **PBI-0 Task 0-5**: End-to-end testing

Upcoming features:

- **PBI-1**: Backend session management with cursor-agent
- **PBI-2**: Cursor-agent process integration
- **PBI-3**: Automatic context initialization
- **PBI-4**: Voice input/output with Web Speech API
- **PBI-5**: Conversation summaries and history
- **PBI-6**: Production configuration and deployment

## 🤝 Contributing

This project follows a strict task-driven development process. See `.cursorrules` for the complete development policy.

### Key Principles

1. All work must be associated with a task
2. Tasks must be linked to Product Backlog Items (PBIs)
3. No changes outside the scope of approved tasks
4. Full documentation and testing required

## 📝 License

MIT

## 🔗 Links

- [Cursor AI](https://cursor.sh/) - AI-powered code editor
- [Tailscale](https://tailscale.com/) - Secure network access
- [Project Documentation](./docs/delivery/) - Detailed PBI and task docs

## ⚡ Quick Tips

- Use **Tailscale** to access the app from your phone while driving
- Backend health check: `curl http://localhost:3000/api/health`
- Frontend and backend must both be running for full functionality
- Check `.env.example` for all available configuration options
- Conversation summaries will be stored in `.cursor-voice/conversation-summaries/` (implemented in PBI-5)

## 🐛 Troubleshooting

### Backend won't start
- Check if port 3000 is already in use: `lsof -i :3000`
- Verify Go is installed: `go version`
- Check environment variables in `.env`

### Frontend won't start
- Check if port 3001 is already in use
- Verify Node.js version: `node --version` (should be 22+)
- Clear cache: `rm -rf web/.next web/node_modules` and reinstall

### Can't connect from phone
- Ensure Tailscale is installed and running on both devices
- Update `NEXT_PUBLIC_API_URL` in `.env.local` to your Tailscale backend URL
- Check firewall settings

---

Built with ❤️ for developers who code on the go
