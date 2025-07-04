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
- `src/types.ts` - TypeScript interfaces and type definitions

**Movie Club TypeScript Types:**

- `MovieClubCycleRating` - Interface for user ratings and thoughts
- `MovieClubCycleResponse` - Enhanced to include `cycleRatings` array
- Type safety across frontend/backend communication

### Backend (Go)

- **Framework**: Gin web framework
- **Database**: SQLite with GORM ORM
- **Authentication**: JWT-based with proxy auth support
- **Integrations**: TMDB, Plex, Jellyfin, Trakt, IGDB
- **Structure**: Single package with feature-based files

Key files:

- `server/watcharr.go` - Main application entry point and database migrations
- `server/routes.go` - HTTP route definitions
- `server/auth.go` - Authentication logic
- `server/content.go` - Content management
- `server/watched.go` - Watch tracking functionality with movie club integration
- `server/movie_club.go` - Movie club cycles, voting, and rating management

### Key Features

- **Content Tracking**: Movies, TV shows, anime, and games
- **Movie Club**: Collaborative movie selection with nomination, voting, and watching phases
  - **Cycle Ratings**: Persistent user ratings and thoughts for winning movies during watching phase
  - **Club Averages**: Aggregate ratings displayed when 2+ members have rated
  - **Eligibility System**: Only users who participated (nominated or voted) can rate cycle movies
  - **Real-time Updates**: Rating and thoughts updates during active cycles, locked when cycle ends
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
- Movie club cycles and nominations
- Movie club cycle ratings (user ratings/thoughts for winning movies)
- Tags and lists
- Activity tracking
- Integration configurations

### Movie Club Database Models

- **MovieClubCycle**: Defines nomination/voting/watching phases with dates and winner
- **MovieClubNomination**: User movie nominations for cycles
- **MovieClubVote**: User voting preferences (1st, 2nd, 3rd choice)
- **MovieClubCycleRating**: User ratings and thoughts for winning movies during watching phase
  - Composite unique index on (user_id, cycle_id) prevents duplicate ratings
  - Only eligible users (who nominated or voted) can create ratings
  - Supports ratings-only, thoughts-only, or combined entries

## External Services

- **TMDB**: Primary metadata source for movies/TV
- **IGDB**: Game metadata (requires Twitch app configuration)
- **Plex/Jellyfin**: Media server integration for automatic tracking
- **Trakt**: Import existing watch data
- **Sonarr/Radarr**: Request missing content

## Movie Club Feature Architecture

The movie club feature enables collaborative movie selection through structured cycles with three phases:

### Backend Implementation (`server/movie_club.go`)

**Core Functions:**

- `IsUserEligibleForCycleRating()` - Checks if user participated (nominated or voted) in cycle
- `GetActiveWatchingCyclesByWinnerContent()` - Finds active cycles for specific movie content
- `ProcessPotentialCycleRating()` - Handles rating capture with eligibility validation
- `CreateOrUpdateCycleRating()` - Creates or updates user ratings with proper error handling

**Integration Points:**

- `addWatched()` in `watched.go` - Automatically captures ratings when users rate winning movies
- `updateWatched()` in `watched.go` - Updates cycle ratings when watch entries are modified
- Auto-migration system in `watcharr.go` includes `MovieClubCycleRating` model

### Frontend Implementation

**Key Components:**

- `src/routes/(app)/movie-club/MovieClubResults.svelte` - Displays cycle ratings in active cycles
- `src/routes/(app)/movie-club/archives/+page.svelte` - Shows cycle ratings in completed cycles
- Both components share identical cycle ratings UI and styling

**Features:**

- Club average rating calculation (requires 2+ ratings)
- Golden asterisk styling matching TMDB ratings (`font-family: Rampart One`)
- Collapsible member ratings list with expandable thoughts
- Responsive design with mobile-optimized spacing
- Real-time updates during active cycles

### Performance Optimizations

**Staged Filtering Approach:**

1. Filter content by type (movies only)
2. Check user eligibility (participated in cycle)
3. Process rating updates for eligible combinations

**Database Efficiency:**

- Composite unique indexes prevent duplicate ratings
- Preloaded relationships reduce N+1 queries
- Conditional processing only when movie club is enabled

## Common File Patterns

- **Svelte Components**: `.svelte` files with TypeScript script blocks
- **SvelteKit Routes**: `+page.svelte` (UI) + `+page.ts` (data loading)
- **Go Handlers**: Functions in main package handling HTTP endpoints
- **SCSS**: Global styles in `norm.scss`, component-specific in `.svelte` files
