package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// MovieClubPhase represents the current phase of a movie club cycle
type MovieClubPhase string

const (
	PHASE_NOMINATION MovieClubPhase = "nomination"
	PHASE_VOTING     MovieClubPhase = "voting"
	PHASE_WATCHING   MovieClubPhase = "watching"
)

// MovieClubCycle represents a complete cycle of nomination -> voting -> watching
type MovieClubCycle struct {
	GormModel
	Name              string          `json:"name"`
	Description       string          `json:"description"`
	Phase             MovieClubPhase  `json:"phase"`
	PhaseStartDate    time.Time       `json:"phaseStartDate"`
	PhaseEndDate      time.Time       `json:"phaseEndDate"`
	NominationEndDate time.Time       `json:"nominationEndDate"`
	VotingEndDate     time.Time       `json:"votingEndDate"`
	WatchingEndDate   time.Time       `json:"watchingEndDate"`
	WinnerContentID   *int            `json:"winnerContentId,omitempty"`
	WinnerContent     *Content        `json:"winnerContent,omitempty" gorm:"foreignKey:WinnerContentID;references:ID"`
	Active            bool            `json:"active"`
	Nominations       []MovieClubNominationGroup `json:"nominations,omitempty" gorm:"-"`
	AllNominations    []MovieClubNomination `json:"-" gorm:"foreignKey:CycleID"`
	Votes             []MovieClubVote `json:"votes,omitempty" gorm:"foreignKey:CycleID"`
}

// MovieClubNomination represents a user's nomination for a movie
type MovieClubNomination struct {
	GormModel
	CycleID     uint     `json:"cycleId" gorm:"index"`
	Cycle       MovieClubCycle `json:"cycle,omitempty" gorm:"foreignKey:CycleID"`
	UserID      uint     `json:"userId" gorm:"index"`
	User        User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	ContentID   int      `json:"contentId" gorm:"index"`
	Content     Content  `json:"content,omitempty" gorm:"foreignKey:ContentID;references:ID"`
	Reason      string   `json:"reason" gorm:"type:text"`
}

// MovieClubVote represents a user's vote for a nominated movie
type MovieClubVote struct {
	GormModel
	CycleID     uint     `json:"cycleId" gorm:"index"`
	Cycle       MovieClubCycle `json:"cycle,omitempty" gorm:"foreignKey:CycleID"`
	UserID      uint     `json:"userId" gorm:"index"`
	User        User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	ContentID   int      `json:"contentId" gorm:"index"`
	Content     Content  `json:"content,omitempty" gorm:"foreignKey:ContentID;references:ID"`
	Priority    int      `json:"priority"` // 1 = first choice, 2 = second choice, etc.
}

// MovieClubSettings holds configurable parameters for the movie club system
type MovieClubSettings struct {
	NominationsPerUser int `json:"nominationsPerUser"` // Default: 1
	VotesPerUser       int `json:"votesPerUser"`       // Default: 2
	PhaseDurationDays  int `json:"phaseDurationDays"`  // Default: 7 (1 week)
	Enabled            bool `json:"enabled"`           // Default: false
}

// MovieClubVoteCount represents vote tallies for a content item
type MovieClubVoteCount struct {
	ContentID    int       `json:"contentId"`
	Content      Content   `json:"content"`
	TotalVotes   int       `json:"totalVotes"`
	FirstChoice  int       `json:"firstChoice"`
	SecondChoice int       `json:"secondChoice"`
	ThirdChoice  int       `json:"thirdChoice"`
	WeightedScore float64  `json:"weightedScore"` // Calculated score based on vote priorities
	NominatedAt  time.Time `json:"nominatedAt"`   // When this content was nominated (for tie-breaking)
}

// MovieClubNominationRequest represents the request to nominate a movie
type MovieClubNominationRequest struct {
	ContentID int    `json:"contentId" binding:"required"`
	Reason    string `json:"reason" binding:"max=500"`
	CycleID   *uint  `json:"cycleId,omitempty"` // Optional: if not provided, uses first active cycle
}

// MovieClubVoteRequest represents the request to cast votes
type MovieClubVoteRequest struct {
	Votes   []MovieClubVoteItem `json:"votes" binding:"required,dive"`
	CycleID *uint               `json:"cycleId,omitempty"` // Optional: if not provided, uses first active cycle
}

type MovieClubVoteItem struct {
	ContentID int `json:"contentId" binding:"required"`
	Priority  int `json:"priority" binding:"required,min=1"`
}

// MovieClubCycleRating represents a user's rating and thoughts for a winning movie in a specific cycle
type MovieClubCycleRating struct {
	GormModel
	CycleID     uint            `json:"cycleId" gorm:"index;uniqueIndex:cycle_user_content"`
	Cycle       MovieClubCycle  `json:"cycle,omitempty" gorm:"foreignKey:CycleID"`
	UserID      uint            `json:"userId" gorm:"index;uniqueIndex:cycle_user_content"`
	User        User            `json:"user,omitempty" gorm:"foreignKey:UserID"`
	ContentID   int             `json:"contentId" gorm:"index;uniqueIndex:cycle_user_content"`
	Content     Content         `json:"content,omitempty" gorm:"foreignKey:ContentID;references:ID"`
	Rating      float64         `json:"rating" gorm:"type:numeric(2,1)"`
	Thoughts    string          `json:"thoughts" gorm:"type:text"`
}

// MovieClubNominationGroup represents a movie with all its nominations grouped together
type MovieClubNominationGroup struct {
	ContentID   int                     `json:"contentId"`
	Content     Content                 `json:"content"`
	Nominations []MovieClubNomination   `json:"nominations"`
	Nominators  []User                  `json:"nominators"`
	Reasons     []string                `json:"reasons"`
}

// MovieClubCycleResponse represents the full cycle data for frontend
type MovieClubCycleResponse struct {
	Cycle          MovieClubCycle       `json:"cycle"`
	UserNominations []MovieClubNomination `json:"userNominations"`
	UserVotes      []MovieClubVote      `json:"userVotes"`
	VoteResults    []MovieClubVoteCount `json:"voteResults,omitempty"`
	CycleRatings   []MovieClubCycleRating `json:"cycleRatings,omitempty"`
	CanNominate    bool                 `json:"canNominate"`
	CanVote        bool                 `json:"canVote"`
}

// Helper methods for MovieClubCycle

// IsNominationPhase checks if the cycle is currently in nomination phase
func (c *MovieClubCycle) IsNominationPhase() bool {
	return c.Phase == PHASE_NOMINATION
}

// IsVotingPhase checks if the cycle is currently in voting phase
func (c *MovieClubCycle) IsVotingPhase() bool {
	return c.Phase == PHASE_VOTING
}

// IsWatchingPhase checks if the cycle is currently in watching phase
func (c *MovieClubCycle) IsWatchingPhase() bool {
	return c.Phase == PHASE_WATCHING
}

// ShouldTransitionPhase checks if the current phase should transition to the next
func (c *MovieClubCycle) ShouldTransitionPhase() bool {
	now := time.Now()
	return now.After(c.PhaseEndDate)
}

// GetNextPhase returns the next phase in the cycle
func (c *MovieClubCycle) GetNextPhase() MovieClubPhase {
	switch c.Phase {
	case PHASE_NOMINATION:
		return PHASE_VOTING
	case PHASE_VOTING:
		return PHASE_WATCHING
	case PHASE_WATCHING:
		return PHASE_NOMINATION
	default:
		return PHASE_NOMINATION
	}
}

// Database helper functions

// GetActiveMovieClubCycles returns all currently active movie club cycles with proper sorting
func GetActiveMovieClubCycles(db *gorm.DB) ([]MovieClubCycle, error) {
	slog.Debug("GetActiveMovieClubCycles: Looking for active cycles")
	
	var cycles []MovieClubCycle
	result := db.Where("active = ?", true).
		Preload("WinnerContent").
		Preload("AllNominations.Content").
		Preload("AllNominations.User").
		Preload("AllNominations", func(db *gorm.DB) *gorm.DB {
			return db.Joins("JOIN contents ON movie_club_nominations.content_id = contents.id").
				Order("contents.title ASC")
		}).
		Find(&cycles)
	
	if result.Error != nil {
		slog.Error("GetActiveMovieClubCycles: Database error", "error", result.Error)
		return nil, result.Error
	}
	
	// Group nominations for each cycle
	for i := range cycles {
		cycles[i].Nominations = GroupNominationsByContent(cycles[i].AllNominations)
	}
	
	// Sort cycles according to the rules:
	// 1. Watching phase cycles first, sorted by time left to end (ascending)
	// 2. Other cycles sorted by time left to start watching phase (ascending)
	SortActiveMovieClubCycles(cycles)
	
	slog.Debug("GetActiveMovieClubCycles: Found active cycles", "count", len(cycles))
	return cycles, nil
}

func GetArchivedMovieClubCycles(db *gorm.DB) ([]MovieClubCycle, error) {
	slog.Debug("GetArchivedMovieClubCycles: Looking for archived cycles")
	
	var cycles []MovieClubCycle
	result := db.Where("active = ? AND phase = ? AND winner_content_id IS NOT NULL", false, PHASE_WATCHING).
		Preload("WinnerContent").
		Preload("AllNominations.Content").
		Preload("AllNominations.User").
		Preload("AllNominations", func(db *gorm.DB) *gorm.DB {
			return db.Joins("JOIN contents ON movie_club_nominations.content_id = contents.id").
				Order("contents.title ASC")
		}).
		Order("watching_end_date DESC").
		Find(&cycles)
	
	if result.Error != nil {
		slog.Error("GetArchivedMovieClubCycles: Database error", "error", result.Error)
		return nil, result.Error
	}
	
	// Group nominations for each cycle
	for i := range cycles {
		cycles[i].Nominations = GroupNominationsByContent(cycles[i].AllNominations)
	}
	
	slog.Debug("GetArchivedMovieClubCycles: Found archived cycles", "count", len(cycles))
	return cycles, nil
}

// GetActiveMovieClubCycle returns the currently active movie club cycle (backwards compatibility)
// Now returns the first cycle from the sorted list of active cycles
func GetActiveMovieClubCycle(db *gorm.DB) (*MovieClubCycle, error) {
	cycles, err := GetActiveMovieClubCycles(db)
	if err != nil {
		return nil, err
	}
	
	if len(cycles) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	
	return &cycles[0], nil
}

// GetMovieClubNominationsForCycle returns all nominations for a specific cycle
func GetMovieClubNominationsForCycle(db *gorm.DB, cycleID uint) ([]MovieClubNomination, error) {
	var nominations []MovieClubNomination
	result := db.Where("cycle_id = ?", cycleID).
		Preload("User").
		Preload("Content").
		Joins("JOIN contents ON movie_club_nominations.content_id = contents.id").
		Order("contents.title ASC").
		Find(&nominations)
	
	if result.Error != nil {
		return nil, result.Error
	}
	
	return nominations, nil
}

// GroupNominationsByContent groups nominations by movie content
func GroupNominationsByContent(nominations []MovieClubNomination) []MovieClubNominationGroup {
	contentMap := make(map[int]*MovieClubNominationGroup)
	
	for _, nomination := range nominations {
		contentID := nomination.ContentID
		
		if group, exists := contentMap[contentID]; exists {
			// Add to existing group
			group.Nominations = append(group.Nominations, nomination)
			group.Nominators = append(group.Nominators, nomination.User)
			if nomination.Reason != "" {
				group.Reasons = append(group.Reasons, nomination.Reason)
			}
		} else {
			// Create new group
			reasons := []string{}
			if nomination.Reason != "" {
				reasons = append(reasons, nomination.Reason)
			}
			
			contentMap[contentID] = &MovieClubNominationGroup{
				ContentID:   contentID,
				Content:     nomination.Content,
				Nominations: []MovieClubNomination{nomination},
				Nominators:  []User{nomination.User},
				Reasons:     reasons,
			}
		}
	}
	
	// Convert map to slice
	var groups []MovieClubNominationGroup
	for _, group := range contentMap {
		groups = append(groups, *group)
	}
	
	// Sort groups by content title for consistent ordering
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Content.Title < groups[j].Content.Title
	})
	
	return groups
}

// SortActiveMovieClubCycles sorts cycles according to the specified rules:
// 1. Watching phase cycles first, sorted by time left to end (ascending - less time left = higher priority)
// 2. Other cycles sorted by time left to start watching phase (ascending - less time to watching = higher priority)
func SortActiveMovieClubCycles(cycles []MovieClubCycle) {
	now := time.Now()
	
	sort.Slice(cycles, func(i, j int) bool {
		cycleA := cycles[i]
		cycleB := cycles[j]
		
		// Both cycles in watching phase - sort by time left to end (ascending)
		if cycleA.IsWatchingPhase() && cycleB.IsWatchingPhase() {
			timeLeftA := cycleA.PhaseEndDate.Sub(now)
			timeLeftB := cycleB.PhaseEndDate.Sub(now)
			return timeLeftA < timeLeftB
		}
		
		// Only A is in watching phase - A comes first
		if cycleA.IsWatchingPhase() && !cycleB.IsWatchingPhase() {
			return true
		}
		
		// Only B is in watching phase - B comes first
		if !cycleA.IsWatchingPhase() && cycleB.IsWatchingPhase() {
			return false
		}
		
		// Both cycles NOT in watching phase - sort by time left to start watching phase
		timeToWatchingA := cycleA.WatchingEndDate.Sub(now)
		timeToWatchingB := cycleB.WatchingEndDate.Sub(now)
		
		// For cycles in nomination/voting phase, calculate time until watching phase starts
		if cycleA.IsNominationPhase() {
			// Time until watching = time to finish nomination + voting duration + time from watching start to end
			timeToWatchingA = cycleA.VotingEndDate.Sub(now)
		} else if cycleA.IsVotingPhase() {
			// Time until watching = time to finish voting + time from watching start to end  
			timeToWatchingA = cycleA.VotingEndDate.Sub(now)
		}
		
		if cycleB.IsNominationPhase() {
			timeToWatchingB = cycleB.VotingEndDate.Sub(now)
		} else if cycleB.IsVotingPhase() {
			timeToWatchingB = cycleB.VotingEndDate.Sub(now)
		}
		
		return timeToWatchingA < timeToWatchingB
	})
}

// GetMovieClubVotesForCycle returns all votes for a specific cycle
func GetMovieClubVotesForCycle(db *gorm.DB, cycleID uint) ([]MovieClubVote, error) {
	var votes []MovieClubVote
	result := db.Where("cycle_id = ?", cycleID).
		Preload("User").
		Preload("Content").
		Order("priority ASC").
		Find(&votes)
	
	if result.Error != nil {
		return nil, result.Error
	}
	
	return votes, nil
}

// GetUserNominationsForCycle returns a user's nominations for a specific cycle
func GetUserNominationsForCycle(db *gorm.DB, cycleID uint, userID uint) ([]MovieClubNomination, error) {
	var nominations []MovieClubNomination
	result := db.Where("cycle_id = ? AND user_id = ?", cycleID, userID).
		Preload("Content").
		Find(&nominations)
	
	if result.Error != nil {
		return nil, result.Error
	}
	
	return nominations, nil
}

// GetUserVotesForCycle returns a user's votes for a specific cycle
func GetUserVotesForCycle(db *gorm.DB, cycleID uint, userID uint) ([]MovieClubVote, error) {
	var votes []MovieClubVote
	result := db.Where("cycle_id = ? AND user_id = ?", cycleID, userID).
		Preload("Content").
		Order("priority ASC").
		Find(&votes)
	
	if result.Error != nil {
		return nil, result.Error
	}
	
	return votes, nil
}

// CalculateVoteResults calculates vote counts and determines winner for a cycle
func CalculateVoteResults(db *gorm.DB, cycleID uint) ([]MovieClubVoteCount, error) {
	// Get all nominations for the cycle to ensure we include all options
	nominations, err := GetMovieClubNominationsForCycle(db, cycleID)
	if err != nil {
		return nil, err
	}
	
	// Get all votes for the cycle
	votes, err := GetMovieClubVotesForCycle(db, cycleID)
	if err != nil {
		return nil, err
	}
	
	// Create a map to track vote counts
	voteCountMap := make(map[int]*MovieClubVoteCount)
	
	// Initialize with nominations
	for _, nomination := range nominations {
		voteCountMap[nomination.ContentID] = &MovieClubVoteCount{
			ContentID: nomination.ContentID,
			Content:   nomination.Content,
			TotalVotes: 0,
			FirstChoice: 0,
			SecondChoice: 0,
			ThirdChoice: 0,
			WeightedScore: 0,
			NominatedAt: nomination.CreatedAt,
		}
	}
	
	// Count votes
	for _, vote := range votes {
		if count, exists := voteCountMap[vote.ContentID]; exists {
			count.TotalVotes++
			
			switch vote.Priority {
			case 1:
				count.FirstChoice++
				count.WeightedScore += 3 // First choice gets 3 points
			case 2:
				count.SecondChoice++
				count.WeightedScore += 2 // Second choice gets 2 points
			case 3:
				count.ThirdChoice++
				count.WeightedScore += 1 // Third choice gets 1 point
			}
		}
	}
	
	// Convert map to slice and sort by weighted score
	var results []MovieClubVoteCount
	for _, count := range voteCountMap {
		results = append(results, *count)
	}
	
	// Sort by weighted score (descending), then by first choice votes, then by total votes, then by nomination time (ascending - earlier nominations win)
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].WeightedScore < results[j].WeightedScore ||
				(results[i].WeightedScore == results[j].WeightedScore && results[i].FirstChoice < results[j].FirstChoice) ||
				(results[i].WeightedScore == results[j].WeightedScore && results[i].FirstChoice == results[j].FirstChoice && results[i].TotalVotes < results[j].TotalVotes) ||
				(results[i].WeightedScore == results[j].WeightedScore && results[i].FirstChoice == results[j].FirstChoice && results[i].TotalVotes == results[j].TotalVotes && results[i].NominatedAt.After(results[j].NominatedAt)) {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
	
	return results, nil
}

// getMovieClubSettings returns the movie club configuration
func (b *BaseRouter) getMovieClubSettings(c *gin.Context) {
	// Check if movie club is enabled
	if !Config.MOVIE_CLUB.Enabled {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Movie club is not enabled"})
		return
	}
	
	c.JSON(http.StatusOK, Config.MOVIE_CLUB)
}

// addMovieClubActivity creates an activity record for movie club actions
func addMovieClubActivity(db *gorm.DB, userID uint, activityType ActivityType, data string) error {
	// Create a dummy watched entry with ID 0 for movie club activities
	// This allows movie club activities to fit into the existing activity system
	activity := Activity{
		UserID:    userID,
		WatchedID: 0, // Special case for movie club activities
		Type:      activityType,
		Data:      data,
	}
	
	if err := db.Create(&activity).Error; err != nil {
		slog.Error("Failed to create movie club activity", "error", err, "type", activityType)
		return err
	}
	
	slog.Debug("Added movie club activity", "userID", userID, "type", activityType)
	return nil
}

// API Routes

// addMovieClubRoutes adds all movie club related routes
func (b *BaseRouter) addMovieClubRoutes() {
	movieClub := b.rg.Group("/movie-club").Use(AuthRequired(b.db))
	
	// Get current cycle and user data
	movieClub.GET("/current", b.getCurrentMovieClubCycle)
	
	// Get all active cycles with user data
	movieClub.GET("/cycles/active", b.getActiveMovieClubCycles)
	
	// Get all archived cycles (completed, non-deleted)
	movieClub.GET("/cycles/archived", b.getArchivedMovieClubCycles)
	
	// Nomination endpoints
	movieClub.POST("/nominate", b.nominateMovie)
	movieClub.DELETE("/nominate/:id", b.removeNomination)
	
	// Voting endpoints
	movieClub.POST("/vote", b.voteForMovies)
	movieClub.DELETE("/vote", b.clearVotes)
	
	// Results endpoint
	movieClub.GET("/results", b.getMovieClubResults)
	
	// Settings endpoint (public for all users)
	movieClub.GET("/settings", b.getMovieClubSettings)
	
	// Admin endpoints
	movieClub.POST("/cycle", AdminRequired(), b.createMovieClubCycle)
	movieClub.PUT("/cycle/:id", AdminRequired(), b.updateMovieClubCycle)
	movieClub.DELETE("/cycle/:id", AdminRequired(), b.deleteMovieClubCycle)
	movieClub.POST("/cycle/:id/transition", AdminRequired(), b.transitionCyclePhase)
	movieClub.GET("/cycles", AdminRequired(), b.getAllMovieClubCycles)
}

// getCurrentMovieClubCycle returns the current active cycle with user-specific data
func (b *BaseRouter) getCurrentMovieClubCycle(c *gin.Context) {
	// Check if movie club is enabled
	if !Config.MOVIE_CLUB.Enabled {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Movie club is not enabled"})
		return
	}
	
	userID := c.GetUint("userId")
	
	// Get active cycle
	cycle, err := GetActiveMovieClubCycle(b.db)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Debug("No active movie club cycle found")
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "No active movie club cycle"})
			return
		}
		slog.Error("Failed to get active movie club cycle", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get cycle"})
		return
	}
	
	slog.Debug("Found active movie club cycle", "cycleId", cycle.ID, "phase", cycle.Phase)
	
	// Get user nominations
	userNominations, err := GetUserNominationsForCycle(b.db, cycle.ID, userID)
	if err != nil {
		slog.Error("Failed to get user nominations", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get nominations"})
		return
	}
	
	// Get user votes
	userVotes, err := GetUserVotesForCycle(b.db, cycle.ID, userID)
	if err != nil {
		slog.Error("Failed to get user votes", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get votes"})
		return
	}
	
	// Calculate abilities
	canNominate := cycle.IsNominationPhase() && len(userNominations) < Config.MOVIE_CLUB.NominationsPerUser
	canVote := cycle.IsVotingPhase() && len(userVotes) < Config.MOVIE_CLUB.VotesPerUser
	
	// Get vote results if in watching phase
	var voteResults []MovieClubVoteCount
	if cycle.IsWatchingPhase() {
		voteResults, err = CalculateVoteResults(b.db, cycle.ID)
		if err != nil {
			slog.Error("Failed to calculate vote results", "error", err)
		}
	}
	
	// Get cycle ratings if in watching phase
	var cycleRatings []MovieClubCycleRating
	if cycle.IsWatchingPhase() {
		cycleRatings, err = GetCycleRatingsForCycle(b.db, cycle.ID)
		if err != nil {
			slog.Error("Failed to get cycle ratings", "error", err)
		}
	}
	
	response := MovieClubCycleResponse{
		Cycle:           *cycle,
		UserNominations: userNominations,
		UserVotes:       userVotes,
		VoteResults:     voteResults,
		CycleRatings:    cycleRatings,
		CanNominate:     canNominate,
		CanVote:         canVote,
	}
	
	c.JSON(http.StatusOK, response)
}

// getActiveMovieClubCycles returns all active cycles with user-specific data
func (b *BaseRouter) getActiveMovieClubCycles(c *gin.Context) {
	// Check if movie club is enabled
	if !Config.MOVIE_CLUB.Enabled {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Movie club is not enabled"})
		return
	}
	
	userID := c.GetUint("userId")
	
	// Get all active cycles
	cycles, err := GetActiveMovieClubCycles(b.db)
	if err != nil {
		slog.Error("Failed to get active movie club cycles", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get cycles"})
		return
	}
	
	if len(cycles) == 0 {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "No active movie club cycles"})
		return
	}
	
	slog.Debug("Found active movie club cycles", "count", len(cycles))
	
	// Build response with user data for each cycle
	var responses []MovieClubCycleResponse
	for _, cycle := range cycles {
		// Get user nominations for this cycle
		userNominations, err := GetUserNominationsForCycle(b.db, cycle.ID, userID)
		if err != nil {
			slog.Error("Failed to get user nominations", "error", err, "cycleId", cycle.ID)
			continue
		}
		
		// Get user votes for this cycle
		userVotes, err := GetUserVotesForCycle(b.db, cycle.ID, userID)
		if err != nil {
			slog.Error("Failed to get user votes", "error", err, "cycleId", cycle.ID)
			continue
		}
		
		// Calculate abilities for this cycle
		canNominate := cycle.IsNominationPhase() && len(userNominations) < Config.MOVIE_CLUB.NominationsPerUser
		canVote := cycle.IsVotingPhase() && len(userVotes) < Config.MOVIE_CLUB.VotesPerUser
		
		// Get vote results if in watching phase
		var voteResults []MovieClubVoteCount
		if cycle.IsWatchingPhase() {
			voteResults, err = CalculateVoteResults(b.db, cycle.ID)
			if err != nil {
				slog.Error("Failed to calculate vote results", "error", err, "cycleId", cycle.ID)
			}
		}
		
		// Get cycle ratings if in watching phase
		var cycleRatings []MovieClubCycleRating
		if cycle.IsWatchingPhase() {
			cycleRatings, err = GetCycleRatingsForCycle(b.db, cycle.ID)
			if err != nil {
				slog.Error("Failed to get cycle ratings", "error", err, "cycleId", cycle.ID)
			}
		}
		
		response := MovieClubCycleResponse{
			Cycle:           cycle,
			UserNominations: userNominations,
			UserVotes:       userVotes,
			VoteResults:     voteResults,
			CycleRatings:    cycleRatings,
			CanNominate:     canNominate,
			CanVote:         canVote,
		}
		
		responses = append(responses, response)
	}
	
	c.JSON(http.StatusOK, responses)
}

// getArchivedMovieClubCycles returns all archived (completed) movie club cycles
func (b *BaseRouter) getArchivedMovieClubCycles(c *gin.Context) {
	// Check if movie club is enabled
	if !Config.MOVIE_CLUB.Enabled {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Movie club is not enabled"})
		return
	}
	
	userID := c.GetUint("userId")
	
	// Get all archived cycles
	cycles, err := GetArchivedMovieClubCycles(b.db)
	if err != nil {
		slog.Error("Failed to get archived movie club cycles", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get archived cycles"})
		return
	}
	
	slog.Debug("Found archived movie club cycles", "count", len(cycles))
	
	// Build response with user data for each cycle
	var responses []MovieClubCycleResponse
	for _, cycle := range cycles {
		// Get user nominations for this cycle
		userNominations, err := GetUserNominationsForCycle(b.db, cycle.ID, userID)
		if err != nil {
			slog.Error("Failed to get user nominations", "error", err, "cycleId", cycle.ID)
			userNominations = []MovieClubNomination{}
		}
		
		// Get user votes for this cycle
		userVotes, err := GetUserVotesForCycle(b.db, cycle.ID, userID)
		if err != nil {
			slog.Error("Failed to get user votes", "error", err, "cycleId", cycle.ID)
			userVotes = []MovieClubVote{}
		}
		
		// Get vote results for this cycle if it's in watching phase (completed)
		var voteResults []MovieClubVoteCount
		if cycle.Phase == PHASE_WATCHING {
			voteResults, err = CalculateVoteResults(b.db, cycle.ID)
			if err != nil {
				slog.Error("Failed to get vote results", "error", err, "cycleId", cycle.ID)
				voteResults = []MovieClubVoteCount{}
			}
		}
		
		// Get cycle ratings for archived cycles
		var cycleRatings []MovieClubCycleRating
		if cycle.Phase == PHASE_WATCHING {
			cycleRatings, err = GetCycleRatingsForCycle(b.db, cycle.ID)
			if err != nil {
				slog.Error("Failed to get cycle ratings", "error", err, "cycleId", cycle.ID)
				cycleRatings = []MovieClubCycleRating{}
			}
		}
		
		// Check if user can nominate/vote (always false for archived cycles)
		canNominate := false
		canVote := false
		
		// Build the response
		response := MovieClubCycleResponse{
			Cycle:           cycle,
			UserNominations: userNominations,
			UserVotes:       userVotes,
			VoteResults:     voteResults,
			CycleRatings:    cycleRatings,
			CanNominate:     canNominate,
			CanVote:         canVote,
		}
		
		responses = append(responses, response)
	}
	
	c.JSON(http.StatusOK, responses)
}

// nominateMovie allows a user to nominate a movie for the current cycle
func (b *BaseRouter) nominateMovie(c *gin.Context) {
	// Check if movie club is enabled
	if !Config.MOVIE_CLUB.Enabled {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "Movie club is not enabled"})
		return
	}
	
	userID := c.GetUint("userId")
	
	var req MovieClubNominationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request"})
		return
	}
	
	// Get the specified cycle or first active cycle
	var cycle *MovieClubCycle
	var err error
	
	if req.CycleID != nil {
		// Get specific cycle by ID
		var targetCycle MovieClubCycle
		result := b.db.Where("id = ? AND active = ?", *req.CycleID, true).First(&targetCycle)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, ErrorResponse{Error: "Specified cycle not found or inactive"})
			} else {
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get cycle"})
			}
			return
		}
		cycle = &targetCycle
	} else {
		// Fall back to first active cycle for backwards compatibility
		cycle, err = GetActiveMovieClubCycle(b.db)
		if err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "No active movie club cycle"})
			return
		}
	}
	
	// Check if we're in nomination phase
	if !cycle.IsNominationPhase() {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "Not currently in nomination phase"})
		return
	}
	
	// Check user nomination limit
	userNominations, err := GetUserNominationsForCycle(b.db, cycle.ID, userID)
	if err != nil {
		slog.Error("Failed to get user nominations", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to check nominations"})
		return
	}
	
	if len(userNominations) >= Config.MOVIE_CLUB.NominationsPerUser {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "Nomination limit reached"})
		return
	}
	
	// Check if content exists, if not create it
	content, err := getOrCacheContent(b.db, MOVIE, req.ContentID)
	if err != nil {
		slog.Error("Failed to get or create content", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to process content"})
		return
	}
	
	// Check if already nominated by this user
	var existingNomination MovieClubNomination
	result := b.db.Where("cycle_id = ? AND user_id = ? AND content_id = ?", cycle.ID, userID, content.ID).
		First(&existingNomination)
	
	if result.Error == nil {
		c.JSON(http.StatusConflict, ErrorResponse{Error: "Already nominated this movie"})
		return
	}
	
	// Create nomination
	nomination := MovieClubNomination{
		CycleID:   cycle.ID,
		UserID:    userID,
		ContentID: content.ID,
		Reason:    req.Reason,
	}
	
	if err := b.db.Create(&nomination).Error; err != nil {
		slog.Error("Failed to create nomination", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create nomination"})
		return
	}
	
	// Load relations for response
	b.db.Preload("Content").First(&nomination, nomination.ID)
	
	// Add activity
	activityData := fmt.Sprintf(`{"contentId": %d, "title": "%s", "reason": "%s"}`, 
		content.ID, content.Title, req.Reason)
	addMovieClubActivity(b.db, userID, MOVIE_CLUB_NOMINATED, activityData)
	
	c.JSON(http.StatusCreated, nomination)
}

// removeNomination allows a user to remove their nomination
func (b *BaseRouter) removeNomination(c *gin.Context) {
	userID := c.GetUint("userId")
	nominationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid nomination ID"})
		return
	}
	
	// First, get the nomination to determine which cycle it belongs to
	var nomination MovieClubNomination
	result := b.db.Where("id = ? AND user_id = ?", nominationID, userID).
		Preload("Cycle").
		First(&nomination)
	
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Nomination not found"})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get nomination"})
		}
		return
	}
	
	// Check if the cycle is still active
	if !nomination.Cycle.Active {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "Cannot remove nomination from inactive cycle"})
		return
	}
	
	// Check if the cycle is in nomination phase
	if !nomination.Cycle.IsNominationPhase() {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "Not currently in nomination phase for this cycle"})
		return
	}
	
	// Delete the nomination
	result = b.db.Delete(&nomination)
	
	if result.Error != nil {
		slog.Error("Failed to delete nomination", "error", result.Error)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete nomination"})
		return
	}
	
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Nomination not found"})
		return
	}
	
	// Add activity
	addMovieClubActivity(b.db, userID, MOVIE_CLUB_NOMINATION_REMOVED, fmt.Sprintf(`{"nominationId": %d}`, nominationID))
	
	c.JSON(http.StatusOK, gin.H{"message": "Nomination removed"})
}

// voteForMovies allows a user to cast their votes
func (b *BaseRouter) voteForMovies(c *gin.Context) {
	// Check if movie club is enabled
	if !Config.MOVIE_CLUB.Enabled {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "Movie club is not enabled"})
		return
	}
	
	userID := c.GetUint("userId")
	
	var req MovieClubVoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request"})
		return
	}
	
	// Get the specified cycle or first active cycle
	var cycle *MovieClubCycle
	var err error
	
	if req.CycleID != nil {
		// Get specific cycle by ID
		var targetCycle MovieClubCycle
		result := b.db.Where("id = ? AND active = ?", *req.CycleID, true).First(&targetCycle)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, ErrorResponse{Error: "Specified cycle not found or inactive"})
			} else {
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get cycle"})
			}
			return
		}
		cycle = &targetCycle
	} else {
		// Fall back to first active cycle for backwards compatibility
		cycle, err = GetActiveMovieClubCycle(b.db)
		if err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "No active movie club cycle"})
			return
		}
	}
	
	// Check if we're in voting phase
	if !cycle.IsVotingPhase() {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "Not currently in voting phase"})
		return
	}
	
	// Validate votes
	if len(req.Votes) > Config.MOVIE_CLUB.VotesPerUser {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Too many votes"})
		return
	}
	
	// Check for valid priorities and no duplicates
	priorityMap := make(map[int]bool)
	contentMap := make(map[int]bool)
	
	for _, vote := range req.Votes {
		if vote.Priority < 1 || vote.Priority > Config.MOVIE_CLUB.VotesPerUser {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid vote priority"})
			return
		}
		
		if priorityMap[vote.Priority] {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Duplicate vote priority"})
			return
		}
		priorityMap[vote.Priority] = true
		
		if contentMap[vote.ContentID] {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Cannot vote for same content multiple times"})
			return
		}
		contentMap[vote.ContentID] = true
	}
	
	// Clear existing votes for this user and cycle
	b.db.Where("cycle_id = ? AND user_id = ?", cycle.ID, userID).Delete(&MovieClubVote{})
	
	// Create new votes
	var votes []MovieClubVote
	for _, voteReq := range req.Votes {
		vote := MovieClubVote{
			CycleID:   cycle.ID,
			UserID:    userID,
			ContentID: voteReq.ContentID,
			Priority:  voteReq.Priority,
		}
		votes = append(votes, vote)
	}
	
	if err := b.db.Create(&votes).Error; err != nil {
		slog.Error("Failed to create votes", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create votes"})
		return
	}
	
	// Load relations for response
	var responseVotes []MovieClubVote
	b.db.Where("cycle_id = ? AND user_id = ?", cycle.ID, userID).
		Preload("Content").
		Order("priority ASC").
		Find(&responseVotes)
	
	// Add activity
	voteData := fmt.Sprintf(`{"voteCount": %d}`, len(votes))
	addMovieClubActivity(b.db, userID, MOVIE_CLUB_VOTED, voteData)
	
	c.JSON(http.StatusCreated, responseVotes)
}

// clearVotes allows a user to clear all their votes
func (b *BaseRouter) clearVotes(c *gin.Context) {
	userID := c.GetUint("userId")
	
	// Get the specified cycle or first active cycle
	var cycle *MovieClubCycle
	var err error
	
	// Check for cycle ID in query parameters
	cycleIDStr := c.Query("cycleId")
	if cycleIDStr != "" {
		// Parse cycle ID from query parameter
		cycleID, parseErr := strconv.ParseUint(cycleIDStr, 10, 32)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid cycle ID"})
			return
		}
		
		// Get specific cycle by ID
		var targetCycle MovieClubCycle
		result := b.db.Where("id = ? AND active = ?", uint(cycleID), true).First(&targetCycle)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, ErrorResponse{Error: "Specified cycle not found or inactive"})
			} else {
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get cycle"})
			}
			return
		}
		cycle = &targetCycle
	} else {
		// Fall back to first active cycle for backwards compatibility
		cycle, err = GetActiveMovieClubCycle(b.db)
		if err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "No active movie club cycle"})
			return
		}
	}
	
	// Check if we're in voting phase
	if !cycle.IsVotingPhase() {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "Not currently in voting phase"})
		return
	}
	
	// Delete all votes for this user and cycle
	result := b.db.Where("cycle_id = ? AND user_id = ?", cycle.ID, userID).Delete(&MovieClubVote{})
	
	if result.Error != nil {
		slog.Error("Failed to clear votes", "error", result.Error)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to clear votes"})
		return
	}
	
	// Add activity
	addMovieClubActivity(b.db, userID, MOVIE_CLUB_VOTES_CLEARED, `{}`)
	
	c.JSON(http.StatusOK, gin.H{"message": "Votes cleared"})
}

// getMovieClubResults returns the voting results for the current cycle
func (b *BaseRouter) getMovieClubResults(c *gin.Context) {
	// Get active cycle
	cycle, err := GetActiveMovieClubCycle(b.db)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "No active movie club cycle"})
		return
	}
	
	// Only show results in watching phase or if voting has ended
	if !cycle.IsWatchingPhase() && !cycle.ShouldTransitionPhase() {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "Results not yet available"})
		return
	}
	
	// Calculate results
	results, err := CalculateVoteResults(b.db, cycle.ID)
	if err != nil {
		slog.Error("Failed to calculate vote results", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to calculate results"})
		return
	}
	
	c.JSON(http.StatusOK, results)
}

// Admin endpoints

// createMovieClubCycle creates a new movie club cycle
func (b *BaseRouter) createMovieClubCycle(c *gin.Context) {
	var cycle MovieClubCycle
	if err := c.ShouldBindJSON(&cycle); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request"})
		return
	}
	
	// Use a transaction to ensure consistency
	tx := b.db.Begin()
	if tx.Error != nil {
		slog.Error("Failed to begin transaction", "error", tx.Error)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create cycle"})
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	// Note: We no longer deactivate existing cycles when creating new ones
	// Multiple cycles can be active simultaneously
	
	// Set up new cycle
	now := time.Now()
	phaseDuration := time.Duration(Config.MOVIE_CLUB.PhaseDurationDays) * 24 * time.Hour
	
	cycle.Phase = PHASE_NOMINATION
	cycle.PhaseStartDate = now
	cycle.PhaseEndDate = now.Add(phaseDuration)
	cycle.NominationEndDate = now.Add(phaseDuration)
	cycle.VotingEndDate = now.Add(2 * phaseDuration)
	cycle.WatchingEndDate = now.Add(3 * phaseDuration)
	cycle.Active = true
	
	if err := tx.Create(&cycle).Error; err != nil {
		tx.Rollback()
		slog.Error("Failed to create movie club cycle", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create cycle"})
		return
	}
	
	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		slog.Error("Failed to commit transaction", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create cycle"})
		return
	}
	
	slog.Info("Successfully created movie club cycle", "cycleId", cycle.ID, "active", cycle.Active, "phase", cycle.Phase)
	
	// Verify the cycle was created and can be retrieved
	var verifyResult MovieClubCycle
	if err := b.db.Where("active = ?", true).First(&verifyResult).Error; err != nil {
		slog.Error("Verification failed: newly created cycle not found", "error", err, "cycleId", cycle.ID)
	} else {
		slog.Debug("Verification successful: cycle found", "cycleId", verifyResult.ID, "active", verifyResult.Active)
	}
	
	c.JSON(http.StatusCreated, cycle)
}

// updateMovieClubCycle updates an existing cycle
func (b *BaseRouter) updateMovieClubCycle(c *gin.Context) {
	cycleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid cycle ID"})
		return
	}
	
	var updateData MovieClubCycle
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request"})
		return
	}
	
	var cycle MovieClubCycle
	if err := b.db.First(&cycle, cycleID).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Cycle not found"})
		return
	}
	
	// Update allowed fields
	cycle.Name = updateData.Name
	cycle.Description = updateData.Description
	
	if err := b.db.Save(&cycle).Error; err != nil {
		slog.Error("Failed to update movie club cycle", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update cycle"})
		return
	}
	
	c.JSON(http.StatusOK, cycle)
}

// deleteMovieClubCycle deletes a cycle
func (b *BaseRouter) deleteMovieClubCycle(c *gin.Context) {
	cycleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid cycle ID"})
		return
	}
	
	// Use GORM soft delete to populate deleted_at and set active to false
	// This preserves the cycle data and history while marking it as deleted
	slog.Info("Soft deleting movie club cycle", "cycleId", cycleID)
	
	// Use a transaction to ensure both operations succeed
	tx := b.db.Begin()
	if tx.Error != nil {
		slog.Error("Failed to begin transaction", "error", tx.Error)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete cycle"})
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	// First set active to false
	if err := tx.Model(&MovieClubCycle{}).Where("id = ?", cycleID).Update("active", false).Error; err != nil {
		tx.Rollback()
		slog.Error("Failed to deactivate cycle", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to deactivate cycle"})
		return
	}
	
	// Then perform soft delete to populate deleted_at
	if err := tx.Delete(&MovieClubCycle{}, cycleID).Error; err != nil {
		tx.Rollback()
		slog.Error("Failed to soft delete cycle", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete cycle"})
		return
	}
	
	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		slog.Error("Failed to commit transaction", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete cycle"})
		return
	}
	
	slog.Info("Successfully soft deleted movie club cycle", "cycleId", cycleID)
	c.JSON(http.StatusOK, gin.H{"message": "Cycle deleted"})
}

// transitionCyclePhase manually transitions the current cycle to the next phase
func (b *BaseRouter) transitionCyclePhase(c *gin.Context) {
	cycleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid cycle ID"})
		return
	}
	
	var cycle MovieClubCycle
	if err := b.db.First(&cycle, cycleID).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Cycle not found"})
		return
	}
	
	if err := TransitionCyclePhase(b.db, &cycle); err != nil {
		slog.Error("Failed to transition cycle phase", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to transition phase"})
		return
	}
	
	c.JSON(http.StatusOK, cycle)
}

// getAllMovieClubCycles returns all cycles for admin management
func (b *BaseRouter) getAllMovieClubCycles(c *gin.Context) {
	var cycles []MovieClubCycle
	result := b.db.Preload("WinnerContent").
		Order("created_at DESC").
		Find(&cycles)
	
	if result.Error != nil {
		slog.Error("Failed to get movie club cycles", "error", result.Error)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get cycles"})
		return
	}
	
	c.JSON(http.StatusOK, cycles)
}

// Helper function to transition cycle phases
func TransitionCyclePhase(db *gorm.DB, cycle *MovieClubCycle) error {
	now := time.Now()
	phaseDuration := time.Duration(Config.MOVIE_CLUB.PhaseDurationDays) * 24 * time.Hour
	
	nextPhase := cycle.GetNextPhase()
	
	// If transitioning from voting to watching, calculate winner
	if cycle.Phase == PHASE_VOTING && nextPhase == PHASE_WATCHING {
		results, err := CalculateVoteResults(db, cycle.ID)
		if err != nil {
			return err
		}
		
		// Set winner (first in results is the winner)
		if len(results) > 0 {
			cycle.WinnerContentID = &results[0].ContentID
		}
	}
	
	// Update cycle
	cycle.Phase = nextPhase
	cycle.PhaseStartDate = now
	
	// Use specific phase end date if available, otherwise calculate from duration
	switch nextPhase {
	case PHASE_NOMINATION:
		if !cycle.NominationEndDate.IsZero() && cycle.NominationEndDate.After(now) {
			cycle.PhaseEndDate = cycle.NominationEndDate
		} else {
			cycle.PhaseEndDate = now.Add(phaseDuration)
		}
	case PHASE_VOTING:
		if !cycle.VotingEndDate.IsZero() && cycle.VotingEndDate.After(now) {
			cycle.PhaseEndDate = cycle.VotingEndDate
		} else {
			cycle.PhaseEndDate = now.Add(phaseDuration)
		}
	case PHASE_WATCHING:
		if !cycle.WatchingEndDate.IsZero() && cycle.WatchingEndDate.After(now) {
			cycle.PhaseEndDate = cycle.WatchingEndDate
		} else {
			cycle.PhaseEndDate = now.Add(phaseDuration)
		}
	default:
		cycle.PhaseEndDate = now.Add(phaseDuration)
	}
	
	return db.Save(cycle).Error
}

// Initialize default movie club settings
func InitializeMovieClubSettings() MovieClubSettings {
	return MovieClubSettings{
		NominationsPerUser: 1,
		VotesPerUser:       2,
		PhaseDurationDays:  7,
		Enabled:            false,
	}
}

// checkMovieClubTransition is called by the task scheduler to check if cycles need phase transitions
func checkMovieClubTransition(db *gorm.DB) {
	// Skip if movie club is not enabled
	if !Config.MOVIE_CLUB.Enabled {
		return
	}
	
	slog.Debug("Checking movie club cycle transitions")
	
	// Get all active cycles
	cycles, err := GetActiveMovieClubCycles(db)
	if err != nil {
		slog.Error("Failed to get active movie club cycles for transition check", "error", err)
		return
	}
	
	// Process each cycle for transitions
	for _, cycle := range cycles {
		// Check if transition is needed
		if !cycle.ShouldTransitionPhase() {
			continue
		}
		
		slog.Info("Transitioning movie club cycle phase", "cycleId", cycle.ID, "currentPhase", cycle.Phase)
		
		// Handle special case for watching phase ending - deactivate cycle
		if cycle.Phase == PHASE_WATCHING {
			// End current cycle - set to inactive
			cycle.Active = false
			if err := db.Save(&cycle).Error; err != nil {
				slog.Error("Failed to deactivate completed cycle", "error", err, "cycleId", cycle.ID)
				continue
			}
			
			slog.Info("Deactivated completed watching phase cycle", "cycleId", cycle.ID)
		} else {
			// Regular phase transition (nomination -> voting, voting -> watching)
			if err := TransitionCyclePhase(db, &cycle); err != nil {
				slog.Error("Failed to transition cycle phase", "error", err, "cycleId", cycle.ID)
				continue
			}
			
			slog.Info("Successfully transitioned cycle phase", "cycleId", cycle.ID, "newPhase", cycle.Phase)
		}
	}
}

// Cycle Rating Helper Functions

// IsUserEligibleForCycleRating checks if a user is eligible to have their rating saved for a cycle
// A user is eligible if they nominated any movie OR voted in the cycle
func IsUserEligibleForCycleRating(db *gorm.DB, userID uint, cycleID uint) bool {
	// Check nomination participation using EXISTS for performance
	var nominationCount int64
	db.Model(&MovieClubNomination{}).
		Where("cycle_id = ? AND user_id = ?", cycleID, userID).
		Count(&nominationCount)
	
	if nominationCount > 0 {
		return true
	}
	
	// Check voting participation using EXISTS for performance
	var voteCount int64
	db.Model(&MovieClubVote{}).
		Where("cycle_id = ? AND user_id = ?", cycleID, userID).
		Count(&voteCount)
	
	return voteCount > 0
}

// GetActiveWatchingCyclesByWinnerContent returns active watching cycles where the given content won
func GetActiveWatchingCyclesByWinnerContent(db *gorm.DB, contentID int) ([]MovieClubCycle, error) {
	var cycles []MovieClubCycle
	result := db.Where("active = ? AND phase = ? AND winner_content_id = ?", 
		true, PHASE_WATCHING, contentID).Find(&cycles)
	
	if result.Error != nil {
		return nil, result.Error
	}
	
	return cycles, nil
}

// CreateOrUpdateCycleRating creates or updates a cycle rating for a user
func CreateOrUpdateCycleRating(db *gorm.DB, userID uint, cycleID uint, contentID int, rating float64, thoughts string) error {
	// Check if rating already exists
	var existingRating MovieClubCycleRating
	result := db.Where("cycle_id = ? AND user_id = ? AND content_id = ?", 
		cycleID, userID, contentID).First(&existingRating)
	
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}
	
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// Create new rating
		newRating := MovieClubCycleRating{
			CycleID:   cycleID,
			UserID:    userID,
			ContentID: contentID,
			Rating:    rating,
			Thoughts:  thoughts,
		}
		
		return db.Create(&newRating).Error
	} else {
		// Update existing rating
		existingRating.Rating = rating
		existingRating.Thoughts = thoughts
		return db.Save(&existingRating).Error
	}
}

// GetCycleRatingsForCycle returns all cycle ratings for a specific cycle
func GetCycleRatingsForCycle(db *gorm.DB, cycleID uint) ([]MovieClubCycleRating, error) {
	var ratings []MovieClubCycleRating
	result := db.Where("cycle_id = ?", cycleID).
		Preload("User").
		Preload("Content").
		Order("created_at ASC").
		Find(&ratings)
	
	if result.Error != nil {
		return nil, result.Error
	}
	
	return ratings, nil
}

// ProcessPotentialCycleRating processes a watched entry to see if it should be saved as a cycle rating
// This is the main entry point for the automatic rating capture system
func ProcessPotentialCycleRating(db *gorm.DB, userID uint, contentID int, rating float64, thoughts string) error {
	// Performance optimization: Stage 1 - Quick check for winning content in active watching cycles
	winningCycles, err := GetActiveWatchingCyclesByWinnerContent(db, contentID)
	if err != nil {
		return err
	}
	
	// Early exit if content is not a winner in any active watching cycle
	if len(winningCycles) == 0 {
		return nil
	}
	
	// Performance optimization: Stage 2 - Check user eligibility for each cycle
	for _, cycle := range winningCycles {
		if IsUserEligibleForCycleRating(db, userID, cycle.ID) {
			// Performance optimization: Stage 3 - Process the rating (lowest frequency operation)
			if err := CreateOrUpdateCycleRating(db, userID, cycle.ID, contentID, rating, thoughts); err != nil {
				slog.Error("Failed to save cycle rating", "error", err, "userID", userID, "cycleID", cycle.ID, "contentID", contentID)
				continue // Continue processing other cycles even if one fails
			}
			
			slog.Debug("Saved cycle rating", "userID", userID, "cycleID", cycle.ID, "contentID", contentID, "rating", rating)
		}
	}
	
	return nil
}