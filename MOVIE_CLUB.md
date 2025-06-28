# 🎬 Movie Club Feature

The Movie Club feature in Watcharr allows you and your friends to collectively decide what to watch together through a structured nomination and voting process. It operates on a cyclical 3-phase system that automates the entire process from movie suggestions to final selection.

## Overview

Movie Club follows a simple 3-week cycle:
1. **Week 1 - Nominations**: Users nominate movies they'd like to watch
2. **Week 2 - Voting**: Everyone votes on their favorite nominations  
3. **Week 3 - Watching**: Results are revealed with a clear winner

The cycle automatically transitions between phases and starts a new cycle when complete, ensuring your movie club always has something new to discover.

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
2. Click **Create New Cycle** (admin only)
3. Give it a name and description
4. The cycle will automatically start in the nomination phase

### For Users

#### Accessing Movie Club
- Click the **🎬 Film** icon in the navigation bar
- Or navigate directly to `/movie-club`

## How to Use Movie Club

### Phase 1: Nomination Week

During the nomination phase, you can suggest movies for the group to consider.

#### Nominating Movies
1. Click **"Nominate a Movie"** button
2. Search for a movie using the search modal
3. Select your chosen movie
4. Optionally add a reason why you're recommending it
5. Click **"Nominate This Movie"**

#### Managing Your Nominations
- View all your nominations in the "Your Nominations" section
- Remove nominations by clicking the **trash icon** (only during nomination phase)
- See all community nominations in the "All Nominations" section

**Note**: You can only nominate up to the configured limit (default: 1 movie per cycle).

### Phase 2: Voting Week

Once nominations close, the voting phase begins. Now everyone votes on their favorites from the nominated movies.

#### Casting Your Votes
1. Browse all nominated movies with their reasons
2. Click on movies to select them for voting
3. **Order matters!** Your votes are prioritized:
   - 1st choice = 3 points
   - 2nd choice = 2 points  
   - 3rd choice = 1 point (if available)

#### Managing Your Votes
- **Reorder votes**: Use the "Higher/Lower" buttons to change priorities
- **Remove votes**: Click a selected movie again to remove your vote
- **Clear all votes**: Use the "Clear Votes" button to start over
- **Submit**: Click "Submit Votes" when you're happy with your selections

**Note**: You can vote for up to the configured limit (default: 2 movies per cycle).

### Phase 3: Watching Week

The results are in! This phase reveals the winner and complete voting breakdown.

#### Viewing Results
- **Winner Spotlight**: The winning movie is prominently displayed with:
  - Total votes received
  - Weighted score
  - Number of first-choice votes
- **Complete Results Table**: See how every nominated movie performed
- **Vote Breakdown**: Detailed view of 1st, 2nd, and 3rd choice votes
- **Summary Statistics**: Total participation and voting metrics

#### During Watching Week
This is when your group watches the winning movie! The results stay visible while nominations for the next cycle can begin (phases overlap).

## Features and Benefits

### Automated Cycle Management
- **No manual intervention needed**: Phases transition automatically
- **Continuous cycles**: New cycles start immediately after the previous one ends
- **Overlapping phases**: Nomination and results phases run simultaneously for smooth transitions

### Fair Voting System
- **Weighted scoring**: Prioritized votes ensure fair representation of preferences
- **Transparent results**: Complete voting breakdown shows exactly how the winner was chosen
- **Tie-breaking**: System handles ties using first-choice votes and total votes

### User-Friendly Interface
- **Movie search integration**: Search uses the same movie database as your watchlist
- **Responsive design**: Works seamlessly on desktop and mobile
- **Real-time updates**: See phase transitions and time remaining
- **Activity tracking**: All movie club actions appear in your activity feed

### Flexible Configuration
Administrators can customize:
- Number of nominations per user
- Number of votes per user  
- Phase duration (days)
- Enable/disable the entire feature

## Understanding the Scoring System

Movie Club uses a weighted voting system to determine winners:

- **1st Choice Vote** = 3 points
- **2nd Choice Vote** = 2 points  
- **3rd Choice Vote** = 1 point

### Example
If "The Matrix" receives:
- 3 first-choice votes (3 × 3 = 9 points)
- 2 second-choice votes (2 × 2 = 4 points)  
- 1 third-choice vote (1 × 1 = 1 point)

**Total Score**: 14 points

In case of tied scores, the movie with more first-choice votes wins. If still tied, the movie with more total votes wins.

## Tips for a Great Movie Club Experience

### For Nominators
- **Write compelling reasons**: Help others understand why your movie is worth watching
- **Consider the group**: Think about movies that would appeal to most members
- **Diversify genres**: Mix up your nominations across cycles for variety

### For Voters  
- **Vote strategically**: Your first choice gets the most points, so prioritize carefully
- **Consider trying something new**: Give unfamiliar movies a chance
- **Vote for multiple options**: Use all your available votes to influence the outcome

### For Groups
- **Discuss nominations**: Talk about why certain movies were chosen
- **Plan watch parties**: Coordinate when and how you'll watch the winner
- **Respect the results**: Even if your favorite doesn't win, give the winner a fair chance

## Troubleshooting

### Common Issues

**"Movie club is not enabled"**
- Contact your administrator to enable the feature in server settings

**"No active movie club cycle"**  
- An administrator needs to create and start a new cycle

**"Nomination limit reached"**
- You've reached the maximum nominations allowed per cycle
- Wait for the next cycle to nominate again

**"Not currently in nomination/voting phase"**
- The current phase doesn't allow this action
- Check the phase status at the top of the page

**Movies not appearing in search**
- Only movies from the TMDB database can be nominated
- Try different search terms or check the movie's official title

### Getting Help

If you encounter issues:
1. Check the phase status and time remaining
2. Refresh the page to ensure you have the latest data
3. Contact your administrator for configuration issues
4. Report bugs to the Watcharr development team

## Technical Details

### Data Privacy
- Your nominations and votes are visible to all movie club members
- Voting is anonymous in results (only totals are shown, not individual votes)
- Activity tracking logs your movie club participation

### Performance
- Movie search uses cached data for fast results
- Phase transitions happen automatically via background tasks
- Results are calculated in real-time when viewing

### Integration
- Movie Club uses the same movie database as your personal watchlist
- Activities appear in your main activity feed
- Navigation integrates seamlessly with the main Watcharr interface

---

Ready to start your movie club? Contact your administrator to enable the feature and create your first cycle. Happy watching! 🍿