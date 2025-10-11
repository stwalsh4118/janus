# PBI-6: Configuration & Setup

[View in Backlog](../backlog.md#user-content-6)

## Overview

Provide configuration management, environment setup tooling, and documentation to enable easy deployment and maintenance of the Cursor Voice Chat system. This includes environment variables, context file templates, setup scripts, and user documentation.

## Problem Statement

Users need to:
- Quickly set up the system on their development machine
- Configure the system for their specific codebase
- Understand how to create effective context files
- Troubleshoot common issues
- Maintain the system over time

## User Stories

- As a developer, I want clear setup instructions so that I can get the system running quickly
- As a developer, I want example context files so that I know what to write
- As a developer, I want environment variable documentation so that I can configure the system correctly
- As a developer, I want a health check endpoint so that I can verify the system is working
- As a developer, I want troubleshooting guidance so that I can resolve common issues myself

## Technical Approach

### Environment Variables

```bash
# Required
CURSOR_API_KEY=sk_xxxxx           # Cursor API key from settings
CODEBASE_PATH=/path/to/code       # Absolute path to codebase

# Optional (with defaults)
PORT=3000                         # Server port
SESSION_TIMEOUT_MINUTES=10        # Inactivity timeout
MAX_CONTEXT_SUMMARIES=3           # Past conversations to include
GIT_RECENT_DAYS=3                 # Days of recent file history
CONTEXT_DIR=.cursor-voice         # Context files location
```

### Configuration File Support

Optional `.cursor-voice.yaml` for advanced settings:

```yaml
server:
  port: 3000
  timeout_minutes: 10

cursor:
  api_key_file: ~/.cursor/api_key  # Alternative to env var
  timeout_seconds: 60

context:
  directory: .cursor-voice
  max_summaries: 3
  git_recent_days: 3

logging:
  level: info
  file: /var/log/cursor-voice.log
```

### Context File Templates

Provide templates in `docs/setup/templates/`:

1. **system-prompt.md** - Voice-specific instructions for cursor
2. **project-overview.md** - Project architecture and key areas
3. **active-context.md** - Current work focus

### Setup Script

`scripts/setup.sh`:
```bash
#!/bin/bash
# 1. Check prerequisites (Go, cursor-agent, Tailscale)
# 2. Create .cursor-voice directory
# 3. Copy template files
# 4. Prompt for environment variables
# 5. Create .env file
# 6. Verify cursor-agent works
# 7. Build Go server
# 8. Build Next.js frontend
# 9. Print Tailscale URL for access
```

### Health Check Endpoint

`GET /api/health` returns:
```json
{
  "status": "ok",
  "version": "1.0.0",
  "codebase_path": "/home/user/myproject",
  "cursor_available": true,
  "tailscale_ip": "100.64.0.5",
  "active_sessions": 2,
  "uptime_seconds": 3600,
  "context_files": {
    "system_prompt": true,
    "project_overview": true,
    "active_context": true,
    "conversation_summaries": 5
  }
}
```

### Documentation Structure

```
docs/
├── setup/
│   ├── README.md                 # Quick start guide
│   ├── prerequisites.md          # System requirements
│   ├── installation.md           # Step-by-step install
│   ├── tailscale-setup.md       # Tailscale configuration
│   ├── context-files.md         # Writing good context files
│   └── templates/
│       ├── system-prompt.md
│       ├── project-overview.md
│       └── active-context.md
├── usage/
│   ├── starting-session.md      # How to use
│   ├── best-practices.md        # Tips for good conversations
│   └── troubleshooting.md       # Common issues
└── development/
    ├── architecture.md          # System design
    ├── api-reference.md         # API documentation
    └── contributing.md          # Development guide
```

### Logging & Diagnostics

- Structured logging with levels (DEBUG, INFO, WARN, ERROR)
- Log rotation for long-running servers
- Request/response logging for debugging
- Process lifecycle events logged
- Performance metrics (response times, session duration)

## UX/UI Considerations

### Setup Experience
- Setup should complete in <10 minutes
- Clear error messages if prerequisites missing
- Validation of configuration before starting server
- Example context files should be high quality

### Operational Experience
- Health endpoint accessible via browser
- Logs clearly indicate problems
- Configuration errors fail fast with helpful messages

## Acceptance Criteria

- [ ] Environment variables are documented in README
- [ ] Example `.env` file is provided
- [ ] Configuration file schema is documented
- [ ] Setup script checks for prerequisites (Go, cursor, Tailscale)
- [ ] Setup script creates `.cursor-voice` directory structure
- [ ] Template context files are provided and documented
- [ ] Setup script creates `.env` with prompted values
- [ ] Setup script verifies cursor-agent is installed and accessible
- [ ] Health check endpoint returns comprehensive status
- [ ] Health check includes context file validation
- [ ] Documentation covers all setup steps clearly
- [ ] Troubleshooting guide covers common issues
- [ ] Best practices for context files are documented
- [ ] API endpoints are documented with examples
- [ ] Logging is implemented with appropriate levels
- [ ] Configuration validation fails fast with clear errors
- [ ] Server prints Tailscale URL on startup if available

## Dependencies

- All PBIs completed (this ties them together)
- Cursor CLI installation documentation
- Tailscale setup knowledge
- Access to test environment

## Open Questions

- Should we support multiple codebases per server?
- Do we need a web UI for configuration?
- Should we provide a Docker image for easier deployment?
- Do we need metrics/monitoring integration (Prometheus)?
- Should context file changes hot-reload or require restart?

## Related Tasks

Tasks will be created when PBI moves to "Agreed" status.

