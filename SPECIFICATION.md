# Composter - Decomposition Outliner Specification

## Purpose

Composter is a web-based outliner designed to help software developers discuss and compose problems into smaller and smaller pieces through hierarchical decomposition. The outliner paradigm provides an intuitive interface for breaking down complex problems using the principles of computational thinking.

## Computational Thinking Tenets

Composter reinforces the core tenets of computational thinking:

### 1. Decomposition
Breaking down complex problems into smaller, more manageable sub-problems. The hierarchical outliner structure naturally supports this by allowing users to create nested items.

### 2. Pattern Recognition
Identifying similarities and patterns across different parts of a problem. The outliner's structure helps visualize relationships and patterns through indentation and grouping.

### 3. Abstraction
Focusing on important information while ignoring irrelevant details. The collapse/expand functionality allows users to hide details and focus on high-level structure.

### 4. Algorithm Design
Creating step-by-step solutions. The ordered, hierarchical nature of outlines supports creating sequential processes and workflows.

## User Interface Commands

### Navigation and Editing

- **Enter** - Create a new line at the same indentation level
- **Tab** - Indent the current line and all its children (move right)
- **Shift+Tab** - Unindent the current line and all its children (move left)
- **Alt+Up** - Move the current item (and all children) up in the outline
- **Alt+Down** - Move the current item (and all children) down in the outline

### View Control

- **Shift+Click** - Toggle visibility of an item's children (collapse/expand)
- **Ctrl+Shift+Click** - Collapse all items with children or expand all items

### Document Management

- **Ctrl+S / Cmd+S** - Save the current outline
- **Save** button - Save the outline and continue editing
- **Close** button - Save the outline and return to the outline list

### Standard Operations

- **Ctrl+Z / Cmd+Z** - Undo (browser native)
- **Ctrl+Y / Cmd+Shift+Z** - Redo (browser native)

## Data Model

### Users
- **ID**: Unique identifier
- **Username**: Login name
- **Password**: Hashed password (bcrypt)
- **IsAdmin**: Administrative privileges flag
- **CreatedAt**: Account creation timestamp

### Outlines
- **ID**: Unique identifier
- **UserID**: Owner of the outline
- **Title**: Outline name
- **Content**: HTML-formatted outline structure with indentation
- **CreatedAt**: Creation timestamp
- **UpdatedAt**: Last modification timestamp

### Outline Structure
Outlines are stored as HTML with indentation represented by margin-left styling:
- Each line is a `<div>` element
- Indentation level is encoded as `margin-left: [level * 30]px`
- Two spaces in the editor = one indentation level
- Children are lines with greater indentation that appear consecutively after their parent

## Architecture

### Backend (Go)
- **Web Framework**: Standard library `net/http`
- **Database**: SQLite with `go-sqlite3` driver
- **Authentication**: Session-based with bcrypt password hashing
- **Middleware**: Authentication and admin authorization

### Frontend
- **Templates**: Go HTML templates
- **Styling**: Custom CSS with responsive design
- **Editor**: ContentEditable div with JavaScript for outline manipulation
- **State Management**: Client-side JavaScript with session storage for collapsed states

### File Structure
```
composter/
├── main.go                      # Application entry point
├── internal/
│   ├── database/
│   │   └── database.go          # Database models and operations
│   ├── handlers/
│   │   └── handlers.go          # HTTP request handlers
│   └── middleware/
│       └── auth.go              # Authentication middleware
├── templates/
│   ├── login.html               # Login page
│   ├── outlines.html            # Outline list page
│   ├── editor.html              # Outline editor
│   └── admin.html               # User management
└── static/
    └── style.css                # Application styling
```

## API Endpoints

### Authentication
- `GET /login` - Display login page
- `POST /login` - Authenticate user
- `GET /logout` - End user session

### Outlines
- `GET /` - List user's outlines
- `GET /editor` - Create new outline or edit existing (query param: `id`)
- `POST /api/outline/save` - Create or update outline
  - Request: `{id: int, title: string, content: string}`
  - Response: `{success: bool, id: int}`
- `POST /api/outline/delete` - Delete outline
  - Request: `{id: int}`
  - Response: `{success: bool}`

### Administration
- `GET /admin` - User management page (admin only)
- `POST /api/admin/user/create` - Create new user
  - Request: `{username: string, password: string, is_admin: bool}`
- `POST /api/admin/user/update` - Update existing user
  - Request: `{id: int, username: string, password: string, is_admin: bool}`
- `POST /api/admin/user/delete` - Delete user
  - Request: `{id: int}`

## Export Capabilities (Future)

The following export features are planned for future implementation:

### 1. GitHub Integration
- Export outline items as GitHub issues
- Create project board/kanban items
- Link outline structure to GitHub milestones
- OAuth authentication with GitHub

### 2. Export Formats
- **Markdown**: Hierarchical markdown format
- **JSON**: Structured data export
- **Plain Text**: Simple indented text
- **OPML**: Standard outline format

### 3. API for External Access
Future RESTful API for programmatic access:
- `GET /api/v1/outlines` - List outlines
- `GET /api/v1/outlines/{id}` - Get outline details
- `POST /api/v1/outlines` - Create outline
- `PUT /api/v1/outlines/{id}` - Update outline
- `DELETE /api/v1/outlines/{id}` - Delete outline
- `GET /api/v1/outlines/{id}/export?format={format}` - Export outline

## Security

- Passwords are hashed using bcrypt with default cost factor
- Session IDs are generated using cryptographic random number generator
- Session cookies are HTTP-only to prevent XSS attacks
- SQL injection prevention through parameterized queries
- User data isolation through user_id foreign key constraints

## Deployment

### Requirements
- Go 1.24.9 or later
- SQLite3
- Modern web browser with JavaScript enabled

### Running the Application
```bash
go build -o composter
./composter
```

The application starts on port 8080 by default.

Default credentials: `admin / admin`

### Database
- SQLite database file: `composter.db`
- Automatically initialized on first run
- Default admin user created if no users exist

## Design Principles

1. **Minimal Dependencies**: Uses Go standard library where possible
2. **Progressive Enhancement**: Core functionality works without JavaScript, enhanced features require it
3. **Data Ownership**: Each user's data is isolated and private
4. **Simplicity**: Clean, focused interface without distractions
5. **Keyboard-First**: All major operations accessible via keyboard shortcuts
6. **Hierarchical Thinking**: Interface reinforces tree-structured problem decomposition
