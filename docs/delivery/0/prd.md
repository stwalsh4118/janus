# PBI-0: Project Structure & Development Environment Setup

[View in Backlog](../backlog.md#user-content-0)

## Overview

Establish the foundational project structure for both the Go backend API and Next.js frontend, with basic build tooling, development workflows, and minimal working endpoints. This enables incremental development and testing of both applications independently and together.

## Problem Statement

Before implementing features, we need:
- A well-organized Go project structure following best practices
- A Next.js application structure with proper configuration
- Build and run scripts for both applications
- A way to develop and test each component independently
- Basic integration between frontend and backend
- Development tooling configured (linting, formatting, etc.)

## User Stories

- As a developer, I want a clear project structure so that I know where to place new code
- As a developer, I want to run the backend independently so that I can test API endpoints
- As a developer, I want to run the frontend independently so that I can test UI components
- As a developer, I want hot reload during development so that I can see changes quickly
- As a developer, I want basic health endpoints so that I can verify the system is running

## Technical Approach

### Repository Structure

```
janus/
├── cmd/
│   └── server/
│       └── main.go              # Server entry point
├── internal/
│   ├── api/
│   │   ├── handlers/            # HTTP handlers
│   │   │   ├── health.go
│   │   │   └── session.go
│   │   ├── middleware/          # HTTP middleware
│   │   │   └── cors.go
│   │   └── router.go            # Route definitions
│   ├── session/
│   │   └── manager.go           # Session management (stub)
│   └── config/
│       └── config.go            # Configuration loading
├── frontend/
│   ├── app/
│   │   ├── layout.tsx
│   │   ├── page.tsx
│   │   └── api/                 # Next.js API routes (if needed)
│   ├── components/
│   │   └── PushToTalk.tsx       # Stub component
│   ├── lib/
│   │   └── api-client.ts        # Backend API client
│   ├── public/
│   ├── package.json
│   ├── tsconfig.json
│   ├── next.config.js
│   └── tailwind.config.js
├── docs/
│   └── delivery/                # PBI documentation (already exists)
├── scripts/
│   ├── setup.sh                 # Initial setup script
│   ├── dev-backend.sh           # Run backend in dev mode
│   ├── dev-frontend.sh          # Run frontend in dev mode
│   └── build.sh                 # Build both applications
├── .cursor-voice/               # Context directory (stub)
│   ├── .gitkeep
│   └── conversation-summaries/
├── go.mod
├── go.sum
├── .env.example
├── .gitignore
└── README.md
```

### Go Backend Setup

1. **Initialize Go Module** (Go 1.25.0+)
   ```bash
   go mod init github.com/sean/janus
   ```

2. **Basic Server Structure** (using Gin)
   - Gin router with graceful shutdown
   - CORS middleware configured
   - Health check endpoint: `GET /api/health`
   - Stub session endpoints (return mock data):
     - `POST /api/session/start`
     - `POST /api/ask?session_id={id}`
     - `POST /api/session/end?session_id={id}`
     - `POST /api/heartbeat?session_id={id}`
   - Install dependencies:
     ```bash
     go get -u github.com/gin-gonic/gin
     go get -u github.com/joho/godotenv
     go get -u github.com/gin-contrib/cors
     ```

3. **Configuration**
   - Load from environment variables
   - Support .env file via godotenv
   - Validation on startup

4. **Development Workflow**
   ```bash
   # Run with hot reload (air already installed globally)
   air
   ```

### Next.js Frontend Setup

1. **Initialize Next.js 15.5.4+** (latest stable)
   ```bash
   cd frontend
   pnpm create next-app@latest . --typescript --tailwind --app --no-src-dir
   ```
   - This will also install React 19.2.0+ (latest stable, October 2025)

2. **Basic Page Structure**
   - Home page with simple UI
   - Health check display from backend
   - Stub PushToTalk button component (non-functional)
   - API client with typed methods

3. **API Client**
   ```typescript
   class CursorVoiceClient {
     constructor(baseUrl: string);
     async healthCheck(): Promise<HealthResponse>;
     async startSession(): Promise<string>;
     // Stub methods for other endpoints
   }
   ```

4. **Development Workflow**
   ```bash
   pnpm dev  # Runs on port 3001 (backend uses 3000)
   ```

### Integration Points

1. **CORS Configuration**
   - Backend allows frontend origin (http://localhost:3001)
   - Configurable for Tailscale URLs later

2. **Environment Variables**
   - Backend: PORT, LOG_LEVEL
   - Frontend: NEXT_PUBLIC_API_URL

3. **Health Check Integration**
   - Frontend calls backend health endpoint on load
   - Displays connection status

### Development Scripts

1. **scripts/setup.sh**
   - Check Go and Node versions
   - Install dependencies (go mod download, pnpm install)
   - Create .env from .env.example
   - Create required directories

2. **scripts/dev-backend.sh**
   - Load .env
   - Run air for hot reload

3. **scripts/dev-frontend.sh**
   - cd to frontend
   - Run pnpm dev

4. **scripts/build.sh**
   - Build Go binary: `go build -o bin/janus cmd/server/main.go`
   - Build Next.js: `cd frontend && pnpm build`

### Testing Strategy

1. **Backend Testing**
   - Manual: curl commands for each endpoint
   - Verify health check returns 200
   - Verify stub session endpoints return mock data

2. **Frontend Testing**
   - Manual: Open browser, verify page loads
   - Verify health check displays backend status
   - Verify UI renders correctly

3. **Integration Testing**
   - Start both applications
   - Verify frontend can call backend health endpoint
   - Verify CORS works correctly

## UX/UI Considerations

### Initial Frontend UI
- Simple, clean landing page
- "System Status" section showing backend connectivity
- Placeholder for voice controls
- Mobile-responsive layout
- Dark mode ready (Tailwind setup)

### Developer Experience
- Fast setup (<5 minutes)
- Clear error messages for missing dependencies
- Hot reload for both applications
- Easy to run both apps simultaneously

## Acceptance Criteria

- [ ] Go module initialized with proper package name
- [ ] Directory structure created following Go best practices
- [ ] Backend HTTP server starts successfully on port 3000
- [ ] `GET /api/health` returns JSON with status and version
- [ ] Stub session endpoints return 200 with mock responses
- [ ] CORS middleware configured and functional
- [ ] Configuration loads from environment variables
- [ ] Next.js 15.5.4+ app initialized in frontend/ directory with React 19.2.0+
- [ ] Frontend runs successfully on port 3001
- [ ] Tailwind CSS configured and working
- [ ] API client implemented with health check method
- [ ] Frontend displays backend health status on load
- [ ] Frontend can successfully call backend API
- [ ] CORS allows frontend to call backend
- [ ] Hot reload works for backend (air)
- [ ] Hot reload works for frontend (Next.js dev)
- [ ] setup.sh script completes successfully
- [ ] dev-backend.sh starts backend with hot reload
- [ ] dev-frontend.sh starts frontend with hot reload
- [ ] build.sh produces working binaries
- [ ] .env.example file created with all variables documented
- [ ] .gitignore properly excludes build artifacts and env files
- [ ] README.md has quick start instructions
- [ ] .cursor-voice/ directory structure created

## Dependencies

- Go 1.25.0+ installed (latest stable, released August 2025)
- Node.js 22+ installed (LTS)
- pnpm installed ([[memory:2879566]])
- air installed globally for hot reload
- Git for version control

### Go Packages
- github.com/gin-gonic/gin (web framework)
- github.com/gin-contrib/cors (CORS middleware)
- github.com/joho/godotenv (environment variable loading)

## Open Questions

✅ **Resolved:**
- ~~Should we use a monorepo tool like Turborepo?~~ **No** - Keep it simple, no monorepo tooling
- ~~Do we need Docker setup at this stage?~~ **No** - Will set up containers later
- ~~Should we set up GitHub Actions CI/CD now or later?~~ **Later** - Focus on local dev for now
- ~~Do we want to use a Go web framework (Gin, Echo) or standard library?~~ **Yes, use Gin**

## Related Tasks

[View Task List](./tasks.md)

| Task ID | Task Name | Status |
|---------|-----------|--------|
| 0-1 | [Initialize Go backend with Gin framework](./0-1.md) | Proposed |
| 0-2 | [Initialize Next.js frontend application](./0-2.md) | Proposed |
| 0-3 | [Create development and build scripts](./0-3.md) | Proposed |
| 0-4 | [Create configuration files and documentation](./0-4.md) | Proposed |
| 0-5 | [E2E CoS Test - Verify full integration](./0-5.md) | Proposed |

