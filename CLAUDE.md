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
- `cd server && go test ./...` - Run all Go tests
- `cd server && go test -v -run "Matrix" .` - Run Matrix integration tests
- `cd server && go test -v -run "TestAppService" .` - Run Application Service tests
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
- `server/matrix.go` - Matrix/Dendrite integration core functionality with hybrid account system
- `server/matrix_api.go` - Matrix API endpoints and validation
- `server/matrix_rooms.go` - Matrix room management and user invitations
- `server/matrix_appservice.go` - Matrix Application Service implementation for virtual users

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

- **MatrixUser**: Legacy model for existing Matrix account links (auto-generated or custom)
- **MatrixUserV2**: Modern hybrid model supporting both Application Service and personal accounts
  - **Account Types**: `appservice` (virtual users) and `personal` (real Matrix accounts)
  - **AS Fields**: `ASManagedUser`, `CreatedViaAS`, `LastSeenAt` for Application Service users
  - **Legacy Compatibility**: Maintains backward compatibility with existing accounts
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

The Matrix integration provides community chat features for Movie Club cycles through self-hosted Dendrite servers with a hybrid account architecture supporting both Application Service virtual users and traditional Matrix accounts.

### Backend Implementation

**Core Matrix Files:**

- `server/matrix.go` - Matrix client initialization, hybrid user management, migration system, database models
- `server/matrix_api.go` - HTTP API endpoints, validation system, troubleshooting, hybrid account management
- `server/matrix_rooms.go` - Room creation, management, user invitations, space organization
- `server/matrix_appservice.go` - Application Service implementation for virtual Matrix users

**Hybrid Account Architecture:**

The Matrix integration supports two account types through a unified API:

1. **Application Service (AS) Accounts** (Default):
   - Virtual Matrix users managed entirely by Watcharr
   - No external Matrix client access required
   - Lightweight and automatic user management
   - Ideal for users who only need Movie Club chat features
   - No access tokens or passwords stored

2. **Personal Matrix Accounts** (Optional):
   - Real Matrix accounts that users can access via Element Web or other clients
   - Created via shared secret registration or custom account linking
   - Full Matrix ecosystem access beyond Watcharr
   - Requires Matrix server configuration for user creation

**User Creation Methods:**

1. **Application Service Virtual Users** (Default):
   - Creates virtual Matrix identities managed by Watcharr AS
   - No real Matrix user accounts on the server
   - Automatic lifecycle management (create/delete)
   - No external client access

2. **Shared Secret Registration** (Personal Accounts):
   - Uses Dendrite's Synapse-compatible shared secret endpoint
   - Creates real Matrix users with working access tokens
   - Requires `registrationSecret` configuration in Matrix settings
   - Users can immediately access Matrix features via Element Web

3. **Custom Account Linking** (Personal Accounts):
   - Links existing Matrix accounts to Watcharr users
   - Requires user to provide Matrix User ID and access token
   - Preserves existing Matrix identity and relationships

4. **Placeholder Mode** (Fallback):
   - Creates database entries with placeholder tokens
   - Used when neither AS nor shared secret is configured
   - Users cannot access actual Matrix features
   - Maintains feature compatibility for development

**Key Functions:**

**Core Management:**
- `InitializeMatrixClient()` - Creates Matrix client with admin credentials
- `GetOrCreateMatrixUserV2()` - Unified user creation using hybrid architecture (AS or personal)
- `UnlinkMatrixUserV2()` - Removes Matrix account links with account-type-specific handling

**Application Service Functions:**
- `InitializeAppService()` - Initializes AS manager with configuration
- `CreateUser()` (AS) - Creates virtual Matrix users via Application Service
- `DeleteUser()` (AS) - Removes AS-managed virtual users
- `ValidateConfig()` (AS) - Validates AS configuration and namespace setup

**Personal Account Functions:**
- `RegisterUserWithSharedSecret()` - Creates real Matrix users via shared secret registration
- `CreatePersonalMatrixUser()` - Creates real Matrix accounts with credentials
- `LinkCustomMatrixUser()` - Links existing Matrix accounts to Watcharr users
- `DeactivateMatrixUser()` - Deactivates real Matrix users on server (if supported)

**Migration and Legacy Support:**
- `MigrateMatrixUsersToV2()` - Migrates all legacy Matrix users to hybrid system
- `MigrateSingleMatrixUserToV2()` - Migrates individual legacy users with conflict detection

**Room and Space Management:**
- `CreateCycleRoom()` - Creates Matrix rooms when cycles enter watching phase
- `InviteUsersToRoom()` - Invites eligible users to cycle rooms (handles both account types)
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

The hybrid architecture provides differentiated user management based on account type:

**Application Service Accounts:**
- Virtual users created and managed entirely by Watcharr AS
- No real Matrix accounts created on the server
- Automatic lifecycle management (create/delete)
- No access tokens or credentials stored
- Immediate cleanup when unlinking accounts

**Personal Accounts:**
- Real Matrix accounts with full server presence
- Created via shared secret registration or custom linking
- Secure credential storage with encryption
- Matrix server deactivation on account removal (auto-generated only)
- Custom accounts preserved when unlinking from Watcharr

**Unified Unlinking Behavior:**
- **AS Accounts**: Immediate deletion from database, no server cleanup needed
- **Auto-generated Personal**: Deactivated on Matrix server and removed from Watcharr
- **Custom Personal**: Only removed from Watcharr, preserving the user's Matrix account
- **Legacy Accounts**: Migrated to V2 system or handled via legacy cleanup process

**Security Features:**
- Account-type-specific validation and permissions
- Secure token handling with encryption for personal accounts
- Hidden credential display in UI for security
- Namespace validation for AS accounts
- Migration safety with conflict detection

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

### Application Service Configuration

**Matrix Application Service Setup:**

The Application Service mode creates virtual Matrix users that exist only within the Matrix protocol but are fully managed by Watcharr. This provides a lightweight alternative to real Matrix accounts.

**Required Configuration:**

1. **Application Service Settings** (in Movie Club → Matrix → Application Service):
   - **AS ID**: Unique identifier for the Application Service (e.g., "watcharr-movieclub")
   - **AS Token**: Authentication token for AS-to-homeserver communication
   - **Homeserver Token**: Authentication token for homeserver-to-AS communication
   - **User Namespace**: Pattern for AS-managed users (e.g., "@watcharr_*:yourdomain.com")
   - **Alias Namespace**: Pattern for AS-managed room aliases (e.g., "#watcharr_*:yourdomain.com")
   - **Sender Localpart**: Bot user localpart for AS communications (e.g., "watcharr-bot")

2. **Dendrite Configuration** (dendrite.yaml):
```yaml
# Add Application Service registration
app_service_api:
  database:
    connection_string: "file:appservice.db"
  config_files:
    - "/path/to/watcharr-registration.yaml"
```

3. **AS Registration File** (watcharr-registration.yaml):
```yaml
id: "watcharr-movieclub"
url: "http://watcharr:8080"  # Watcharr server URL
as_token: "your-as-token-here"
hs_token: "your-hs-token-here"
sender_localpart: "watcharr-bot"
namespaces:
  users:
    - exclusive: true
      regex: "@watcharr_.*:yourdomain.com"
  aliases:
    - exclusive: true
      regex: "#watcharr_.*:yourdomain.com"
rate_limited: false
```

**Benefits of Application Service Mode:**

- **No Real User Accounts**: Virtual users don't consume Matrix server resources
- **Automatic Management**: User lifecycle completely handled by Watcharr
- **Simplified Setup**: No shared secret or user registration configuration needed
- **Clean Namespace**: All Movie Club users clearly identified by prefix
- **Event Processing**: AS receives all events for managed users and rooms

**AS HTTP Endpoints:**

Watcharr implements the Matrix Application Service API endpoints:

- `PUT /_matrix/app/v1/transactions/{txnId}` - Receives Matrix events
- `GET /_matrix/app/v1/users/{userId}` - User existence queries from homeserver
- `GET /_matrix/app/v1/rooms/{roomAlias}` - Room alias queries from homeserver

### Migration System

**Legacy User Migration:**

The hybrid system includes comprehensive migration support for existing Matrix users:

**Automatic Migration:**
- API endpoint: `POST /api/matrix/migrate-users-to-v2` (Admin only)
- Migrates all legacy `MatrixUser` records to `MatrixUserV2`
- Preserves user data, credentials, and preferences
- Soft-deletes legacy records after successful migration
- Provides detailed migration results and error reporting

**Migration Process:**
1. **Discovery**: Finds all legacy Matrix users in database
2. **Conflict Detection**: Checks for existing V2 accounts to prevent duplicates
3. **Data Migration**: Transfers user data with proper field mapping
4. **Account Type Assignment**: Sets account type to "personal" for legacy users
5. **Cleanup**: Soft-deletes legacy records after successful migration

**Migration Safety:**
- Atomic transactions ensure data integrity
- Failed migrations don't affect successful ones
- Comprehensive error logging and reporting
- No data loss during migration process
- Rollback capability through soft deletion

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
- Hybrid user invitation system supporting both AS virtual users and real Matrix accounts
- Automatic user invitation based on participation (nomination or voting)
- Space organization for easy navigation
- Account-type-aware invitation handling (AS users receive AS invitations, personal users receive Matrix invitations)

## Testing Architecture

### Matrix Integration Test Suite

The Matrix integration includes comprehensive test coverage for both the Application Service and hybrid account systems:

**Test Files:**

- `server/matrix_appservice_test.go` - Application Service functionality tests
- `server/matrix_hybrid_test.go` - Hybrid account system and migration tests

**Application Service Tests (`matrix_appservice_test.go`):**

1. **Configuration Validation** (7 test cases):
   - Valid AS configuration acceptance
   - Invalid/missing configuration rejection
   - Required field validation (ID, tokens, namespaces)

2. **User Management** (4 test cases):
   - AS user creation and lifecycle
   - Duplicate user prevention
   - User deletion and cleanup
   - Database consistency verification

3. **Namespace Validation** (8 test cases):
   - User ID namespace pattern matching
   - Room alias namespace validation
   - Malformed ID handling
   - Server name extraction from namespaces

4. **Authentication & Security** (4 test cases):
   - Bearer token validation
   - Invalid token rejection
   - Missing authentication handling
   - Authorization header parsing

5. **Event Processing** (4 test cases):
   - Matrix membership event handling
   - Message event processing
   - Unknown event type graceful handling
   - Malformed event error handling

6. **HTTP API Endpoints** (3 test cases):
   - AS route registration and accessibility
   - Transaction endpoint functionality
   - User/room query endpoint responses

7. **Lifecycle Management** (4 test cases):
   - AS manager start/stop functionality
   - Configuration validation during startup
   - Running status tracking
   - Error handling for invalid configurations

**Hybrid System Tests (`matrix_hybrid_test.go`):**

1. **Account Creation** (2 test cases):
   - AS account creation in AS-enabled mode
   - Personal account creation (integration test - skipped in unit tests)

2. **Account Unlinking** (4 test cases):
   - AS account unlinking and cleanup
   - Personal account unlinking with deactivation
   - Legacy account migration and unlinking
   - Non-existent account error handling

3. **Migration System** (4 test cases):
   - Bulk legacy user migration to V2 system
   - Migration with no legacy users
   - Single user migration with conflict detection
   - Migration error handling and rollback

4. **Personal Account Management** (1 test case):
   - Personal account creation (integration test - skipped in unit tests)

**Test Coverage:**

- **Database Operations**: Create, read, update, delete operations for all Matrix models
- **Error Handling**: Comprehensive error scenarios and edge cases
- **Integration Points**: API endpoints, middleware, and cross-service functionality
- **Security Validation**: Authentication, authorization, and token handling
- **Migration Safety**: Data integrity during legacy system migration
- **Account Type Handling**: Differentiated behavior for AS vs personal accounts

**Test Infrastructure:**

- In-memory SQLite databases for isolated testing
- Test helper functions for user, AS manager, and database setup
- Mock HTTP clients for API endpoint testing
- Comprehensive assertions with detailed error messages
- Cleanup procedures to prevent test interference

**Running Tests:**

```bash
# Run all Matrix tests
cd server && go test -v -run "Matrix" .

# Run Application Service tests only
cd server && go test -v -run "TestAppService" .

# Run hybrid system tests only
cd server && go test -v -run "TestGetOrCreateMatrixUserV2\|TestUnlinkMatrixUserV2\|TestMigrateMatrixUsersToV2" .

# Run specific test function
cd server && go test -v -run "TestAppServiceManager_ValidateConfig" .
```

## Common File Patterns

- **Svelte Components**: `.svelte` files with TypeScript script blocks
- **SvelteKit Routes**: `+page.svelte` (UI) + `+page.ts` (data loading)
- **Go Handlers**: Functions in main package handling HTTP endpoints
- **SCSS**: Global styles in `norm.scss`, component-specific in `.svelte` files
