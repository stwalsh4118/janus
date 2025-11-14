# Making the Session Package Reusable

The session package has been moved from `internal/session` to `pkg/session`, making it importable by other packages and projects.

## Current Setup (Option 1: Same Module)

The package is now at `github.com/sean/janus/pkg/session` and can be imported by:

1. **Other packages in the same module** (already done)
2. **Other Go projects** that import your module

### Using in Other Projects

In another Go project, add to `go.mod`:

```bash
go get github.com/sean/janus/pkg/session
```

Then import:

```go
import "github.com/sean/janus/pkg/session"
```

**Note**: This requires your `github.com/sean/janus` repository to be publicly accessible (or accessible to the project using it).

## Option 2: Separate Module (For Maximum Reusability)

If you want to make it a completely standalone, versioned package, you can create a separate module:

### Steps:

1. **Create a new repository** (e.g., `github.com/sean/session`)

2. **Copy the package files** to the new repo:
   ```bash
   mkdir -p ~/session-package
   cp -r api/pkg/session/* ~/session-package/
   ```

3. **Create a new `go.mod`** in the new repo:
   ```bash
   cd ~/session-package
   go mod init github.com/sean/session
   ```

4. **Add dependencies**:
   ```bash
   go get github.com/rs/zerolog
   go get github.com/google/uuid
   ```

5. **Update the module name** in all files (if needed)

6. **Tag and release**:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

7. **Use in other projects**:
   ```bash
   go get github.com/sean/session@v1.0.0
   ```

## Option 3: Keep in Same Module but Use Go Workspaces

If you want to keep it in the same repo but make it easier to develop separately:

1. Create a `go.work` file in the root:
   ```go
   go 1.25.1

   use (
       ./api
       ./session-package  // if you create a separate module
   )
   ```

## Best Practices for Reusable Packages

1. **Minimal Dependencies**: Only depend on what's necessary (zerolog, uuid)
2. **Clear API**: Well-defined interfaces (like `Manager`)
3. **Documentation**: README with examples
4. **Tests**: Comprehensive test coverage
5. **Versioning**: Use semantic versioning (v1.0.0, v1.1.0, etc.)
6. **Backwards Compatibility**: Be careful with breaking changes

## Current Package Structure

```
pkg/session/
├── README.md           # Package documentation
├── USAGE.md           # This file
├── cleanup.go         # Cleanup service
├── cleanup_test.go    # Cleanup tests
├── constants.go       # Constants
├── interface.go       # Manager interface
├── memory_manager.go  # Implementation
├── memory_manager_test.go
└── types.go           # Session and Message types
```

## Dependencies

The package requires:
- `github.com/rs/zerolog` - For logging
- `github.com/google/uuid` - For session ID generation

These are standard dependencies and won't cause conflicts in most projects.

