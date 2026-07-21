# Unifidoc

> Unified API documentation generator for OpenAPI, gRPC, WebSocket, AsyncAPI, webhooks, events & MCP, one plugin-based CLI, one docs site

[![CI](https://github.com/sirrobot01/unifydoc/actions/workflows/ci.yml/badge.svg)](https://github.com/sirrobot01/unifydoc/actions/workflows/ci.yml)

Unifidoc is a powerful CLI tool that generates unified, professional documentation from multiple protocol specifications through a plugin-based architecture. It supports OpenAPI, gRPC, WebSocket, AsyncAPI, Webhooks, Events, MCP, and custom protocols.

## Features

- **Multi-Protocol Support**: Document REST, gRPC, WebSocket, AsyncAPI, Webhooks, Events, MCP, and custom protocols in one place
- **Plugin Architecture**: Extensible system with 9 built-in plugins
- **Code Examples**: Auto-generated code examples in cURL, JavaScript, Python, and Go
- **Client-Side Search**: Fast `⌘K` search across every protocol
- **Light & Dark Themes**: Built-in theme toggle, preference is remembered
- **Live Preview**: Development server with live reload
- **Zero Config**: Auto-detection of specification files
- **Modern UI**: Responsive three-column reference layout (nav · content · code panel)

## Installation

### Prebuilt Binaries

Download a binary for your platform from the
[Releases](https://github.com/sirrobot01/unifydoc/releases) page (Linux, macOS,
and Windows; amd64 and arm64).

### Using Go

```bash
go install github.com/sirrobot01/unifydoc/cmd/unifidoc@latest
```

### From Source

```bash
git clone https://github.com/sirrobot01/unifydoc
cd unifydoc
make install
```

### Using Make

```bash
make build
```

## Quick Start

### 1. Initialize a New Project

```bash
unifidoc init
```

This creates a `unifidoc.yaml` configuration file. You can also use auto-detection:

```bash
unifidoc init --auto-detect
```

### 2. Generate Documentation

```bash
unifidoc generate
```

### 3. Preview Locally

```bash
unifidoc serve
```

Open `http://localhost:8080` in your browser.

## Configuration

### Basic Configuration (`unifidoc.yaml`)

```yaml
project:
  name: "My API Documentation"
  version: "1.0.0"
  description: "Complete API documentation"

output:
  dir: "./docs"
  format: "html"
  theme: "default"

protocols:
  - plugin: openapi
    spec: ./specs/api.yaml
    enabled: true
  - plugin: grpc
    spec: ./specs/service.proto
    enabled: true
  - plugin: websocket
    spec: ./specs/ws.yaml
    enabled: true

features:
  search: true
  interactive: false
  darkMode: true
  codeSnippets:
    languages: ["curl", "javascript", "python", "go"]

server:
  port: 8080
  livereload: true
```

## Supported Protocols

### 1. OpenAPI (REST APIs)

```yaml
- plugin: openapi
  spec: ./api/openapi.yaml
  enabled: true
```

Supports OpenAPI 3.x specifications.

### 2. gRPC

```yaml
- plugin: grpc
  spec: ./protos/service.proto   # a single file, or a directory
  enabled: true
```

Parses Protocol Buffer (`.proto`) files. `spec` may point at a single file or a
**directory** — a directory compiles every `.proto` together, so imports
between files (e.g. `import "common.proto";`) resolve. A single file also
resolves imports relative to its own directory.

> **Directory specs** work for every protocol: if `spec` is a directory, all
> matching spec files in it (searched **recursively**) are parsed and merged
> into one section. For gRPC that means joint compilation with import
> resolution; for the YAML-based protocols it merges `*.yaml`, `*.yml` and
> `*.json` files.

### 3. WebSocket

```yaml
- plugin: websocket
  spec: ./specs/websocket.yaml
  enabled: true
```

Custom YAML format for WebSocket events. Example:

```yaml
websocket:
  url: wss://api.example.com/ws
  description: Real-time chat
  events:
    - name: message.send
      direction: send
      description: Send a message
      payload:
        type: object
        properties:
          text:
            type: string
```

### 4. AsyncAPI

```yaml
- plugin: asyncapi
  spec: ./specs/asyncapi.yaml
  enabled: true
```

Supports AsyncAPI 2.x/3.x for event-driven architectures.

### 5. Webhooks

```yaml
- plugin: webhook
  spec: ./specs/webhooks.yaml
  enabled: true
```

### 6. Events

```yaml
- plugin: events
  spec: ./specs/events.yaml
  enabled: true
```

### 7. MCP (Model Context Protocol)

```yaml
- plugin: mcp
  spec: ./specs/mcp.yaml
  enabled: true
```

### 8. API (Simple REST)

```yaml
- plugin: api
  spec: ./specs/api.yaml
  enabled: true
```

Simplified REST format without full OpenAPI complexity.

### 9. Custom

```yaml
- plugin: custom
  spec: ./specs/custom.yaml
  enabled: true
```

User-defined protocols with flexible schema.

## CLI Commands

### `unifidoc init`

Initialize a new Unifidoc project.

**Flags:**
- `--auto-detect`: Auto-detect specification files
- `--framework <name>`: Framework-specific init (express, fastapi, gin, spring)
- `--ci github`: Also generate a GitHub Actions workflow that builds and deploys docs
- `--with-hooks`: Install git pre-commit (validate) and pre-push (regenerate) hooks

### `unifidoc generate`

Generate documentation from specifications.

**Flags:**
- `--config <file>`: Specify config file
- `--output <dir>`: Override output directory
- `--watch`: Watch for changes and regenerate

### `unifidoc validate`

Validate specification files.

**Flags:**
- `--config <file>`: Specify config file
- `--strict`: Strict validation mode

### `unifidoc serve`

Start development server with live reload. Watches your spec files and
regenerates the docs on change, refreshing the browser automatically.

**Flags:**
- `--port <n>`: Server port (default: 8080)
- `--no-open`: Don't open browser automatically
- `--no-watch`: Disable file watching and live reload

### `unifidoc version`

Show version information (version, Go toolchain, OS/architecture).

### `unifidoc plugin list`

List all available plugins.

**Flags:**
- `--builtin`: Show only built-in plugins
- `--installed`: Show only installed plugins

### `unifidoc plugin add <plugin>`

Add a built-in protocol plugin to `unifidoc.yaml`, pointing it at a spec file.

**Flags:**
- `--spec <file>`: Path to the spec file (required)
- `--template <file>`: Optional custom template
- `--disabled`: Add the protocol but leave it disabled

### `unifidoc import`

Import a specification from another tool into your project.

**Flags:**
- `--from <tool>`: Source tool — `postman` (converts a v2.x collection) or `openapi`
- `--spec <file>`: Path to the source file (required)
- `--output <file>`: Output spec file (Postman only; defaults to `<name>.api.yaml`)

## Examples

Check the `examples/` directory for sample specifications:

- `examples/openapi/petstore.yaml` - OpenAPI Pet Store API
- `examples/websocket/chat.yaml` - WebSocket Chat API
- `examples/grpc/user.proto` - gRPC User Service

## Development

### Setup

```bash
make dev-setup
```

### Build

```bash
make build
```

### Run Tests

```bash
make test
```

### Format Code

```bash
make fmt
```

### Generate Example Docs

```bash
make example
```

## Architecture

Unifidoc uses a plugin-based architecture:

1. **CLI Layer**: Command parsing with Cobra
2. **Config Manager**: Configuration loading with Viper
3. **Plugin Manager**: Dynamic plugin discovery and loading
4. **Protocol Plugins**: 9 built-in parsers
5. **IR Engine**: Common intermediate representation
6. **Template Engine**: HTML rendering
7. **Generator**: Documentation assembly
8. **Dev Server**: Live preview with hot-reload

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see LICENSE file for details

## Support

- **Issues**: [GitHub Issues](https://github.com/sirrobot01/unifydoc/issues)
- **Documentation**: See `docs/` directory

## Project Status

Unifidoc is under active development. Current parser maturity:

- **OpenAPI**: full OpenAPI 3.x support (via `kin-openapi`).
- **WebSocket, Webhooks, Events, MCP, API, Custom** — supported via Unifidoc's
  YAML schemas.
- **gRPC / AsyncAPI**: simplified parsers; being deepened toward full `.proto`
  message resolution and full AsyncAPI 2.x/3.x support.

Code examples are currently HTTP-oriented and are being made protocol-aware
(e.g. `grpcurl` for gRPC, WebSocket clients for events).

## Roadmap

- [ ] Protocol-aware code examples (gRPC, WebSocket, events)
- [ ] Full `.proto` parsing with message resolution
- [ ] Full AsyncAPI 2.x/3.x support
- [ ] GraphQL support
- [ ] PDF export
- [ ] Postman collection export
- [ ] VS Code extension
