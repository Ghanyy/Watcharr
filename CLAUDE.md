# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Frontend (SvelteKit)

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run preview` - Preview production build
- `npm run check` - Run Svelte type checking
- `npm run check:watch` - Run type checking in watch mode
- `npm run lint` - Run linting (Prettier + ESLint)
- `npm run format` - Format code with Prettier

### Backend (Go)

- `npm run server` - Start Go server in development mode (equivalent to `cd ./server && MODE=DEV go run .`)
- `cd server && go run .` - Run Go server directly
- `cd server && go build` - Build Go binary
- `cd server && go test ./...` - Run Go tests

### Docker

- `docker-compose up` - Start full application stack
- `docker-compose -f docker-compose.dev.yml up` - Start development stack

## Architecture Overview

Watcharr is a full-stack web application for tracking watched movies, TV shows, and games. The architecture consists of:

### Frontend (SvelteKit)

- **Framework**: SvelteKit with TypeScript
- **Styling**: SCSS with custom mixins and variables
- **Structure**: File-based routing in `src/routes/`
- **Components**: Organized in `src/lib/` by feature (nav, poster, rating, etc.)
- **State**: Svelte stores in `store.svelte.ts`
- **API**: Axios-based HTTP client in `src/lib/util/api.ts`

Key directories:

- `src/routes/` - SvelteKit file-based routing
- `src/lib/` - Reusable components organized by feature
- `src/lib/util/` - Utility functions and API client

### Backend (Go)

- **Framework**: Gin web framework
- **Database**: SQLite with GORM ORM
- **Authentication**: JWT-based with proxy auth support
- **Integrations**: TMDB, Plex, Jellyfin, Trakt, IGDB
- **Structure**: Single package with feature-based files

Key files:

- `server/watcharr.go` - Main application entry point
- `server/routes.go` - HTTP route definitions
- `server/auth.go` - Authentication logic
- `server/content.go` - Content management
- `server/watched.go` - Watch tracking functionality

### Key Features

- **Content Tracking**: Movies, TV shows, anime, and games
- **External Integrations**: Plex/Jellyfin sync, Trakt import, Sonarr/Radarr requests
- **User Management**: Multi-user support with different permission levels
- **Import/Export**: CSV and Trakt data import
- **Social Features**: Following users, activity feeds

## Development Workflow

1. Frontend changes: Work in `src/` directory, use `npm run dev`
2. Backend changes: Work in `server/` directory, use `npm run server`
3. Full stack: Use Docker Compose for complete environment
4. Always run `npm run lint` before committing frontend changes
5. Check type safety with `npm run check`

## Database Schema

The Go backend uses GORM with SQLite. Key models are defined inline in the Go files:

- Users and authentication
- Watched content with ratings and status
- Tags and lists
- Activity tracking
- Integration configurations

## External Services

- **TMDB**: Primary metadata source for movies/TV
- **IGDB**: Game metadata (requires Twitch app configuration)
- **Plex/Jellyfin**: Media server integration for automatic tracking
- **Trakt**: Import existing watch data
- **Sonarr/Radarr**: Request missing content

## Common File Patterns

- **Svelte Components**: `.svelte` files with TypeScript script blocks
- **SvelteKit Routes**: `+page.svelte` (UI) + `+page.ts` (data loading)
- **Go Handlers**: Functions in main package handling HTTP endpoints
- **SCSS**: Global styles in `norm.scss`, component-specific in `.svelte` files
