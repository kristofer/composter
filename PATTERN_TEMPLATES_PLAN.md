# Pattern Templates Implementation Plan

## Overview

Pattern Templates will help software engineers quickly start decomposing problems using proven patterns for common software architectures and workflows. This feature enables both pre-built system templates and custom user-created templates.

## Phase 1: Data Model & Storage (Foundation)

**Goal**: Create the infrastructure to store and manage templates

**Tasks**:
- Add `templates` table to database schema
  - id, name, description, content, category, is_system, user_id, created_at, updated_at
- Add `Template` struct to database models
- Implement CRUD operations for templates
  - CreateTemplate, GetTemplate, GetAllTemplates, UpdateTemplate, DeleteTemplate
  - GetSystemTemplates (shared/built-in)
  - GetUserTemplates (custom user-created)
- Add template category enum/constants (MVC, Microservices, API, DataPipeline, etc.)

**Deliverable**: Database layer ready to store templates

---

## Phase 2: Pre-built System Templates (Content)

**Goal**: Create valuable default templates users can use immediately

**Tasks**:
- Design outline structures for core patterns:
  - **MVC Application** (Models, Views, Controllers breakdown)
  - **REST API Design** (endpoints, auth, validation, error handling, docs)
  - **Microservice Architecture** (service boundaries, communication, deployment)
  - **Data Pipeline** (ingestion, transformation, validation, storage, monitoring)
  - **Feature Development** (requirements, design, implementation, testing, deployment)
  - **Bug Fix Process** (reproduce, diagnose, fix, test, deploy)
- Create seed data/migration to populate system templates
- Add template initialization to `db.Init()`

**Deliverable**: 6-8 ready-to-use system templates

---

## Phase 3: Template UI - Browsing & Using (Read-Only UX)

**Goal**: Let users discover and instantiate templates

**Tasks**:
- Create template browser page (`/templates`)
  - List view with categories
  - Search/filter by category or name
  - Template preview showing structure
- Add "New from Template" button to outlines list page
- Implement template instantiation
  - API endpoint: `POST /api/template/instantiate`
  - Copy template content to new outline
  - Redirect to editor with new outline
- Update navigation to include templates link

**Deliverable**: Users can browse and use system templates

---

## Phase 4: Custom Template Creation (Write UX)

**Goal**: Enable users to create their own reusable templates

**Tasks**:
- Add "Save as Template" feature to editor
  - Button in editor UI
  - Modal dialog for template metadata (name, description, category)
  - API endpoint: `POST /api/template/create`
- Implement "My Templates" view
  - Filter to show only user's custom templates
  - Edit template metadata
  - Delete custom templates
- Add template ownership/permissions
  - Users can only edit/delete their own templates
  - System templates are read-only

**Deliverable**: Users can create custom templates from outlines

---

## Phase 5: Template Sharing & Export (Collaboration)

**Goal**: Allow teams to share templates

**Tasks**:
- Implement template export
  - Export to JSON format
  - Download template file
  - API endpoint: `GET /api/template/{id}/export`
- Implement template import
  - Upload JSON template file
  - Validate structure
  - Import as user template
  - API endpoint: `POST /api/template/import`
- Add "Share Template" feature
  - Generate shareable template link (optional)
  - Template library page for organization (future)

**Deliverable**: Templates are portable and shareable

---

## Phase 6: Advanced Features (Enhancement)

**Goal**: Make templates more powerful and intelligent

**Tasks**:
- Template variables/placeholders
  - Support `{{PROJECT_NAME}}`, `{{API_VERSION}}` in template content
  - Prompt user for values during instantiation
- Template versioning
  - Track template versions
  - Update existing templates
  - Roll back to previous versions
- Template analytics
  - Track usage count
  - Most popular templates
  - User favorites
- Template tags for better organization
  - Multi-dimensional categorization
  - User-defined tags

**Deliverable**: Templates are dynamic and easier to discover

---

## Implementation Order

**MVP (Phases 1-3)**: Users can browse and use system templates
**Full Feature (Phases 1-5)**: Users can create and share custom templates
**Advanced (Phase 6)**: Enhanced features and intelligence

## Current Status

- [x] Phase 1: Data Model & Storage
- [x] Phase 2: Pre-built System Templates
- [x] Phase 3: Template UI - Browsing & Using
- [x] Phase 4: Custom Template Creation
- [x] Phase 5: Template Sharing & Export
- [ ] Phase 6: Advanced Features
