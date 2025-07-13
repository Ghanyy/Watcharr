# 🎬 Movie Club Feature

The Movie Club feature in Watcharr allows you and your friends to collectively decide what to watch together through a structured nomination and voting process. It operates on a cyclical 3-phase system that guides the entire process from movie suggestions to final selection.

## Overview

Movie Club follows a simple 3-phase cycle:

1. **Nomination Phase**: Users nominate movies they'd like to watch
2. **Voting Phase**: Everyone votes on their favorite nominations using a priority-based system
3. **Watching Phase**: Results are revealed with a clear winner and detailed voting breakdown

The system supports **multiple active cycles** running simultaneously, allowing for continuous movie club activity and overlapping phases for smooth transitions.

## Getting Started

### For Administrators

#### 1. Enable Movie Club

Movie Club is disabled by default. To enable it:

1. Navigate to **Server Settings** (admin only)
2. Find the **Movie Club** section
3. Enable the feature and configure settings:
   - **Nominations per user**: How many movies each user can nominate (default: 1)
   - **Votes per user**: How many votes each user can cast (default: 2)
   - **Phase duration**: How long each phase lasts in days (default: 7)
4. Optionally enable **Community Features** for chat integration:
   - **Enable Movie Club Community**: Advanced navigation and community features
   - **Matrix Chat Integration**: Self-hosted chat rooms for cycle discussions

#### 2. Start Your First Cycle

1. Go to the Movie Club page
2. Click **"Start New Cycle"** button (admin only)
3. Give it a name and optional description
4. The cycle will automatically start in the nomination phase

#### 3. Managing Cycles

- **Create Multiple Cycles**: Run several movie club cycles simultaneously
- **Remove Cycles**: Hide cycles from the active list (data is preserved but cycle becomes inactive)
- **Monitor Progress**: View real-time phase progress and time remaining for all active cycles

### For Users

#### Accessing Movie Club

- Click the **🎬 Film** icon in the navigation bar
- Or navigate directly to `/movie-club`
- If community features are enabled, use the **Movie Club** dropdown menu to access:
  - **Dashboard**: Main movie club interface for nominations, voting, and results
  - **Community**: Chat rooms and community discussions for active cycles

## How to Use Movie Club

### Phase 1: Nomination Phase

During the nomination phase, you can suggest movies for the group to consider.

#### Nominating Movies

1. Click **"Nominate a Movie"** or **"Nominate Another Movie"** button
2. Use the search modal to find movies:
   - Type in the search box (searches TMDB database)
   - Browse results with poster thumbnails
   - Click on your chosen movie
3. Optionally add a reason for your nomination (up to 500 characters)
4. Click **"Nominate This Movie"**

#### Managing Your Nominations

- **View Your Nominations**: See all your nominations in the "Your Nominations" section with count/limit display
- **Remove Nominations**: Click the **trash icon** on any of your nominations (only during nomination phase)
- **View All Nominations**: See community nominations in the "All Nominations" section
- **Read Nomination Reasons**: Click "Show more" to expand long nomination explanations

**Note**: You can only nominate up to the configured limit (typically 1-2 movies per cycle). The system prevents nominations beyond your limit.

### Phase 2: Voting Phase

Once nominations close, the voting phase begins. Vote on your favorites using a priority-based ranking system.

#### How Voting Works

The voting system uses **priority-based ranking** where your vote order matters:

- **1st Choice**: Highest priority (typically worth 3 points)
- **2nd Choice**: Medium priority (typically worth 2 points)
- **3rd Choice**: Lower priority (typically worth 1 point)

#### Casting Your Votes

1. Browse all nominated movies with their nomination reasons
2. Click on movies to select them for voting
3. **Order matters!** Your first selection becomes your 1st choice, second becomes 2nd choice, etc.
4. Selected movies show priority badges (1st Choice, 2nd Choice, etc.)

#### Managing Your Votes

- **Reorder Priorities**: Use the "Higher/Lower" buttons to adjust movie rankings
- **Remove Votes**: Click a selected movie again to remove it from your votes
- **Clear All Votes**: Use the "Clear Votes" button to start over
- **Submit Votes**: Click "Submit Votes" when satisfied with your selections
- **View Vote Summary**: See your current vote rankings in the summary section

**Note**: You can vote for up to the configured limit (default: 2 movies per cycle).

### Phase 3: Watching Phase

The results are revealed! This phase shows the winner and complete voting breakdown.

#### Viewing Results

- **Winner Spotlight**: The winning movie is prominently displayed with:

  - Golden winner badge and styling
  - Total votes received
  - Weighted score (calculated from priority votes)
  - Number of first-choice votes
  - Movie overview and details

- **Complete Results Table**: Ranked list of all nominated movies showing:

  - Final ranking position (#1, #2, #3, etc.)
  - Movie title with direct links to movie details
  - Total votes received
  - Weighted score
  - Vote breakdown (number of 1st, 2nd, 3rd choice votes)

- **Voting Summary Statistics**:
  - Total votes cast across all movies
  - Number of movies nominated
  - Number of movies that received votes
  - Participation metrics

#### During Watching Phase

This is when your group watches the winning movie! Results remain visible while new cycles can be started, enabling continuous movie club activity.

**Community Features (if enabled):**

- **Chat Rooms**: Dedicated Matrix chat rooms are automatically created for the winning movie
- **Community Access**: Only users who participated (nominated or voted) can access the cycle's chat room
- **Persistent Discussions**: Chat rooms remain available even after the cycle ends
- **Real-time Communication**: Discuss the movie, coordinate watch parties, share reactions

## Features and Benefits

### Multi-Cycle Support

- **Concurrent Cycles**: Run multiple movie club cycles simultaneously
- **Continuous Activity**: Start new cycles while others are in progress
- **Independent Management**: Each cycle operates independently with its own phases and timelines

### Advanced Voting System

- **Priority-Based Ranking**: Your vote order determines point allocation
- **Weighted Scoring**: Sophisticated scoring system values higher-priority votes more
- **Transparent Results**: Complete voting breakdown shows exactly how winners were determined
- **Fair Representation**: Multiple votes per user ensure diverse preference representation

### Intuitive User Interface

- **Real-Time Search**: Instant movie search using TMDB integration
- **Visual Priority Management**: Drag-and-drop style priority controls
- **Responsive Design**: Seamless experience across desktop and mobile devices
- **Live Updates**: Real-time phase progress and countdown timers
- **Rich Movie Information**: Poster thumbnails, release dates, and overviews

### Smart Nomination System

- **Duplicate Prevention**: System tracks all nominations to prevent duplicates
- **Reason Support**: Users can explain their movie choices to influence voting
- **Limit Enforcement**: Automatic enforcement of nomination limits per user
- **Search Integration**: Full TMDB movie database access for nominations

### Administrative Control

- **Flexible Configuration**: Customize limits and phase durations
- **Cycle Management**: Create, monitor, and manage multiple cycles
- **User Oversight**: Monitor participation and voting patterns
- **Feature Toggle**: Enable/disable entire feature as needed
- **Community Settings**: Configure Matrix chat integration and validation tools
- **Matrix Management**: Comprehensive setup validation and troubleshooting guides

### Community Features (Matrix Integration)

Movie Club can integrate with self-hosted Matrix/Dendrite servers to provide dedicated chat rooms for cycle discussions.

#### Setup Requirements

- **Self-hosted Matrix server**: Dendrite recommended for local deployments
- **Admin configuration**: Matrix server URL, admin token, and user credentials
- **Community enablement**: Both Movie Club and Community features must be enabled
- **Application Service configuration**: For optimal performance and security (recommended)
- **Registration shared secret**: For real Matrix account creation (optional)

#### How Community Features Work

**Automatic Room Creation:**

- Chat rooms are created automatically when cycles enter the watching phase
- Room names include the date and movie title for easy identification
- Only users who participated (nominated or voted) are invited to the room

**Access Control:**

- **Participation-based access**: Only users who engaged with the cycle can join
- **Persistent access**: Rooms remain available even after cycles end
- **Local-only federation**: Rooms are private to your Dendrite instance
- **Admin moderation**: Watcharr administrators have full room management

**User Experience:**

- **Profile setup**: Users can create auto-generated Matrix accounts or link existing ones
- **Community page**: Centralized access to all available chat rooms
- **Mobile-friendly**: Responsive design works across all devices
- **External clients**: Rooms can be accessed through standard Matrix clients

#### Matrix Account Management

Users have three types of Matrix accounts available:

**Application Service Accounts (Recommended):**

- Virtual Matrix users managed entirely by Watcharr
- No external access - only work within Movie Club community features
- Automatic account creation and management
- Best performance and security for Watcharr-only usage

**Personal Accounts:**

*Auto-generated:*
- Real Matrix accounts created automatically by Watcharr
- Work with Element Web and other Matrix clients
- Secure credentials managed by the system
- Requires server configuration with registration shared secret

*Custom Linked:*
- Link existing Matrix accounts to Watcharr profiles
- Requires Matrix user ID and access token
- Full control over Matrix client and settings
- Preserves your existing Matrix account and settings

**Legacy Accounts:**

- Older account format from previous versions
- Automatically migrated to Personal accounts when accessed
- Consider updating to newer account types for better functionality

#### Administrator Tools

**Setup Validation:**

- Comprehensive validation of Matrix server configuration
- Real-time testing of connectivity, permissions, and room creation
- Detailed troubleshooting guide with step-by-step solutions
- Connection testing tools for quick diagnostics

**Configuration Options:**

- Matrix server URL and admin credentials
- Server name and space organization settings
- Admin user ID for room management
- Customizable space names for room organization
- Application Service settings for hybrid account management
- Registration shared secret for real Matrix account creation
- Account type preferences and migration tools

#### Privacy and Security

- **Local-only federation**: Rooms don't federate with external Matrix servers
- **Encrypted storage**: Matrix credentials are securely stored
- **Access logging**: Room access and membership tracked
- **Graceful degradation**: Movie Club works normally if Matrix is disabled

## Understanding the Voting System

Movie Club uses a **weighted priority voting system** to ensure fair and representative results:

### Scoring Mechanics

- **1st Choice Vote** = 3 points (highest priority)
- **2nd Choice Vote** = 2 points (medium priority)
- **3rd Choice Vote** = 1 point (lower priority)

### Example Calculation

If "The Matrix" receives:

- 3 first-choice votes (3 × 3 = 9 points)
- 2 second-choice votes (2 × 2 = 4 points)
- 1 third-choice vote (1 × 1 = 1 point)

**Total Weighted Score**: 14 points

### Tie-Breaking Rules

1. **Primary**: Movie with higher weighted score wins
2. **Secondary**: If scores are tied, movie with more first-choice votes wins
3. **Tertiary**: If still tied, movie with more total votes wins

## Tips for a Great Movie Club Experience

### For Nominators

- **Write compelling reasons**: Help others understand why your movie choice is worth watching
- **Consider the group**: Think about movies that would appeal to your community
- **Diversify your choices**: Mix genres and styles across different cycles
- **Use the full character limit**: Detailed reasons can influence voting decisions

### For Voters

- **Vote strategically**: Your first choice gets the most points, so prioritize carefully
- **Use all your votes**: Don't waste available voting power
- **Read the reasons**: Consider why movies were nominated before voting
- **Try something new**: Give unfamiliar movies a chance based on peer recommendations

### For Administrators

- **Monitor participation**: Ensure all members are engaging with the cycles
- **Adjust settings**: Fine-tune limits and durations based on group size and activity
- **Create regular cycles**: Maintain momentum with consistent new cycles
- **Communicate phases**: Keep members informed about phase transitions

### For Groups

- **Discuss nominations**: Talk about movie choices and reasons
- **Plan watch parties**: Coordinate viewing of winning movies
- **Respect the results**: Give winning movies a fair chance even if they weren't your choice
- **Share reactions**: Discuss movies after watching to build community
- **Use community chats**: Take advantage of Matrix rooms for real-time discussions
- **Coordinate viewing**: Use chat rooms to plan watch parties and viewing schedules

## Advanced Features

### Phase Management

- **Automatic Transitions**: Phases transition automatically based on configured durations
- **Time Tracking**: Real-time countdown shows exact time remaining in current phase
- **Phase Overlap**: New nominations can begin while previous results are still visible

### Data and Privacy

- **Vote Privacy**: Individual votes are private during voting phase
- **Results Transparency**: Only vote totals and breakdowns are shown in results
- **Activity Integration**: Movie club actions appear in your main Watcharr activity feed
- **Data Persistence**: Removed cycles preserve all data while hiding from active view

### Search and Discovery

- **TMDB Integration**: Access to comprehensive movie database for nominations
- **Smart Search**: Debounced search prevents API overload while providing instant results
- **Rich Metadata**: Movie posters, release dates, and overviews help inform decisions
- **Direct Links**: Results link directly to full movie detail pages in Watcharr

## Troubleshooting

### Common Issues

**"Movie club is not enabled"**

- Contact your administrator to enable the feature in server settings

**"No active movie club cycle"**

- An administrator needs to create and start a new cycle
- Multiple cycles can run simultaneously for continuous activity

**"Nomination limit reached"**

- You've reached the maximum nominations allowed per cycle
- Wait for a new cycle to nominate again, or participate in other active cycles

**"Not currently in nomination/voting phase"**

- The current phase doesn't allow this action
- Check the phase status and time remaining at the top of the page
- Consider participating in other active cycles that may be in different phases

**"Movies not appearing in search"**

- Only movies from the TMDB database can be nominated
- Try different search terms or check the movie's official title
- Ensure you're using the exact movie title as it appears in TMDB

**"Priority controls not working"**

- Ensure you have multiple votes selected before trying to reorder
- Priority can only be adjusted when you have 2 or more active votes
- Try refreshing the page if controls become unresponsive

**"Community features not available"**

- Administrator must enable both Movie Club and Community features
- Matrix integration requires additional server setup and configuration
- Check with administrator about Matrix server availability

**"Can't access community chats"**

- You must have participated in the cycle (nominated or voted) to access its chat room
- Set up your Matrix account in your profile settings first
- Chat rooms are only created when cycles reach the watching phase
- Check your account type - Application Service accounts work automatically
- Personal accounts may require additional setup or migration from legacy accounts

**"Matrix connection issues"**

- Administrators can use the Matrix Setup Validation tool in server settings
- Check the troubleshooting guide accessible from the validation modal
- Verify Matrix server is running and accessible from Watcharr server

### Getting Help

If you encounter issues:

1. Check the phase status and time remaining for the specific cycle
2. Refresh the page to ensure you have the latest data
3. Try participating in a different active cycle to isolate the issue
4. Contact your administrator for configuration or permissions issues
5. Report bugs to the Watcharr development team with specific error messages

## Technical Details

### Performance Optimizations

- **Debounced Search**: 300ms delay prevents excessive API calls during typing
- **Efficient Rendering**: Svelte reactivity ensures minimal re-renders
- **Cached Results**: Movie data and images are cached for faster loading
- **Real-time Updates**: Live phase progress without requiring page refreshes

### Data Integration

- **TMDB Synchronization**: Movie Club uses the same movie database as personal watchlists
- **Activity Tracking**: All movie club actions are logged in the main activity feed
- **User Preferences**: Voting patterns and participation history are tracked
- **Cross-feature Compatibility**: Nominated movies link to full Watcharr movie pages
- **Matrix Integration**: Community chat data stored separately with secure credentials
- **Room Management**: Chat room memberships and access tracked independently

### Security and Validation

- **Server-side Validation**: All limits and permissions enforced on backend
- **Frontend Prevention**: UI prevents invalid actions before server requests
- **Session Management**: Voting and nomination sessions persist across browser sessions
- **Data Integrity**: Comprehensive error handling prevents data corruption
- **Matrix Security**: Encrypted credential storage and secure room management
- **Access Control**: Participation-based chat access with automatic user management

---

Ready to start your movie club? Contact your administrator to enable the feature and create your first cycle. With support for multiple concurrent cycles and optional community chat integration, your community can always have something new to discover, vote on, and discuss together! 🍿

**New to Matrix/Dendrite?** Community features are completely optional - Movie Club works perfectly without them. But if you want to add real-time chat discussions to your movie club experience, administrators can easily set up Matrix integration with comprehensive validation tools and troubleshooting guides. The hybrid account system supports both simple Application Service accounts for Watcharr-only usage and real Matrix accounts for full Matrix client compatibility.
