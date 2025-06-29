package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
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
	Nominations       []MovieClubNomination `json:"nominations,omitempty" gorm:"foreignKey:CycleID"`
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
	ContentID    int     `json:"contentId"`
	Content      Content `json:"content"`
	TotalVotes   int     `json:"totalVotes"`
	FirstChoice  int     `json:"firstChoice"`
	SecondChoice int     `json:"secondChoice"`
	ThirdChoice  int     `json:"thirdChoice"`
	WeightedScore float64 `json:"weightedScore"` // Calculated score based on vote priorities
}

// MovieClubNominationRequest represents the request to nominate a movie
type MovieClubNominationRequest struct {
	ContentID int    `json:"contentId" binding:"required"`
	Reason    string `json:"reason" binding:"max=500"`
}

// MovieClubVoteRequest represents the request to cast votes
type MovieClubVoteRequest struct {
	Votes []MovieClubVoteItem `json:"votes" binding:"required,dive"`
}

type MovieClubVoteItem struct {
	ContentID int `json:"contentId" binding:"required"`
	Priority  int `json:"priority" binding:"required,min=1"`
}

// MovieClubCycleResponse represents the full cycle data for frontend
type MovieClubCycleResponse struct {
	Cycle          MovieClubCycle       `json:"cycle"`
	UserNominations []MovieClubNomination `json:"userNominations"`
	UserVotes      []MovieClubVote      `json:"userVotes"`
	VoteResults    []MovieClubVoteCount `json:"voteResults,omitempty"`
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

// GetActiveMovieClubCycle returns the currently active movie club cycle
func GetActiveMovieClubCycle(db *gorm.DB) (*MovieClubCycle, error) {
	slog.Debug("GetActiveMovieClubCycle: Looking for active cycle")
	
	var cycle MovieClubCycle
	result := db.Where("active = ?", true).
		Preload("WinnerContent").
		Preload("Nominations.Content").
		Preload("Nominations.User").
		First(&cycle)
	
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			slog.Debug("GetActiveMovieClubCycle: No active cycle found")
		} else {
			slog.Error("GetActiveMovieClubCycle: Database error", "error", result.Error)
		}
		return nil, result.Error
	}
	
	slog.Debug("GetActiveMovieClubCycle: Found active cycle", "cycleId", cycle.ID, "phase", cycle.Phase, "active", cycle.Active)
	return &cycle, nil
}

// GetMovieClubNominationsForCycle returns all nominations for a specific cycle
func GetMovieClubNominationsForCycle(db *gorm.DB, cycleID uint) ([]MovieClubNomination, error) {
	var nominations []MovieClubNomination
	result := db.Where("cycle_id = ?", cycleID).
		Preload("User").
		Preload("Content").
		Find(&nominations)
	
	if result.Error != nil {
		return nil, result.Error
	}
	
	return nominations, nil
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
	
	// Sort by weighted score (descending), then by first choice votes, then by total votes
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].WeightedScore < results[j].WeightedScore ||
				(results[i].WeightedScore == results[j].WeightedScore && results[i].FirstChoice < results[j].FirstChoice) ||
				(results[i].WeightedScore == results[j].WeightedScore && results[i].FirstChoice == results[j].FirstChoice && results[i].TotalVotes < results[j].TotalVotes) {
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
	
	response := MovieClubCycleResponse{
		Cycle:           *cycle,
		UserNominations: userNominations,
		UserVotes:       userVotes,
		VoteResults:     voteResults,
		CanNominate:     canNominate,
		CanVote:         canVote,
	}
	
	c.JSON(http.StatusOK, response)
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
	
	// Get active cycle
	cycle, err := GetActiveMovieClubCycle(b.db)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "No active movie club cycle"})
		return
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
	
	// Get active cycle
	cycle, err := GetActiveMovieClubCycle(b.db)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "No active movie club cycle"})
		return
	}
	
	// Check if we're in nomination phase
	if !cycle.IsNominationPhase() {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "Not currently in nomination phase"})
		return
	}
	
	// Find and delete the nomination (only if it belongs to the user)
	result := b.db.Where("id = ? AND user_id = ? AND cycle_id = ?", nominationID, userID, cycle.ID).
		Delete(&MovieClubNomination{})
	
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
	
	// Get active cycle
	cycle, err := GetActiveMovieClubCycle(b.db)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "No active movie club cycle"})
		return
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
	
	// Get active cycle
	cycle, err := GetActiveMovieClubCycle(b.db)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "No active movie club cycle"})
		return
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
	
	// Deactivate any existing active cycle
	if err := tx.Model(&MovieClubCycle{}).Where("active = ?", true).Update("active", false).Error; err != nil {
		tx.Rollback()
		slog.Error("Failed to deactivate existing cycles", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create cycle"})
		return
	}
	
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
	
	// Delete cycle and all related data
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
	
	slog.Info("Deleting movie club cycle", "cycleId", cycleID)
	
	if err := tx.Where("cycle_id = ?", cycleID).Delete(&MovieClubVote{}).Error; err != nil {
		tx.Rollback()
		slog.Error("Failed to delete votes", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete votes"})
		return
	}
	
	if err := tx.Where("cycle_id = ?", cycleID).Delete(&MovieClubNomination{}).Error; err != nil {
		tx.Rollback()
		slog.Error("Failed to delete nominations", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete nominations"})
		return
	}
	
	if err := tx.Delete(&MovieClubCycle{}, cycleID).Error; err != nil {
		tx.Rollback()
		slog.Error("Failed to delete cycle", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete cycle"})
		return
	}
	
	if err := tx.Commit().Error; err != nil {
		slog.Error("Failed to commit transaction", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete cycle"})
		return
	}
	
	slog.Info("Successfully deleted movie club cycle", "cycleId", cycleID)
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
	cycle.PhaseEndDate = now.Add(phaseDuration)
	
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
	
	// Get active cycle
	cycle, err := GetActiveMovieClubCycle(db)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("Failed to get active movie club cycle for transition check", "error", err)
		}
		return
	}
	
	// Check if transition is needed
	if !cycle.ShouldTransitionPhase() {
		return
	}
	
	slog.Info("Transitioning movie club cycle phase", "cycleId", cycle.ID, "currentPhase", cycle.Phase)
	
	// Handle special case for watching -> nomination transition (start new cycle)
	if cycle.Phase == PHASE_WATCHING {
		// End current cycle
		cycle.Active = false
		if err := db.Save(cycle).Error; err != nil {
			slog.Error("Failed to deactivate completed cycle", "error", err)
			return
		}
		
		// Create new cycle
		now := time.Now()
		phaseDuration := time.Duration(Config.MOVIE_CLUB.PhaseDurationDays) * 24 * time.Hour
		
		newCycle := MovieClubCycle{
			Name:              "Movie Club Cycle",
			Description:       "Automated movie club cycle",
			Phase:             PHASE_NOMINATION,
			PhaseStartDate:    now,
			PhaseEndDate:      now.Add(phaseDuration),
			NominationEndDate: now.Add(phaseDuration),
			VotingEndDate:     now.Add(2 * phaseDuration),
			WatchingEndDate:   now.Add(3 * phaseDuration),
			Active:            true,
		}
		
		if err := db.Create(&newCycle).Error; err != nil {
			slog.Error("Failed to create new movie club cycle", "error", err)
			return
		}
		
		slog.Info("Created new movie club cycle", "cycleId", newCycle.ID)
	} else {
		// Regular phase transition
		if err := TransitionCyclePhase(db, cycle); err != nil {
			slog.Error("Failed to transition cycle phase", "error", err)
			return
		}
		
		slog.Info("Successfully transitioned cycle phase", "cycleId", cycle.ID, "newPhase", cycle.Phase)
	}
}