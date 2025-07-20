package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
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

	// 5. Application Service validation
	validation.Results = append(validation.Results, validateApplicationService(settings)...)

	// Note: Shared secret registration removed for simplicity
	// Matrix integration uses Application Service for virtual users and manual linking for real accounts

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

	// Note: Registration Secret removed - Matrix integration now uses Application Service for virtual users
	// and manual account linking for personal accounts

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

// validateApplicationService validates the Application Service configuration
func validateApplicationService(settings MatrixSettings) []MatrixValidationResult {
	var results []MatrixValidationResult

	// Check if Application Service is enabled
	if !settings.AppService.Enabled {
		results = append(results, MatrixValidationResult{
			Check:   "Application Service",
			Status:  "warning",
			Message: "Application Service is not enabled",
			Details: "AS provides virtual Matrix users. Without AS, only manual account linking is available.",
		})
		return results
	}

	// Check AS ID
	if settings.AppService.ID == "" {
		results = append(results, MatrixValidationResult{
			Check:   "AS ID",
			Status:  "error",
			Message: "Application Service ID is not configured",
			Details: "AS ID is required for Application Service registration",
		})
	} else {
		results = append(results, MatrixValidationResult{
			Check:   "AS ID",
			Status:  "success",
			Message: "Application Service ID is configured",
			Details: settings.AppService.ID,
		})
	}

	// Check AS Token
	if settings.AppService.AppServiceToken == "" {
		results = append(results, MatrixValidationResult{
			Check:   "AS Token",
			Status:  "error",
			Message: "Application Service token is not configured",
			Details: "AS token is required for AS-to-homeserver communication",
		})
	} else {
		results = append(results, MatrixValidationResult{
			Check:   "AS Token",
			Status:  "success",
			Message: "Application Service token is configured",
			Details: "Token provided (hidden for security)",
		})
	}

	// Check Homeserver Token
	if settings.AppService.HomeServerToken == "" {
		results = append(results, MatrixValidationResult{
			Check:   "Homeserver Token",
			Status:  "error",
			Message: "Homeserver token is not configured",
			Details: "HS token is required for homeserver-to-AS communication",
		})
	} else {
		results = append(results, MatrixValidationResult{
			Check:   "Homeserver Token",
			Status:  "success",
			Message: "Homeserver token is configured",
			Details: "Token provided (hidden for security)",
		})
	}

	// Check User Namespace
	if settings.AppService.UserNamespace == "" {
		results = append(results, MatrixValidationResult{
			Check:   "User Namespace",
			Status:  "error",
			Message: "User namespace is not configured",
			Details: "User namespace pattern is required (e.g., @watcharr_*:domain.com)",
		})
	} else {
		results = append(results, MatrixValidationResult{
			Check:   "User Namespace",
			Status:  "success",
			Message: "User namespace is configured",
			Details: settings.AppService.UserNamespace,
		})
	}

	// Check Alias Namespace
	if settings.AppService.AliasNamespace == "" {
		results = append(results, MatrixValidationResult{
			Check:   "Alias Namespace",
			Status:  "error",
			Message: "Alias namespace is not configured",
			Details: "Alias namespace pattern is required (e.g., #watcharr_*:domain.com)",
		})
	} else {
		results = append(results, MatrixValidationResult{
			Check:   "Alias Namespace",
			Status:  "success",
			Message: "Alias namespace is configured",
			Details: settings.AppService.AliasNamespace,
		})
	}

	// Check Sender Localpart
	if settings.AppService.SenderLocalpart == "" {
		results = append(results, MatrixValidationResult{
			Check:   "Sender Localpart",
			Status:  "error",
			Message: "Sender localpart is not configured",
			Details: "Sender localpart is required for AS bot user (e.g., watcharr-bot)",
		})
	} else {
		results = append(results, MatrixValidationResult{
			Check:   "Sender Localpart",
			Status:  "success",
			Message: "Sender localpart is configured",
			Details: settings.AppService.SenderLocalpart,
		})
	}

	// Validate AS configuration consistency if all basic fields are present
	if settings.AppService.ID != "" && settings.AppService.AppServiceToken != "" && 
	   settings.AppService.HomeServerToken != "" && settings.AppService.UserNamespace != "" {
		
		// Try to validate the AS configuration
		if appServiceManager != nil {
			err := appServiceManager.ValidateConfig()
			if err != nil {
				results = append(results, MatrixValidationResult{
					Check:   "AS Configuration",
					Status:  "error",
					Message: "Application Service configuration is invalid",
					Details: err.Error(),
				})
			} else {
				results = append(results, MatrixValidationResult{
					Check:   "AS Configuration",
					Status:  "success",
					Message: "Application Service configuration is valid",
					Details: "All AS settings are properly configured and validated",
				})
			}
		} else {
			results = append(results, MatrixValidationResult{
				Check:   "AS Configuration",
				Status:  "warning",
				Message: "Application Service not initialized",
				Details: "AS manager not initialized - restart server to enable AS features",
			})
		}
	}

	return results
}

// Note: validateSharedSecretRegistration removed - feature deprecated for simplicity

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

	// Check for Matrix account in both legacy and new models
	var matrixUser MatrixUser
	var matrixUserV2 MatrixUserV2
	
	hasLegacyAccount := b.db.Where("user_id = ?", userID).First(&matrixUser).Error == nil
	hasNewAccount := b.db.Where("user_id = ?", userID).First(&matrixUserV2).Error == nil
	
	if !hasLegacyAccount && !hasNewAccount {
		// No Matrix user found - this is okay, they can create one
		c.JSON(http.StatusOK, gin.H{
			"hasMatrixAccount": false,
			"matrixEnabled":    true,
		})
		return
	}

	// Return information from whichever account exists (prefer new model)
	if hasNewAccount {
		c.JSON(http.StatusOK, gin.H{
			"hasMatrixAccount": true,
			"matrixEnabled":    true,
			"matrixUserId":     matrixUserV2.MatrixUserID,
			"isAutoGenerated":  matrixUserV2.IsAutoGenerated,
			"accountType":      matrixUserV2.AccountType,
			"asManagedUser":    matrixUserV2.ASManagedUser,
		})
	} else {
		// Legacy account
		c.JSON(http.StatusOK, gin.H{
			"hasMatrixAccount": true,
			"matrixEnabled":    true,
			"matrixUserId":     matrixUser.MatrixUserID,
			"isAutoGenerated":  matrixUser.IsAutoGenerated,
			"accountType":      "legacy", // Indicate this is a legacy account
			"asManagedUser":    false,
		})
	}
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

	// Check if user already has Matrix account (check both old and new models)
	var existingMatrixUser MatrixUser
	var existingMatrixUserV2 MatrixUserV2
	hasLegacyAccount := b.db.Where("user_id = ?", userID).First(&existingMatrixUser).Error == nil
	hasNewAccount := b.db.Where("user_id = ?", userID).First(&existingMatrixUserV2).Error == nil
	
	if hasLegacyAccount || hasNewAccount {
		c.JSON(http.StatusConflict, ErrorResponse{Error: "User already has a Matrix account"})
		return
	}

	// Create Matrix user using hybrid system
	matrixUser, err := GetOrCreateMatrixUserV2(b.db, userID, user.Username)
	if err != nil {
		// Provide better error messages for different failure scenarios
		if strings.Contains(err.Error(), "user_exists_no_record") {
			c.JSON(http.StatusConflict, ErrorResponse{Error: "Matrix username already exists but no previous Watcharr record found. Please use a different username or link a custom Matrix account."})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create Matrix user"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"matrixUserId":    matrixUser.MatrixUserID,
		"isAutoGenerated": matrixUser.IsAutoGenerated,
		"accountType":     matrixUser.AccountType,
		"asManagedUser":   matrixUser.ASManagedUser,
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

// unlinkUserMatrixAccount removes a user's Matrix account link
func (b *BaseRouter) unlinkUserMatrixAccount(c *gin.Context) {
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

	// Remove user's Matrix account link
	if err := UnlinkMatrixUser(b.db, userID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to unlink Matrix account: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Matrix account unlinked successfully",
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
			ID:    room.RoomID,
			Alias: room.RoomAlias,
		}
		
		// Set CycleID if present
		if room.CycleID != nil {
			roomResponse.CycleID = *room.CycleID
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
			// Remove trailing slash from server URL to avoid double slashes
			serverURL := strings.TrimSuffix(Config.MOVIE_CLUB.Matrix.ServerURL, "/")
			roomResponse.JoinURL = fmt.Sprintf("%s/#/room/%s", serverURL, room.RoomAlias)
		}

		responseRooms = append(responseRooms, roomResponse)
	}

	c.JSON(http.StatusOK, gin.H{
		"rooms": responseRooms,
	})
}

// ExportCredentialsRequest represents a request to export Matrix credentials
type ExportCredentialsRequest struct {
	DeletePasswordAfterExport bool `json:"deletePasswordAfterExport"`
}

// exportUserMatrixCredentials exports Matrix credentials for auto-generated users
func (b *BaseRouter) exportUserMatrixCredentials(c *gin.Context) {
	if !Config.MOVIE_CLUB.Matrix.Enabled {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Matrix integration is not enabled"})
		return
	}

	userID := c.GetUint("userId")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User ID not found in session"})
		return
	}

	var req ExportCredentialsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request format"})
		return
	}

	// Export the credentials
	export, err := ExportMatrixCredentials(b.db, userID, req.DeletePasswordAfterExport)
	if err != nil {
		// Check for specific error types to provide better user messages
		errMsg := err.Error()
		statusCode := http.StatusInternalServerError
		
		switch {
		case err.Error() == "no Matrix account linked for this user":
			statusCode = http.StatusNotFound
			errMsg = "No Matrix account found for your user"
		case err.Error() == "credential export is only available for auto-generated Matrix accounts":
			statusCode = http.StatusForbidden
			errMsg = "Credential export is only available for auto-generated Matrix accounts"
		case err.Error() == "no password available for export - account may be a placeholder":
			statusCode = http.StatusBadRequest
			errMsg = "No password available for export - your Matrix account appears to be a placeholder"
		case err.Error() == "maximum number of password exports exceeded for security":
			statusCode = http.StatusTooManyRequests
			errMsg = "Maximum number of password exports exceeded for security reasons"
		case strings.Contains(fmt.Sprintf("%v", err), "please wait at least 5 minutes"):
			statusCode = http.StatusTooManyRequests
			errMsg = err.Error()
		}
		
		c.JSON(statusCode, ErrorResponse{Error: errMsg})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"credentials": export,
		"message": "Matrix credentials exported successfully",
	})
}

// createRoomsForExistingCycles creates Matrix rooms for active watching cycles that don't have them
func (b *BaseRouter) createRoomsForExistingCycles(c *gin.Context) {
	if !Config.MOVIE_CLUB.Matrix.Enabled {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Matrix integration is not enabled"})
		return
	}

	// Check if Matrix client is initialized
	if matrixClient == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "Matrix client is not initialized"})
		return
	}

	// First, get a preview of what cycles would be affected
	cycles, err := GetActiveWatchingCyclesWithoutRooms(b.db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to check for cycles needing rooms: " + err.Error()})
		return
	}

	if len(cycles) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "No active watching cycles found that need Matrix rooms",
			"result": map[string]interface{}{
				"totalCycles":    0,
				"createdRooms":   0,
				"failedRooms":    0,
				"createdRoomIds": []string{},
				"errors":         []string{},
			},
		})
		return
	}

	// Create rooms for the cycles
	result, err := CreateRoomsForExistingCycles(b.db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create rooms for existing cycles: " + err.Error()})
		return
	}

	// Determine response status based on results
	statusCode := http.StatusOK
	message := "Successfully created Matrix rooms for existing cycles"
	
	if result.FailedRooms > 0 && result.CreatedRooms == 0 {
		statusCode = http.StatusInternalServerError
		message = "Failed to create any rooms for existing cycles"
	} else if result.FailedRooms > 0 {
		statusCode = http.StatusPartialContent
		message = fmt.Sprintf("Created %d rooms successfully, but %d failed", result.CreatedRooms, result.FailedRooms)
	}

	c.JSON(statusCode, gin.H{
		"success": result.CreatedRooms > 0,
		"message": message,
		"result":  result,
	})
}

// migrateMatrixUsersToV2 migrates all legacy Matrix users to the new V2 system
func (b *BaseRouter) migrateMatrixUsersToV2(c *gin.Context) {
	// Check if Matrix is enabled
	if !Config.MOVIE_CLUB.Matrix.Enabled {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Matrix integration is not enabled"})
		return
	}

	slog.Info("Starting Matrix users migration to V2 system")

	// Perform the migration
	result, err := MigrateMatrixUsersToV2(b.db)
	if err != nil {
		slog.Error("Matrix user migration failed", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Migration failed: " + err.Error()})
		return
	}

	// Determine response status based on results
	statusCode := http.StatusOK
	message := "Matrix user migration completed successfully"
	
	if result.FailedMigrations > 0 && result.MigratedUsers == 0 {
		statusCode = http.StatusInternalServerError
		message = "Failed to migrate any Matrix users"
	} else if result.FailedMigrations > 0 {
		statusCode = http.StatusPartialContent
		message = fmt.Sprintf("Migrated %d users successfully, but %d failed", result.MigratedUsers, result.FailedMigrations)
	} else if result.TotalLegacyUsers == 0 {
		message = "No legacy Matrix users found to migrate"
	}

	slog.Info("Matrix user migration completed", 
		"total_users", result.TotalLegacyUsers,
		"migrated", result.MigratedUsers,
		"skipped", result.SkippedUsers,
		"failed", result.FailedMigrations)

	c.JSON(statusCode, gin.H{
		"success": result.MigratedUsers > 0 || result.TotalLegacyUsers == 0,
		"message": message,
		"result":  result,
	})
}

// Matrix API route registration
func (b *BaseRouter) setupMatrixRoutes() {
	matrix := b.rg.Group("/matrix")
	
	// Admin routes
	matrix.POST("/test-connection", AuthRequired(b.db), AdminRequired(), b.testMatrixConnection)
	matrix.GET("/validate", AuthRequired(b.db), AdminRequired(), b.validateMatrixSetup)
	matrix.POST("/create-rooms-for-existing-cycles", AuthRequired(b.db), AdminRequired(), b.createRoomsForExistingCycles)
	matrix.POST("/migrate-users-to-v2", AuthRequired(b.db), AdminRequired(), b.migrateMatrixUsersToV2)
	matrix.GET("/registration-file", AuthRequired(b.db), AdminRequired(), b.generateRegistrationFile)
	
	// User routes
	matrix.GET("/info", AuthRequired(b.db), b.getUserMatrixInfo)
	matrix.POST("/create-account", AuthRequired(b.db), b.createUserMatrixAccount)
	matrix.POST("/link-account", AuthRequired(b.db), b.linkCustomMatrixAccount)
	matrix.DELETE("/unlink-account", AuthRequired(b.db), b.unlinkUserMatrixAccount)
	matrix.POST("/export-credentials", AuthRequired(b.db), b.exportUserMatrixCredentials)
	matrix.GET("/rooms", AuthRequired(b.db), b.getUserMatrixRooms)
}

// generateRegistrationFile generates and returns the Matrix Application Service registration file
func (b *BaseRouter) generateRegistrationFile(c *gin.Context) {
	if !Config.MOVIE_CLUB.Matrix.Enabled {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Matrix integration is not enabled"})
		return
	}

	settings := Config.MOVIE_CLUB.Matrix
	asSettings := settings.AppService

	// Generate registration file content
	registrationFile, err := GenerateASRegistrationFile(&asSettings, &settings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to generate registration file: " + err.Error()})
		return
	}

	// Set headers for file download
	c.Header("Content-Disposition", "attachment; filename=watcharr-registration.yaml")
	c.Header("Content-Type", "application/x-yaml")
	c.String(http.StatusOK, registrationFile)
}