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

### Security and Validation

- **Server-side Validation**: All limits and permissions enforced on backend
- **Frontend Prevention**: UI prevents invalid actions before server requests
- **Session Management**: Voting and nomination sessions persist across browser sessions
- **Data Integrity**: Comprehensive error handling prevents data corruption

---

Ready to start your movie club? Contact your administrator to enable the feature and create your first cycle. With support for multiple concurrent cycles, your community can always have something new to discover and vote on! 🍿
