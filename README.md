# Composter

A decomposition outliner designed for software developers to break down complex problems using computational thinking principles.

## Overview

Composter is a web-based hierarchical outliner that helps developers decompose problems into smaller, manageable pieces. It provides a clean interface with keyboard-driven controls for manipulating outline structures, perfect for planning, brainstorming, and problem-solving.

## Features

- **Hierarchical Outlining**: Create nested structures to decompose complex problems
- **Item Movement**: Move items up/down while maintaining their children
- **Indentation Control**: Tab/Shift+Tab to adjust hierarchy levels
- **Collapse/Expand**: Focus on high-level structure by hiding details
- **Keyboard-First Design**: Efficient keyboard shortcuts for all operations
- **Multi-User Support**: User authentication and data isolation
- **Auto-Save**: Changes are preserved with Ctrl+S or manual save

## Quick Start

```bash
# Build the application
go build -o composter

# Run the server
./composter
```

The application starts on `http://localhost:8080`

Default credentials: `admin / admin`

## Testing

### Backend Tests (Go)
```bash
# Run all Go tests
go test ./...
```

### Frontend Tests (JavaScript)
```bash
# Install Node.js dependencies (first time only)
npm install

# Run frontend tests
npm test

# Run tests in watch mode
npm run test:watch

# Generate coverage report
npm run test:coverage
```

See [OUTLINER_TESTS.md](OUTLINER_TESTS.md) for detailed information about the frontend test suite.

## Documentation

For complete documentation, see [SPECIFICATION.md](SPECIFICATION.md)

## Keyboard Shortcuts

- **Tab** - Indent item (and children)
- **Shift+Tab** - Unindent item (and children)
- **Alt+Up** - Move item up (with children)
- **Alt+Down** - Move item down (with children)
- **Enter** - New line
- **Shift+Click** - Toggle children visibility
- **Ctrl+Shift+Click** - Collapse/Expand all
- **Ctrl+S** - Save outline

## License

See [LICENSE](LICENSE) file for details.

