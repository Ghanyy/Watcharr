package main

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
	"gorm.io/gorm"
)

// Room creation and management functions

// CreateCycleRoom creates a Matrix room for a movie club cycle
func CreateCycleRoom(db *gorm.DB, cycle *MovieClubCycle) (*MatrixRoom, error) {
	if matrixClient == nil {
		return nil, errors.New("matrix client not initialized")
	}

	if !Config.MOVIE_CLUB.Matrix.Enabled {
		return nil, errors.New("matrix integration is disabled")
	}

	// Check if room already exists for this cycle
	var existingRoom MatrixRoom
	if err := db.Where("cycle_id = ?", cycle.ID).First(&existingRoom).Error; err == nil {
		slog.Debug("Room already exists for cycle", "cycle_id", cycle.ID, "room_id", existingRoom.RoomID)
		return &existingRoom, nil
	}

	// Generate room alias and name
	alias := generateRoomAlias(cycle)
	roomName := fmt.Sprintf("Movie Club: %s (%s)",
		cycle.WinnerContent.Title,
		cycle.VotingEndDate.Format("2006-01-02"))

	// Create room configuration
	createReq := &mautrix.ReqCreateRoom{
		RoomAliasName: alias,
		Name:          roomName,
		Topic:         fmt.Sprintf("Discussion for %s - Movie Club Cycle", cycle.WinnerContent.Title),
		Preset:        "private_chat",
		Visibility:    "private",
		IsDirect:      false,
		Federate:      false, // Local-only as specified
		PowerLevelContentOverride: &event.PowerLevelsEventContent{
			UsersDefault: 0, // Regular users
			Users: map[id.UserID]int{
				id.UserID(Config.MOVIE_CLUB.Matrix.AdminUserID): 100, // Watcharr admin
			},
		},
		InitialState: []event.Event{
			{
				Type: event.StateTopic,
				Content: event.Content{
					Parsed: &event.TopicEventContent{
						Topic: fmt.Sprintf("Movie Club discussion for %s", cycle.WinnerContent.Title),
					},
				},
			},
		},
	}

	// Create the room
	resp, err := matrixClient.CreateRoom(createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create Matrix room: %w", err)
	}

	// Store room in database
	matrixRoom := &MatrixRoom{
		CycleID:   cycle.ID,
		RoomID:    resp.RoomID.String(),
		RoomAlias: fmt.Sprintf("#%s:%s", alias, Config.MOVIE_CLUB.Matrix.ServerName),
	}

	if err := db.Create(matrixRoom).Error; err != nil {
		return nil, fmt.Errorf("failed to store Matrix room: %w", err)
	}

	slog.Info("Created Matrix room for cycle",
		"cycle_id", cycle.ID,
		"room_id", resp.RoomID,
		"alias", matrixRoom.RoomAlias,
		"movie", cycle.WinnerContent.Title)

	// Add room to Movie Club space
	if err := AddRoomToMovieClubSpace(db, resp.RoomID.String()); err != nil {
		slog.Warn("Failed to add room to Movie Club space", "error", err, "room_id", resp.RoomID)
	}

	// Invite eligible users to the room
	if err := InviteUsersToRoom(db, cycle.ID); err != nil {
		slog.Warn("Failed to invite users to room", "error", err, "room_id", resp.RoomID)
	}

	return matrixRoom, nil
}

// InviteUsersToRoom invites all eligible users to a cycle room
func InviteUsersToRoom(db *gorm.DB, cycleID uint) error {
	if matrixClient == nil {
		return errors.New("matrix client not initialized")
	}

	// Get the room
	var room MatrixRoom
	if err := db.Where("cycle_id = ?", cycleID).First(&room).Error; err != nil {
		return fmt.Errorf("room not found for cycle %d: %w", cycleID, err)
	}

	// Get eligible users (those who nominated or voted in this cycle)
	eligibleUsers, err := GetEligibleUsersForCycle(db, cycleID)
	if err != nil {
		return fmt.Errorf("failed to get eligible users: %w", err)
	}

	roomID := id.RoomID(room.RoomID)
	successCount := 0
	errorCount := 0

	for _, user := range eligibleUsers {
		// Get Matrix user for this Watcharr user
		matrixUser, err := GetOrCreateMatrixUser(db, user.ID, user.Username)
		if err != nil {
			slog.Warn("Failed to get Matrix user for invitation",
				"watcharr_user_id", user.ID,
				"username", user.Username,
				"error", err)
			errorCount++
			continue
		}

		// Invite user to room
		_, err = matrixClient.InviteUser(roomID, &mautrix.ReqInviteUser{
			UserID: id.UserID(matrixUser.MatrixUserID),
		})

		if err != nil {
			slog.Warn("Failed to invite user to room",
				"matrix_user_id", matrixUser.MatrixUserID,
				"room_id", roomID,
				"error", err)
			errorCount++
			continue
		}

		// Record membership in database
		membership := &MatrixRoomMember{
			RoomID:   room.ID,
			UserID:   user.ID,
			JoinedAt: time.Now(),
		}

		if err := db.Create(membership).Error; err != nil {
			slog.Warn("Failed to record room membership", "error", err)
		}

		successCount++
		slog.Debug("Invited user to room",
			"matrix_user_id", matrixUser.MatrixUserID,
			"room_id", roomID)
	}

	slog.Info("Room invitation complete",
		"cycle_id", cycleID,
		"room_id", roomID,
		"invited", successCount,
		"errors", errorCount)

	return nil
}

// GetEligibleUsersForCycle returns users who participated in a cycle (nominated or voted)
func GetEligibleUsersForCycle(db *gorm.DB, cycleID uint) ([]User, error) {
	var users []User

	// Get users who nominated in this cycle
	err := db.Raw(`
		SELECT DISTINCT u.* FROM users u
		INNER JOIN movie_club_nominations mcn ON u.id = mcn.user_id
		WHERE mcn.cycle_id = ?
		UNION
		SELECT DISTINCT u.* FROM users u
		INNER JOIN movie_club_votes mcv ON u.id = mcv.user_id
		WHERE mcv.cycle_id = ?
	`, cycleID, cycleID).Scan(&users).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get eligible users: %w", err)
	}

	return users, nil
}

// Movie Club Space Management

// EnsureMovieClubSpace creates or gets the Movie Club space
func EnsureMovieClubSpace(db *gorm.DB) (*MatrixSpace, error) {
	if matrixClient == nil {
		return nil, errors.New("matrix client not initialized")
	}

	// Check if space already exists
	var existingSpace MatrixSpace
	if err := db.First(&existingSpace).Error; err == nil {
		return &existingSpace, nil
	}

	spaceName := Config.MOVIE_CLUB.Matrix.SpaceName
	if spaceName == "" {
		spaceName = "Movie Club"
	}

	// Create space
	createReq := &mautrix.ReqCreateRoom{
		Name:       spaceName,
		Topic:      "Movie Club discussions and activities",
		Preset:     "private_chat",
		Visibility: "private",
		Federate:   false,
		CreationContent: map[string]interface{}{
			"type": "m.space",
		},
		PowerLevelContentOverride: &event.PowerLevelsEventContent{
			UsersDefault: 0,
			Users: map[id.UserID]int{
				id.UserID(Config.MOVIE_CLUB.Matrix.AdminUserID): 100,
			},
		},
	}

	resp, err := matrixClient.CreateRoom(createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create Movie Club space: %w", err)
	}

	// Store space in database
	space := &MatrixSpace{
		SpaceID:   resp.RoomID.String(),
		SpaceName: spaceName,
	}

	if err := db.Create(space).Error; err != nil {
		return nil, fmt.Errorf("failed to store Movie Club space: %w", err)
	}

	slog.Info("Created Movie Club space", "space_id", resp.RoomID, "name", spaceName)
	return space, nil
}

// AddRoomToMovieClubSpace adds a room to the Movie Club space
func AddRoomToMovieClubSpace(db *gorm.DB, roomID string) error {
	if matrixClient == nil {
		return errors.New("matrix client not initialized")
	}

	// Ensure Movie Club space exists
	space, err := EnsureMovieClubSpace(db)
	if err != nil {
		return fmt.Errorf("failed to ensure Movie Club space: %w", err)
	}

	spaceID := id.RoomID(space.SpaceID)
	childRoomID := id.RoomID(roomID)

	// Add room as child of space
	stateKey := roomID
	_, err = matrixClient.SendStateEvent(spaceID, event.StateSpaceChild, stateKey, map[string]interface{}{
		"via": []string{Config.MOVIE_CLUB.Matrix.ServerName},
	})

	if err != nil {
		return fmt.Errorf("failed to add room to space: %w", err)
	}

	// Add space as parent of room
	_, err = matrixClient.SendStateEvent(childRoomID, event.StateSpaceParent, space.SpaceID, map[string]interface{}{
		"via": []string{Config.MOVIE_CLUB.Matrix.ServerName},
	})

	if err != nil {
		slog.Warn("Failed to set space as parent of room", "error", err)
		// Don't return error as the main relationship is established
	}

	slog.Debug("Added room to Movie Club space", "room_id", roomID, "space_id", space.SpaceID)
	return nil
}

// Room lifecycle management

// ArchiveOldRooms archives rooms for completed cycles (future enhancement)
func ArchiveOldRooms() error {
	// Implementation for archiving old rooms
	// For now, we keep all rooms accessible as specified
	return nil
}

// GetUserRooms returns Matrix rooms that a user has access to
func GetUserRooms(db *gorm.DB, userID uint) ([]MatrixRoom, error) {
	var rooms []MatrixRoom

	err := db.Raw(`
		SELECT mr.* FROM matrix_rooms mr
		INNER JOIN matrix_room_members mrm ON mr.id = mrm.room_id
		WHERE mrm.user_id = ?
		ORDER BY mr.created_at DESC
	`, userID).Scan(&rooms).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get user rooms: %w", err)
	}

	// Preload cycle and content information
	for i := range rooms {
		db.Preload("Cycle.WinnerContent").Find(&rooms[i], rooms[i].ID)
	}

	return rooms, nil
}

// CleanupMatrixResources cleans up Matrix resources when features are disabled
func CleanupMatrixResources() error {
	// Implementation for cleanup when Matrix is disabled
	// For now, we don't automatically remove rooms to preserve chat history
	slog.Info("Matrix cleanup requested - preserving existing rooms and data")
	return nil
}

// Utility functions

// generateRoomAlias creates a Matrix room alias for a cycle
func generateRoomAlias(cycle *MovieClubCycle) string {
	// Format: movieclub-YYYY-MM-DD-normalized-title
	dateStr := cycle.VotingEndDate.Format("2006-01-02")
	normalizedTitle := normalizeTitle(cycle.WinnerContent.Title)
	return fmt.Sprintf("movieclub-%s-%s", dateStr, normalizedTitle)
}

// normalizeTitle normalizes a movie title for use in Matrix room alias
func normalizeTitle(title string) string {
	// Convert to lowercase and replace non-alphanumeric characters with hyphens
	normalized := strings.ToLower(title)
	
	// Replace multiple spaces and special characters with single hyphen
	var result strings.Builder
	lastWasHyphen := false
	
	for _, r := range normalized {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			result.WriteRune(r)
			lastWasHyphen = false
		} else if !lastWasHyphen {
			result.WriteRune('-')
			lastWasHyphen = true
		}
	}
	
	// Remove trailing hyphen if present
	finalResult := strings.TrimSuffix(result.String(), "-")
	
	// Limit length to 50 characters to avoid Matrix alias length limits
	if len(finalResult) > 50 {
		finalResult = finalResult[:50]
		// Remove trailing hyphen that might have been created by truncation
		finalResult = strings.TrimSuffix(finalResult, "-")
	}
	
	return finalResult
}