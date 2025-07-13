package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Matrix Application Service implementation for Watcharr
// This provides virtual Matrix users for Movie Club functionality
// without requiring real Matrix accounts that users can access externally.

// Use AppServiceSettings from movie_club.go instead of defining our own config

// Use MatrixAccountType and MatrixUserV2 from matrix.go

// Application Service Manager
type AppServiceManager struct {
	config    *AppServiceSettings
	db        *gorm.DB
	router    *gin.Engine
	isRunning bool
}

// Initialize the Application Service
func NewAppServiceManager(db *gorm.DB, config *AppServiceSettings) *AppServiceManager {
	return &AppServiceManager{
		config: config,
		db:     db,
	}
}

// Application Service HTTP Endpoints
// These endpoints are called by the Matrix homeserver

// Transaction handling for AS events
type ASTransaction struct {
	Events []ASEvent `json:"events"`
	TxnID  string    `json:"txn_id"`
}

// Matrix event received via AS
type ASEvent struct {
	Type     string                 `json:"type"`
	EventID  string                 `json:"event_id"`
	Sender   string                 `json:"sender"`
	RoomID   string                 `json:"room_id"`
	Content  map[string]interface{} `json:"content"`
	StateKey *string                `json:"state_key,omitempty"`
}

// AS Query endpoints response
type ASUserQueryResponse struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

type ASRoomQueryResponse struct {
	RoomAlias string `json:"room_alias"`
	RoomID    string `json:"room_id,omitempty"`
}

// Validation functions

// ValidateAppServiceConfig validates the AS configuration
func (asm *AppServiceManager) ValidateConfig() error {
	if asm.config == nil {
		return errors.New("application service configuration is nil")
	}

	if !asm.config.Enabled {
		return errors.New("application service is disabled")
	}

	if asm.config.ID == "" {
		return errors.New("application service ID is required")
	}

	if asm.config.AppServiceToken == "" {
		return errors.New("application service token is required")
	}

	if asm.config.HomeServerToken == "" {
		return errors.New("homeserver token is required")
	}

	if asm.config.UserNamespace == "" {
		return errors.New("user namespace is required")
	}

	if asm.config.SenderLocalpart == "" {
		return errors.New("sender localpart is required")
	}

	return nil
}

// Authentication middleware for AS endpoints
func (asm *AppServiceManager) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate homeserver token
		authHeader := c.GetHeader("Authorization")
		expectedToken := "Bearer " + asm.config.HomeServerToken

		if authHeader != expectedToken {
			slog.Warn("Invalid AS authentication attempt",
				"remote_addr", c.ClientIP(),
				"auth_header_present", authHeader != "")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid homeserver token"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// HTTP Handlers for Application Service endpoints

// PUT /_matrix/app/v1/transactions/{txnId}
func (asm *AppServiceManager) handleTransaction(c *gin.Context) {
	txnID := c.Param("txnId")

	var transaction ASTransaction
	if err := c.ShouldBindJSON(&transaction); err != nil {
		slog.Error("Failed to parse AS transaction", "error", err, "txn_id", txnID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction format"})
		return
	}

	transaction.TxnID = txnID

	slog.Debug("Received AS transaction",
		"txn_id", txnID,
		"event_count", len(transaction.Events))

	// Process each event in the transaction
	for _, event := range transaction.Events {
		if err := asm.processEvent(event); err != nil {
			slog.Error("Failed to process AS event",
				"error", err,
				"event_id", event.EventID,
				"event_type", event.Type,
				"txn_id", txnID)
			// Continue processing other events even if one fails
		}
	}

	// Return empty JSON object to indicate success
	c.JSON(http.StatusOK, gin.H{})
}

// GET /_matrix/app/v1/users/{userId}
func (asm *AppServiceManager) handleUserQuery(c *gin.Context) {
	userID := c.Param("userId")

	slog.Debug("AS user query", "user_id", userID)

	// Check if this user is in our namespace
	if !asm.isUserInNamespace(userID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Check if user exists in our database
	var matrixUser MatrixUserV2
	if err := asm.db.Where("matrix_user_id = ? AND account_type = ?", userID, AccountTypeAppService).First(&matrixUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		slog.Error("Database error during user query", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Load the associated Watcharr user for display name
	if err := asm.db.Preload("User").First(&matrixUser, matrixUser.ID).Error; err != nil {
		slog.Warn("Failed to load user details", "error", err, "user_id", userID)
	}

	response := ASUserQueryResponse{
		UserID: userID,
	}

	// Set display name from Watcharr username
	if matrixUser.User.Username != "" {
		response.DisplayName = fmt.Sprintf("%s (Watcharr)", matrixUser.User.Username)
	}

	c.JSON(http.StatusOK, response)
}

// GET /_matrix/app/v1/rooms/{roomAlias}
func (asm *AppServiceManager) handleRoomQuery(c *gin.Context) {
	roomAlias := c.Param("roomAlias")

	slog.Debug("AS room query", "room_alias", roomAlias)

	// Check if this room alias is in our namespace
	if !asm.isRoomAliasInNamespace(roomAlias) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	// For now, we don't create rooms on demand via AS
	// Movie Club rooms are created explicitly
	c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
}

// Helper functions

// Process an individual event from the homeserver
func (asm *AppServiceManager) processEvent(event ASEvent) error {
	slog.Debug("Processing AS event",
		"type", event.Type,
		"event_id", event.EventID,
		"sender", event.Sender,
		"room_id", event.RoomID)

	switch event.Type {
	case "m.room.member":
		return asm.processMembershipEvent(event)
	case "m.room.message":
		return asm.processMessageEvent(event)
	default:
		// Log unknown events but don't fail
		slog.Debug("Ignoring unknown event type", "type", event.Type)
		return nil
	}
}

// Process room membership events
func (asm *AppServiceManager) processMembershipEvent(event ASEvent) error {
	membership, ok := event.Content["membership"].(string)
	if !ok {
		return errors.New("invalid membership event: missing membership field")
	}

	stateKey := ""
	if event.StateKey != nil {
		stateKey = *event.StateKey
	}

	slog.Debug("Processing membership event",
		"membership", membership,
		"state_key", stateKey,
		"sender", event.Sender,
		"room_id", event.RoomID)

	// Update last seen timestamp if this is one of our AS users
	if asm.isUserInNamespace(stateKey) {
		asm.updateUserLastSeen(stateKey)
	}

	// Handle specific membership changes
	switch membership {
	case "join":
		return asm.handleUserJoined(event.RoomID, stateKey)
	case "leave":
		return asm.handleUserLeft(event.RoomID, stateKey)
	default:
		return nil
	}
}

// Process message events (for activity tracking)
func (asm *AppServiceManager) processMessageEvent(event ASEvent) error {
	// Update last seen for AS users who send messages
	if asm.isUserInNamespace(event.Sender) {
		asm.updateUserLastSeen(event.Sender)
	}

	// Message events are primarily for activity tracking
	// Movie Club specific logic would go here
	return nil
}

// Handle user joining a room
func (asm *AppServiceManager) handleUserJoined(roomID, userID string) error {
	slog.Debug("AS user joined room", "user_id", userID, "room_id", roomID)
	// Implementation for tracking room memberships
	return nil
}

// Handle user leaving a room
func (asm *AppServiceManager) handleUserLeft(roomID, userID string) error {
	slog.Debug("AS user left room", "user_id", userID, "room_id", roomID)
	// Implementation for tracking room memberships
	return nil
}

// Update the last seen timestamp for an AS user
func (asm *AppServiceManager) updateUserLastSeen(userID string) {
	if err := asm.db.Model(&MatrixUserV2{}).
		Where("matrix_user_id = ? AND account_type = ?", userID, AccountTypeAppService).
		Update("last_seen_at", time.Now()).Error; err != nil {
		slog.Warn("Failed to update user last seen", "error", err, "user_id", userID)
	}
}

// Check if a user ID is in our namespace
func (asm *AppServiceManager) isUserInNamespace(userID string) bool {
	if asm.config.UserNamespace == "" {
		return false
	}

	// Extract just the localpart prefix from namespace pattern
	// "@watcharr_*:server.com" becomes "@watcharr_"
	namespace := asm.config.UserNamespace
	
	// First remove server part if present
	if colonIndex := strings.Index(namespace, ":"); colonIndex != -1 {
		namespace = namespace[:colonIndex] // "@watcharr_*"
	}
	
	// Then remove the wildcard
	namespace = strings.TrimSuffix(namespace, "*") // "@watcharr_"
	
	return strings.HasPrefix(userID, namespace)
}

// Check if a room alias is in our namespace
func (asm *AppServiceManager) isRoomAliasInNamespace(roomAlias string) bool {
	if asm.config.AliasNamespace == "" {
		return false
	}

	namespace := asm.config.AliasNamespace
	
	// First remove server part if present
	if colonIndex := strings.Index(namespace, ":"); colonIndex != -1 {
		namespace = namespace[:colonIndex]
	}
	
	// Then remove the wildcard
	namespace = strings.TrimSuffix(namespace, "*")
	
	return strings.HasPrefix(roomAlias, namespace)
}

// User Management Functions

// Create a new Application Service user
func (asm *AppServiceManager) CreateUser(watcharrUserID uint, username string) (*MatrixUserV2, error) {
	if err := asm.ValidateConfig(); err != nil {
		return nil, fmt.Errorf("invalid AS configuration: %w", err)
	}

	// Generate Matrix user ID in our namespace
	serverName := extractServerFromNamespace(asm.config.UserNamespace)
	if serverName == "" {
		return nil, errors.New("cannot extract server name from user namespace")
	}

	// Create localpart: watcharr_{userID}_{sanitized_username}
	sanitizedUsername := sanitizeUsernameForAS(username)
	localpart := fmt.Sprintf("watcharr_%d_%s", watcharrUserID, sanitizedUsername)
	matrixUserID := fmt.Sprintf("@%s:%s", localpart, serverName)

	// Check if user already exists
	var existingUser MatrixUserV2
	if err := asm.db.Where("user_id = ? AND account_type = ?", watcharrUserID, AccountTypeAppService).First(&existingUser).Error; err == nil {
		return nil, errors.New("user already has an Application Service account")
	}

	// Create the AS user record
	matrixUser := &MatrixUserV2{
		UserID:          watcharrUserID,
		MatrixUserID:    matrixUserID,
		AccountType:     AccountTypeAppService,
		IsAutoGenerated: true,
		ASManagedUser:   true,
		CreatedViaAS:    true,
		LastSeenAt:      time.Now(),
	}

	if err := asm.db.Create(matrixUser).Error; err != nil {
		return nil, fmt.Errorf("failed to create AS user: %w", err)
	}

	slog.Info("Created Application Service user",
		"watcharr_user_id", watcharrUserID,
		"matrix_user_id", matrixUserID,
		"username", username)

	return matrixUser, nil
}

// Delete an Application Service user
func (asm *AppServiceManager) DeleteUser(watcharrUserID uint) error {
	var matrixUser MatrixUserV2
	if err := asm.db.Where("user_id = ? AND account_type = ?", watcharrUserID, AccountTypeAppService).First(&matrixUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("no Application Service account found for user")
		}
		return fmt.Errorf("database error: %w", err)
	}

	// Delete the user record
	if err := asm.db.Delete(&matrixUser).Error; err != nil {
		return fmt.Errorf("failed to delete AS user: %w", err)
	}

	slog.Info("Deleted Application Service user",
		"watcharr_user_id", watcharrUserID,
		"matrix_user_id", matrixUser.MatrixUserID)

	return nil
}

// Utility functions

// Extract server name from namespace pattern
func extractServerFromNamespace(namespace string) string {
	// Extract server from "@watcharr_*:server.name" format
	if !strings.Contains(namespace, ":") {
		return ""
	}
	parts := strings.SplitN(namespace, ":", 2)
	return parts[1]
}

// Sanitize username for AS localpart
func sanitizeUsernameForAS(username string) string {
	// Matrix localparts can only contain [a-z0-9._=-]
	var result strings.Builder
	for _, r := range strings.ToLower(username) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '.' || r == '_' || r == '=' || r == '-' {
			result.WriteRune(r)
		} else {
			result.WriteRune('_')
		}
	}
	return result.String()
}

// Setup HTTP routes for the Application Service
func (asm *AppServiceManager) SetupRoutes(router *gin.Engine) {
	// Application Service endpoints
	asGroup := router.Group("/_matrix/app/v1")
	asGroup.Use(asm.authMiddleware())

	asGroup.PUT("/transactions/:txnId", asm.handleTransaction)
	asGroup.GET("/users/:userId", asm.handleUserQuery)
	asGroup.GET("/rooms/:roomAlias", asm.handleRoomQuery)

	slog.Info("Application Service routes configured")
}

// Start the Application Service
func (asm *AppServiceManager) Start() error {
	if err := asm.ValidateConfig(); err != nil {
		return fmt.Errorf("cannot start AS: %w", err)
	}

	asm.isRunning = true
	slog.Info("Application Service started",
		"id", asm.config.ID,
		"user_namespace", asm.config.UserNamespace,
		"alias_namespace", asm.config.AliasNamespace)

	return nil
}

// Stop the Application Service
func (asm *AppServiceManager) Stop() {
	asm.isRunning = false
	slog.Info("Application Service stopped")
}

// Check if the AS is running
func (asm *AppServiceManager) IsRunning() bool {
	return asm.isRunning
}

// Global AS manager instance
var appServiceManager *AppServiceManager

// Initialize the global AS manager
func InitializeAppService(db *gorm.DB, config *AppServiceSettings) error {
	appServiceManager = NewAppServiceManager(db, config)

	if config.Enabled {
		return appServiceManager.Start()
	}

	return nil
}