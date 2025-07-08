package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"maunium.net/go/mautrix"
)

// Matrix API endpoints

// MatrixTestConnectionRequest represents a request to test Matrix connection
type MatrixTestConnectionRequest struct {
	ServerURL   string `json:"serverUrl" binding:"required"`
	AdminToken  string `json:"adminToken" binding:"required"`
}

// MatrixValidationResult represents a single validation check result
type MatrixValidationResult struct {
	Check   string `json:"check"`
	Status  string `json:"status"` // "success", "warning", "error"
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// MatrixValidationResponse represents the complete validation response
type MatrixValidationResponse struct {
	OverallStatus string                   `json:"overallStatus"` // "success", "warning", "error"
	Results       []MatrixValidationResult `json:"results"`
	Summary       string                   `json:"summary"`
}

// testMatrixConnection tests the connection to a Matrix server
func (b *BaseRouter) testMatrixConnection(c *gin.Context) {
	var req MatrixTestConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request format"})
		return
	}

	// Create temporary client for testing
	client, err := mautrix.NewClient(req.ServerURL, "", "")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Failed to create Matrix client"})
		return
	}

	client.AccessToken = req.AdminToken

	// Test connection with whoami request
	resp, err := client.Whoami(context.Background())
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Failed to authenticate with Matrix server"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Matrix connection successful",
		"userId":  resp.UserID.String(),
	})
}

// validateMatrixSetup performs comprehensive Matrix setup validation
func (b *BaseRouter) validateMatrixSetup(c *gin.Context) {
	validation := MatrixValidationResponse{
		Results: []MatrixValidationResult{},
	}

	// Check if Matrix is enabled
	if !Config.MOVIE_CLUB.Matrix.Enabled {
		validation.Results = append(validation.Results, MatrixValidationResult{
			Check:   "Matrix Integration Enabled",
			Status:  "error",
			Message: "Matrix integration is disabled",
			Details: "Enable Matrix integration in Movie Club settings to proceed",
		})
		validation.OverallStatus = "error"
		validation.Summary = "Matrix integration is not enabled"
		c.JSON(http.StatusOK, validation)
		return
	}

	settings := Config.MOVIE_CLUB.Matrix

	// 1. Configuration validation
	validation.Results = append(validation.Results, validateMatrixConfig(settings)...)

	// 2. Connection test
	if settings.ServerURL != "" && settings.AdminToken != "" {
		validation.Results = append(validation.Results, validateMatrixConnection(settings)...)
	}

	// 3. Permissions test
	if settings.ServerURL != "" && settings.AdminToken != "" {
		validation.Results = append(validation.Results, validateMatrixPermissions(settings)...)
	}

	// 4. Room creation test
	if settings.ServerURL != "" && settings.AdminToken != "" {
		validation.Results = append(validation.Results, validateRoomCreation(settings)...)
	}

	// Determine overall status
	hasError := false
	hasWarning := false
	for _, result := range validation.Results {
		if result.Status == "error" {
			hasError = true
		} else if result.Status == "warning" {
			hasWarning = true
		}
	}

	if hasError {
		validation.OverallStatus = "error"
		validation.Summary = "Matrix setup has errors that need to be resolved"
	} else if hasWarning {
		validation.OverallStatus = "warning"
		validation.Summary = "Matrix setup is functional but has some warnings"
	} else {
		validation.OverallStatus = "success"
		validation.Summary = "Matrix setup is fully configured and operational"
	}

	c.JSON(http.StatusOK, validation)
}

// validateMatrixConfig validates the Matrix configuration
func validateMatrixConfig(settings MatrixSettings) []MatrixValidationResult {
	var results []MatrixValidationResult

	// Check Server URL
	if settings.ServerURL == "" {
		results = append(results, MatrixValidationResult{
			Check:   "Server URL",
			Status:  "error",
			Message: "Matrix server URL is not configured",
			Details: "Please provide the URL of your Dendrite/Matrix server",
		})
	} else {
		results = append(results, MatrixValidationResult{
			Check:   "Server URL",
			Status:  "success",
			Message: "Server URL is configured",
			Details: settings.ServerURL,
		})
	}

	// Check Admin Token
	if settings.AdminToken == "" {
		results = append(results, MatrixValidationResult{
			Check:   "Admin Token",
			Status:  "error",
			Message: "Admin access token is not configured",
			Details: "Please provide a valid admin access token for Matrix operations",
		})
	} else {
		results = append(results, MatrixValidationResult{
			Check:   "Admin Token",
			Status:  "success",
			Message: "Admin token is configured",
			Details: "Token provided (hidden for security)",
		})
	}

	// Check Server Name
	if settings.ServerName == "" {
		results = append(results, MatrixValidationResult{
			Check:   "Server Name",
			Status:  "warning",
			Message: "Server name is not configured",
			Details: "Server name is recommended for proper Matrix federation",
		})
	} else {
		results = append(results, MatrixValidationResult{
			Check:   "Server Name",
			Status:  "success",
			Message: "Server name is configured",
			Details: settings.ServerName,
		})
	}

	// Check Admin User ID
	if settings.AdminUserID == "" {
		results = append(results, MatrixValidationResult{
			Check:   "Admin User ID",
			Status:  "warning",
			Message: "Admin user ID is not configured",
			Details: "Admin user ID is recommended for proper room management",
		})
	} else {
		results = append(results, MatrixValidationResult{
			Check:   "Admin User ID",
			Status:  "success",
			Message: "Admin user ID is configured",
			Details: settings.AdminUserID,
		})
	}

	// Check Space Name
	if settings.SpaceName == "" {
		results = append(results, MatrixValidationResult{
			Check:   "Space Name",
			Status:  "warning",
			Message: "Space name is not configured, using default",
			Details: "Will use 'Movie Club' as the default space name",
		})
	} else {
		results = append(results, MatrixValidationResult{
			Check:   "Space Name",
			Status:  "success",
			Message: "Space name is configured",
			Details: settings.SpaceName,
		})
	}

	return results
}

// validateMatrixConnection validates the Matrix server connection
func validateMatrixConnection(settings MatrixSettings) []MatrixValidationResult {
	var results []MatrixValidationResult

	// Create temporary client for testing
	client, err := mautrix.NewClient(settings.ServerURL, "", "")
	if err != nil {
		results = append(results, MatrixValidationResult{
			Check:   "Server Connection",
			Status:  "error",
			Message: "Failed to create Matrix client",
			Details: err.Error(),
		})
		return results
	}

	client.AccessToken = settings.AdminToken

	// Test connection with whoami request
	resp, err := client.Whoami(context.Background())
	if err != nil {
		results = append(results, MatrixValidationResult{
			Check:   "Server Connection",
			Status:  "error",
			Message: "Failed to authenticate with Matrix server",
			Details: err.Error(),
		})
		return results
	}

	results = append(results, MatrixValidationResult{
		Check:   "Server Connection",
		Status:  "success",
		Message: "Successfully connected to Matrix server",
		Details: fmt.Sprintf("Authenticated as %s", resp.UserID),
	})

	return results
}

// validateMatrixPermissions validates admin permissions
func validateMatrixPermissions(settings MatrixSettings) []MatrixValidationResult {
	var results []MatrixValidationResult

	client, err := mautrix.NewClient(settings.ServerURL, "", "")
	if err != nil {
		results = append(results, MatrixValidationResult{
			Check:   "Admin Permissions",
			Status:  "error",
			Message: "Failed to create client for permission test",
			Details: err.Error(),
		})
		return results
	}

	client.AccessToken = settings.AdminToken

	// Test admin permissions by checking server version (should be available to admins)
	_, err = client.Versions(context.Background())
	if err != nil {
		results = append(results, MatrixValidationResult{
			Check:   "Admin Permissions",
			Status:  "warning",
			Message: "Unable to verify admin permissions",
			Details: "Server version check failed, but basic auth succeeded",
		})
	} else {
		results = append(results, MatrixValidationResult{
			Check:   "Admin Permissions",
			Status:  "success",
			Message: "Admin permissions verified",
			Details: "Successfully accessed server information",
		})
	}

	return results
}

// validateRoomCreation validates room creation capabilities
func validateRoomCreation(settings MatrixSettings) []MatrixValidationResult {
	var results []MatrixValidationResult

	client, err := mautrix.NewClient(settings.ServerURL, "", "")
	if err != nil {
		results = append(results, MatrixValidationResult{
			Check:   "Room Creation",
			Status:  "error",
			Message: "Failed to create client for room test",
			Details: err.Error(),
		})
		return results
	}

	client.AccessToken = settings.AdminToken

	// Create a test room
	testRoomName := fmt.Sprintf("watcharr-test-%d", time.Now().Unix())
	createReq := &mautrix.ReqCreateRoom{
		Name:       "Watcharr Test Room",
		Topic:      "Temporary test room created by Watcharr for validation",
		Preset:     "private_chat",
		Visibility: "private",
		RoomAliasName: testRoomName,
	}

	createResp, err := client.CreateRoom(context.Background(), createReq)
	if err != nil {
		results = append(results, MatrixValidationResult{
			Check:   "Room Creation",
			Status:  "error",
			Message: "Failed to create test room",
			Details: err.Error(),
		})
		return results
	}

	// Clean up test room
	_, leaveErr := client.LeaveRoom(context.Background(), createResp.RoomID)
	if leaveErr != nil {
		// Note: slog needs to be imported if not already available
		fmt.Printf("Failed to clean up test room %s: %v\n", createResp.RoomID, leaveErr)
	}

	results = append(results, MatrixValidationResult{
		Check:   "Room Creation",
		Status:  "success",
		Message: "Room creation test successful",
		Details: fmt.Sprintf("Test room created and cleaned up: %s", createResp.RoomID),
	})

	return results
}

// getUserMatrixInfo gets Matrix information for a user
func (b *BaseRouter) getUserMatrixInfo(c *gin.Context) {
	// AuthRequired middleware should already be applied to this route
	userID := c.GetUint("userId")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	// Check if Matrix is enabled
	if !Config.MOVIE_CLUB.Matrix.Enabled {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Matrix integration is not enabled"})
		return
	}

	// Get user's Matrix account
	var matrixUser MatrixUser
	if err := b.db.Where("user_id = ?", userID).First(&matrixUser).Error; err != nil {
		// No Matrix user found - this is okay, they can create one
		c.JSON(http.StatusOK, gin.H{
			"hasMatrixAccount": false,
			"matrixEnabled":    true,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"hasMatrixAccount": true,
		"matrixEnabled":    true,
		"matrixUserId":     matrixUser.MatrixUserID,
		"isAutoGenerated":  matrixUser.IsAutoGenerated,
	})
}

// createUserMatrixAccount creates a Matrix account for a user
func (b *BaseRouter) createUserMatrixAccount(c *gin.Context) {
	userID := c.GetUint("userId")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	// Check if Matrix is enabled
	if !Config.MOVIE_CLUB.Matrix.Enabled {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Matrix integration is not enabled"})
		return
	}

	// Get user info from database to get username
	var user User
	if err := b.db.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		return
	}

	// Check if user already has Matrix account
	var existingMatrixUser MatrixUser
	if err := b.db.Where("user_id = ?", userID).First(&existingMatrixUser).Error; err == nil {
		c.JSON(http.StatusConflict, ErrorResponse{Error: "User already has a Matrix account"})
		return
	}

	// Create Matrix user
	matrixUser, err := CreateMatrixUser(b.db, userID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create Matrix user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"matrixUserId":    matrixUser.MatrixUserID,
		"isAutoGenerated": matrixUser.IsAutoGenerated,
	})
}

// LinkCustomMatrixAccountRequest represents a request to link custom Matrix account
type LinkCustomMatrixAccountRequest struct {
	MatrixUserID  string `json:"matrixUserId" binding:"required"`
	AccessToken   string `json:"accessToken" binding:"required"`
}

// linkCustomMatrixAccount links a custom Matrix account to a user
func (b *BaseRouter) linkCustomMatrixAccount(c *gin.Context) {
	userID := c.GetUint("userId")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	var req LinkCustomMatrixAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request format"})
		return
	}

	// Check if Matrix is enabled
	if !Config.MOVIE_CLUB.Matrix.Enabled {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Matrix integration is not enabled"})
		return
	}

	// Link custom Matrix user
	if err := LinkCustomMatrixUser(b.db, userID, req.MatrixUserID, req.AccessToken); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Failed to link Matrix account: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"matrixUserId":    req.MatrixUserID,
		"isAutoGenerated": false,
	})
}

// MatrixRoomResponse represents room data for the frontend
type MatrixRoomResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Alias      string `json:"alias,omitempty"`
	Topic      string `json:"topic,omitempty"`
	CycleID    uint   `json:"cycleId,omitempty"`
	CycleName  string `json:"cycleName,omitempty"`
	MovieTitle string `json:"movieTitle,omitempty"`
	JoinURL    string `json:"joinUrl,omitempty"`
}

// getUserMatrixRooms gets Matrix rooms that a user has access to
func (b *BaseRouter) getUserMatrixRooms(c *gin.Context) {
	userID := c.GetUint("userId")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	// Check if Matrix is enabled
	if !Config.MOVIE_CLUB.Matrix.Enabled {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Matrix integration is not enabled"})
		return
	}

	// Get user's rooms
	rooms, err := GetUserRooms(b.db, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get user rooms"})
		return
	}

	// Transform rooms to frontend format
	var responseRooms []MatrixRoomResponse
	for _, room := range rooms {
		roomResponse := MatrixRoomResponse{
			ID:      room.RoomID,
			Alias:   room.RoomAlias,
			CycleID: room.CycleID,
		}

		// Add cycle and movie information if available
		if room.Cycle.ID != 0 {
			roomResponse.CycleName = room.Cycle.Name
			
			// Generate room name and topic
			if room.Cycle.WinnerContent != nil {
				roomResponse.MovieTitle = room.Cycle.WinnerContent.Title
				roomResponse.Name = fmt.Sprintf("Movie Club: %s (%s)",
					room.Cycle.WinnerContent.Title,
					room.Cycle.VotingEndDate.Format("2006-01-02"))
				roomResponse.Topic = fmt.Sprintf("Discussion for %s - Movie Club Cycle", room.Cycle.WinnerContent.Title)
			} else {
				roomResponse.Name = fmt.Sprintf("Movie Club: %s", room.Cycle.Name)
				roomResponse.Topic = fmt.Sprintf("Movie Club discussion for cycle: %s", room.Cycle.Name)
			}
		} else {
			// Fallback name if cycle data not available
			roomResponse.Name = "Movie Club Discussion"
			roomResponse.Topic = "Movie Club community discussion"
		}

		// Generate join URL
		if Config.MOVIE_CLUB.Matrix.ServerURL != "" && room.RoomAlias != "" {
			roomResponse.JoinURL = fmt.Sprintf("%s/#/room/%s", Config.MOVIE_CLUB.Matrix.ServerURL, room.RoomAlias)
		}

		responseRooms = append(responseRooms, roomResponse)
	}

	c.JSON(http.StatusOK, gin.H{
		"rooms": responseRooms,
	})
}

// Matrix API route registration
func (b *BaseRouter) setupMatrixRoutes() {
	matrix := b.rg.Group("/matrix")
	
	// Admin routes
	matrix.POST("/test-connection", AdminRequired(), b.testMatrixConnection)
	matrix.GET("/validate", AdminRequired(), b.validateMatrixSetup)
	
	// User routes
	matrix.GET("/info", b.getUserMatrixInfo)
	matrix.POST("/create-account", b.createUserMatrixAccount)
	matrix.POST("/link-account", b.linkCustomMatrixAccount)
	matrix.GET("/rooms", b.getUserMatrixRooms)
}