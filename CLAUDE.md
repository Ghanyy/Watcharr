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
- `go mod tidy` - Update dependencies (includes Matrix SDK: maunium.net/go/mautrix v0.21.0)

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
- `src/lib/matrix/` - Matrix-specific components (validation, troubleshooting)
- `src/lib/nav/` - Navigation components including MovieClubMenu
- `src/lib/util/` - Utility functions and API client
- `src/types.ts` - TypeScript interfaces and type definitions

**Movie Club TypeScript Types:**

- `MovieClubCycleRating` - Interface for user ratings and thoughts
- `MovieClubCycleResponse` - Enhanced to include `cycleRatings` array
- `MovieClubSettings` - Movie club configuration including Matrix settings
- `MatrixSettings` - Matrix/Dendrite server configuration interface
- Type safety across frontend/backend communication

### Backend (Go)

- **Framework**: Gin web framework
- **Database**: SQLite with GORM ORM
- **Authentication**: JWT-based with proxy auth support
- **Integrations**: TMDB, Plex, Jellyfin, Trakt, IGDB, Matrix/Dendrite
- **Structure**: Single package with feature-based files

Key files:

- `server/watcharr.go` - Main application entry point and database migrations
- `server/routes.go` - HTTP route definitions
- `server/auth.go` - Authentication logic
- `server/content.go` - Content management
- `server/watched.go` - Watch tracking functionality with movie club integration
- `server/movie_club.go` - Movie club cycles, voting, and rating management
- `server/matrix.go` - Matrix/Dendrite integration core functionality
- `server/matrix_api.go` - Matrix API endpoints and validation
- `server/matrix_rooms.go` - Matrix room management and user invitations

### Key Features

- **Content Tracking**: Movies, TV shows, anime, and games
- **Movie Club**: Collaborative movie selection with nomination, voting, and watching phases
  - **Cycle Ratings**: Persistent user ratings and thoughts for winning movies during watching phase
  - **Club Averages**: Aggregate ratings displayed when 2+ members have rated
  - **Eligibility System**: Only users who participated (nominated or voted) can rate cycle movies
  - **Real-time Updates**: Rating and thoughts updates during active cycles, locked when cycle ends
  - **Community Chats**: Matrix/Dendrite integration for cycle-specific chat rooms
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
- Matrix integration (users, rooms, spaces, memberships)
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

### Matrix Integration Database Models

- **MatrixUser**: Links Watcharr users to Matrix accounts (auto-generated or custom)
- **MatrixRoom**: Stores Matrix rooms created for movie club cycles
- **MatrixRoomMember**: Tracks room memberships and join dates
- **MatrixSpace**: Stores Movie Club space information for room organization

## External Services

- **TMDB**: Primary metadata source for movies/TV
- **IGDB**: Game metadata (requires Twitch app configuration)
- **Plex/Jellyfin**: Media server integration for automatic tracking
- **Trakt**: Import existing watch data
- **Sonarr/Radarr**: Request missing content
- **Matrix/Dendrite**: Self-hosted chat server for Movie Club community features

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
- `TransitionCyclePhase()` - Creates Matrix rooms when cycles enter watching phase
- Auto-migration system in `watcharr.go` includes `MovieClubCycleRating` model

### Frontend Implementation

**Key Components:**

- `src/routes/(app)/movie-club/MovieClubResults.svelte` - Displays cycle ratings in active cycles
- `src/routes/(app)/movie-club/archives/+page.svelte` - Shows cycle ratings in completed cycles
- `src/routes/(app)/movie-club/community/+page.svelte` - Community chat access interface
- `src/lib/nav/MovieClubMenu.svelte` - Navigation dropdown with Dashboard/Community options
- Both rating components share identical cycle ratings UI and styling

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
- Asynchronous Matrix room creation to avoid blocking cycle transitions

## Matrix/Dendrite Integration Architecture

The Matrix integration provides community chat features for Movie Club cycles through self-hosted Dendrite servers with full user lifecycle management.

### Backend Implementation

**Core Matrix Files:**

- `server/matrix.go` - Matrix client initialization, user management, shared secret registration, database models
- `server/matrix_api.go` - HTTP API endpoints, validation system, troubleshooting
- `server/matrix_rooms.go` - Room creation, management, user invitations, space organization

**User Creation Methods:**

1. **Shared Secret Registration** (Recommended):
   - Uses Dendrite's Synapse-compatible shared secret endpoint
   - Creates real Matrix users with working access tokens
   - Requires `registrationSecret` configuration in Matrix settings
   - Users can immediately access Matrix features via Element Web

2. **Placeholder Mode** (Fallback):
   - Creates database entries with placeholder tokens
   - Used when shared secret is not configured
   - Users cannot access actual Matrix features
   - Maintains feature compatibility for development

**Key Functions:**

- `InitializeMatrixClient()` - Creates Matrix client with admin credentials
- `RegisterUserWithSharedSecret()` - Creates real Matrix users via shared secret registration
- `CreateMatrixUser()` - Auto-generates Matrix accounts (real or placeholder based on config)
- `LinkCustomMatrixUser()` - Links existing Matrix accounts to Watcharr users
- `UnlinkMatrixUser()` - Removes Matrix account links with proper deactivation for auto-generated accounts
- `DeactivateMatrixUser()` - Deactivates auto-generated users on Matrix server (if supported)
- `CreateCycleRoom()` - Creates Matrix rooms when cycles enter watching phase
- `InviteUsersToRoom()` - Invites eligible real Matrix users to cycle rooms (skips placeholder users)
- `EnsureMovieClubSpace()` - Creates and manages Movie Club space for room organization

**Integration Points:**

- Automatic room creation when cycles transition to watching phase
- User eligibility based on participation (nomination or voting)
- Room organization under Movie Club space
- Local-only federation for privacy
- Admin-controlled moderation with power levels
- Smart invitation system that only invites real Matrix users

### Frontend Implementation

**Core Components:**

- `src/lib/matrix/MatrixValidation.svelte` - Comprehensive setup validation modal
- `src/lib/matrix/MatrixTroubleshooting.svelte` - Detailed troubleshooting guide
- `src/routes/(app)/movie-club/community/+page.svelte` - Community chat interface
- `src/routes/(app)/profile/+page.svelte` - User Matrix account management

**Admin Configuration (`src/routes/(app)/server/+page.svelte`):**

- Matrix server URL, admin token, server name configuration
- Admin user ID and space name settings
- Registration shared secret for automatic user creation (optional)
- Connection testing and comprehensive validation
- Cascading disable functionality (disabling Movie Club disables Matrix)

**User Experience:**

- Auto-generated Matrix accounts or custom account linking
- Profile page Matrix account management
- Community page with room access and status
- Responsive design with mobile optimization

### Matrix Setup Validation System

**Validation Categories:**

1. **Configuration Validation** - Checks all Matrix settings and required fields
2. **Connection Testing** - Verifies Matrix server connectivity and authentication
3. **Permission Validation** - Tests admin privileges and server access capabilities
4. **Room Creation Testing** - Creates and cleans up test rooms to verify functionality
5. **Shared Secret Registration Testing** - Tests user creation via shared secret (if configured)

**Troubleshooting Guide:**

- Connection Issues (server access, network problems, firewall)
- Permission Issues (admin rights, token problems, user management)
- Room Creation Issues (resource problems, alias conflicts)
- Configuration Issues (server name mismatches, ID format errors)
- Dendrite-Specific Issues (startup problems, database connectivity)

**Features:**

- Real-time validation with color-coded status indicators
- Expandable details for each validation check
- Comprehensive troubleshooting with step-by-step solutions
- External documentation links to official Dendrite resources
- Auto-cleanup of test resources

### Security and Privacy

**Design Principles:**

- **Local-only federation** - Rooms are private to the Dendrite instance
- **Participation-based access** - Only users who nominated or voted can access cycle rooms
- **Admin-controlled moderation** - Watcharr admin has full room management privileges
- **Persistent chat history** - Rooms remain accessible after cycles end
- **Graceful degradation** - System functions normally when Matrix is disabled

**User Management:**

- Automatic Matrix user creation with secure credentials (real or placeholder based on configuration)
- Custom Matrix account linking with token validation
- Differentiated unlinking behavior:
  - Auto-generated accounts: Deactivated on Matrix server and removed from Watcharr
  - Custom accounts: Only removed from Watcharr, preserving the user's Matrix account
- Secure token handling with encryption and hidden display in UI

### Shared Secret Registration Configuration

**Dendrite Server Setup:**

To enable real Matrix user creation, configure your Dendrite server with a shared secret:

```yaml
# dendrite.yaml
user_api:
  registration_shared_secret: "your-secure-secret-here"
```

**Watcharr Configuration:**

1. Navigate to Server Settings → Movie Club → Matrix Configuration
2. Fill in the "Registration Shared Secret" field with the same secret from your Dendrite configuration
3. Run validation to test the registration endpoint
4. Users can now create real Matrix accounts that work with Element Web

**Benefits of Shared Secret Registration:**

- **Real Matrix Users**: Created users are actual Matrix accounts, not placeholders
- **Immediate Access**: Users can log into Element Web and other Matrix clients
- **Full Functionality**: Complete access to Matrix rooms and community features
- **Security**: Uses HMAC-SHA1 signatures for secure user creation
- **Compatibility**: Works with Dendrite's Synapse-compatible endpoint

### Room Management

**Room Lifecycle:**

1. Room created when cycle enters watching phase
2. Eligible users automatically invited
3. Room added to Movie Club space for organization
4. Persistent access throughout and after cycle
5. Room alias format: `#movieclub-YYYY-MM-DD-normalized-title:server.name`

**Room Features:**

- Private visibility with local-only federation
- Admin power levels for Watcharr management
- Movie-specific topics and descriptions
- Smart invitation system that only invites real Matrix users (skips placeholder users)
- Automatic user invitation based on participation (nomination or voting)
- Space organization for easy navigation

## Common File Patterns

- **Svelte Components**: `.svelte` files with TypeScript script blocks
- **SvelteKit Routes**: `+page.svelte` (UI) + `+page.ts` (data loading)
- **Go Handlers**: Functions in main package handling HTTP endpoints
- **SCSS**: Global styles in `norm.scss`, component-specific in `.svelte` files
