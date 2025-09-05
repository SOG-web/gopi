package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gopi.com/api/http/dto"
	"gopi.com/internal/app/challenge"
	"gopi.com/internal/app/user"
	"gopi.com/internal/apperr"
	challengeModel "gopi.com/internal/domain/challenge/model"
	userModel "gopi.com/internal/domain/user/model"
)

type ChallengeHandler struct {
	challengeService *challenge.ChallengeService
	userService      *user.UserService
}

func NewChallengeHandler(challengeService *challenge.ChallengeService, userService *user.UserService) *ChallengeHandler {
	return &ChallengeHandler{
		challengeService: challengeService,
		userService:      userService,
	}
}

// CreateChallenge godoc
// @Summary Create a new challenge
// @Description Create a new challenge with the provided details
// @Tags challenges
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param challenge body dto.CreateChallengeRequest true "Challenge details"
// @Success 201 {object} dto.ChallengeResponse "Challenge created successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request body"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /challenges [post]
func (h *ChallengeHandler) CreateChallenge(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		respondError(c, apperr.E("CreateChallenge", apperr.Unauthorized, nil, "User not authenticated"))
		return
	}

	var req dto.CreateChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperr.E("CreateChallenge", apperr.InvalidInput, err, "Invalid request body"))
		return
	}

	// Create challenge
	challenge, err := h.challengeService.CreateChallenge(
		userID.(string),
		req.Name,
		req.Description,
		req.Condition,
		req.Goal,
		req.Location,
		challengeModel.ChallengeMode(req.Mode),
		req.DistanceToCover,
		req.TargetAmount,
		req.TargetAmountPerKm,
		req.StartDuration,
		req.EndDuration,
		req.NoOfWinner,
	)
	if err != nil {
		respondError(c, apperr.E("CreateChallenge", apperr.Internal, err, "Failed to create challenge"))
		return
	}

	// Get user details for response
	user, err := h.userService.GetUserByID(userID.(string))
	if err != nil {
		respondError(c, apperr.E("CreateChallenge", apperr.Internal, err, "Failed to get user details"))
		return
	}

	response := h.challengeToResponse(challenge, user)
	c.JSON(http.StatusCreated, response)
}

// GetChallenges godoc
// @Summary List all challenges
// @Description Get a paginated list of all challenges
// @Tags challenges
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(10)
// @Success 200 {object} dto.ChallengeListResponse "Challenges retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /challenges [get]
func (h *ChallengeHandler) GetChallenges(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	challenges, err := h.challengeService.ListChallenges(limit, offset)
	if err != nil {
		respondError(c, apperr.E("GetChallenges", apperr.Internal, err, "Failed to retrieve challenges"))
		return
	}

	response := dto.ChallengeListResponse{
		Challenges: make([]dto.ChallengeResponse, 0, len(challenges)),
		Page:       page,
		Limit:      limit,
		Total:      len(challenges),
	}

	for _, challenge := range challenges {
		// Get owner details for each challenge
		owner, err := h.userService.GetUserByID(challenge.OwnerID)
		if err != nil {
			continue // Skip if owner not found
		}
		response.Challenges = append(response.Challenges, h.challengeToResponse(challenge, owner))
	}

	c.JSON(http.StatusOK, response)
}

// GetChallengeByID godoc
// @Summary Get challenge by ID
// @Description Get a specific challenge by its ID
// @Tags challenges
// @Accept json
// @Produce json
// @Param id path string true "Challenge ID"
// @Success 200 {object} dto.ChallengeResponse "Challenge retrieved successfully"
// @Failure 404 {object} dto.ErrorResponse "Challenge not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /challenges/id/{id} [get]
func (h *ChallengeHandler) GetChallengeByID(c *gin.Context) {
	id := c.Param("id")

	challenge, err := h.challengeService.GetChallengeByID(id)
	if err != nil {
		respondError(c, apperr.E("GetChallengeByID", apperr.NotFound, err, "Challenge not found"))
		return
	}

	// Get owner details
	owner, err := h.userService.GetUserByID(challenge.OwnerID)
	if err != nil {
		respondError(c, apperr.E("GetChallengeByID", apperr.Internal, err, "Failed to get owner details"))
		return
	}

	response := h.challengeToResponse(challenge, owner)
	c.JSON(http.StatusOK, response)
}

// GetChallengeBySlug godoc
// @Summary Get challenge by slug
// @Description Get a specific challenge by its slug
// @Tags challenges
// @Accept json
// @Produce json
// @Param slug path string true "Challenge slug"
// @Success 200 {object} dto.ChallengeResponse "Challenge retrieved successfully"
// @Failure 404 {object} dto.ErrorResponse "Challenge not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /challenges/slug/{slug} [get]
func (h *ChallengeHandler) GetChallengeBySlug(c *gin.Context) {
	slug := c.Param("slug")

	challenge, err := h.challengeService.GetChallengeBySlug(slug)
	if err != nil {
		respondError(c, apperr.E("GetChallengeBySlug", apperr.NotFound, err, "Challenge not found"))
		return
	}

	// Get owner details
	owner, err := h.userService.GetUserByID(challenge.OwnerID)
	if err != nil {
		respondError(c, apperr.E("GetChallengeBySlug", apperr.Internal, err, "Failed to get owner details"))
		return
	}

	response := h.challengeToResponse(challenge, owner)
	c.JSON(http.StatusOK, response)
}

// JoinChallenge godoc
// @Summary Join a challenge
// @Description Join a challenge as a member
// @Tags challenges
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Challenge ID"
// @Success 200 {object} dto.MessageResponse "Successfully joined challenge"
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Challenge not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /challenges/{id}/join [post]
func (h *ChallengeHandler) JoinChallenge(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		respondError(c, apperr.E("JoinChallenge", apperr.Unauthorized, nil, "User not authenticated"))
		return
	}

	challengeID := c.Param("id")

	err := h.challengeService.JoinChallenge(challengeID, userID.(string))
	if err != nil {
		respondError(c, apperr.E("JoinChallenge", apperr.Internal, err, "Failed to join challenge"))
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Successfully joined challenge",
	})
}

// CreateCause godoc
// @Summary Create a new cause
// @Description Create a new cause within a challenge
// @Tags causes
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param cause body dto.CreateCauseRequest true "Cause details"
// @Success 201 {object} dto.CauseResponse "Cause created successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request body"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /causes [post]
func (h *ChallengeHandler) CreateCause(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		respondError(c, apperr.E("CreateCause", apperr.Unauthorized, nil, "User not authenticated"))
		return
	}

	var req dto.CreateCauseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperr.E("CreateCause", apperr.InvalidInput, err, "Invalid request body"))
		return
	}

	// Create cause
	cause, err := h.challengeService.CreateCause(
		req.ChallengeID,
		userID.(string),
		req.Name,
		req.Problem,
		req.Solution,
		req.ProductDescription,
		challengeModel.Activity(req.Activity),
		req.Location,
		req.Description,
		req.IsCommercial,
		req.AmountPerPiece,
		req.FundAmount,
		req.WillingAmount,
		req.UnitPrice,
	)
	if err != nil {
		respondError(c, apperr.E("CreateCause", apperr.Internal, err, "Failed to create cause"))
		return
	}

	// Get user details for response
	user, err := h.userService.GetUserByID(userID.(string))
	if err != nil {
		respondError(c, apperr.E("CreateCause", apperr.Internal, err, "Failed to get user details"))
		return
	}

	response := h.causeToResponse(cause, user)
	c.JSON(http.StatusCreated, response)
}

// GetCausesByChallenge godoc
// @Summary Get causes by challenge
// @Description Get all causes for a specific challenge
// @Tags causes
// @Accept json
// @Produce json
// @Param challenge_id path string true "Challenge ID"
// @Success 200 {object} dto.CauseListResponse "Causes retrieved successfully"
// @Failure 404 {object} dto.ErrorResponse "Challenge not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /challenges/{challenge_id}/causes [get]
func (h *ChallengeHandler) GetCausesByChallenge(c *gin.Context) {
	challengeID := c.Param("challenge_id")

	causes, err := h.challengeService.GetCausesByChallenge(challengeID)
	if err != nil {
		respondError(c, apperr.E("GetCausesByChallenge", apperr.Internal, err, "Failed to retrieve causes"))
		return
	}

	response := dto.CauseListResponse{
		Causes: make([]dto.CauseResponse, 0, len(causes)),
	}

	for _, cause := range causes {
		// Get owner details for each cause
		owner, err := h.userService.GetUserByID(cause.OwnerID)
		if err != nil {
			continue // Skip if owner not found
		}
		response.Causes = append(response.Causes, h.causeToResponse(cause, owner))
	}

	c.JSON(http.StatusOK, response)
}

// GetCauseByID godoc
// @Summary Get cause by ID
// @Description Get a specific cause by its ID
// @Tags causes
// @Accept json
// @Produce json
// @Param id path string true "Cause ID"
// @Success 200 {object} dto.CauseResponse "Cause retrieved successfully"
// @Failure 404 {object} dto.ErrorResponse "Cause not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /causes/{id} [get]
func (h *ChallengeHandler) GetCauseByID(c *gin.Context) {
	id := c.Param("id")

	cause, err := h.challengeService.GetCauseByID(id)
	if err != nil {
		respondError(c, apperr.E("GetCauseByID", apperr.NotFound, err, "Cause not found"))
		return
	}

	// Get owner details
	owner, err := h.userService.GetUserByID(cause.OwnerID)
	if err != nil {
		respondError(c, apperr.E("GetCauseByID", apperr.Internal, err, "Failed to get owner details"))
		return
	}

	response := h.causeToResponse(cause, owner)
	c.JSON(http.StatusOK, response)
}

// JoinCause godoc
// @Summary Join a cause
// @Description Join a cause as a member
// @Tags causes
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Cause ID"
// @Success 200 {object} dto.MessageResponse "Successfully joined cause"
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Cause not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /causes/{id}/join [post]
func (h *ChallengeHandler) JoinCause(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		respondError(c, apperr.E("JoinCause", apperr.Unauthorized, nil, "User not authenticated"))
		return
	}

	causeID := c.Param("id")

	err := h.challengeService.JoinCause(causeID, userID.(string))
	if err != nil {
		respondError(c, apperr.E("JoinCause", apperr.Internal, err, "Failed to join cause"))
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Successfully joined cause",
	})
}

// RecordCauseActivity godoc
// @Summary Record cause activity
// @Description Record activity for a cause (distance covered, etc.)
// @Tags causes
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param activity body dto.RecordActivityRequest true "Activity details"
// @Success 200 {object} dto.MessageResponse "Activity recorded successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request body"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /causes/activity [post]
func (h *ChallengeHandler) RecordCauseActivity(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		respondError(c, apperr.E("RecordCauseActivity", apperr.Unauthorized, nil, "User not authenticated"))
		return
	}

	var req dto.RecordActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperr.E("RecordCauseActivity", apperr.InvalidInput, err, "Invalid request body"))
		return
	}

	err := h.challengeService.RecordCauseActivity(
		req.CauseID,
		userID.(string),
		req.DistanceToCover,
		req.DistanceCovered,
		req.Duration,
		req.Activity,
	)
	if err != nil {
		respondError(c, apperr.E("RecordCauseActivity", apperr.Internal, err, "Failed to record activity"))
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Activity recorded successfully",
	})
}

// SponsorChallenge godoc
// @Summary Sponsor a challenge
// @Description Sponsor a challenge with amount per kilometer
// @Tags sponsorship
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param sponsorship body dto.SponsorChallengeRequest true "Sponsorship details"
// @Success 200 {object} dto.MessageResponse "Challenge sponsored successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request body"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /challenges/sponsor [post]
func (h *ChallengeHandler) SponsorChallenge(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		respondError(c, apperr.E("SponsorChallenge", apperr.Unauthorized, nil, "User not authenticated"))
		return
	}

	var req dto.SponsorChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperr.E("SponsorChallenge", apperr.InvalidInput, err, "Invalid request body"))
		return
	}

	err := h.challengeService.SponsorChallenge(
		req.ChallengeID,
		userID.(string),
		req.Distance,
		req.AmountPerKm,
	)
	if err != nil {
		respondError(c, apperr.E("SponsorChallenge", apperr.Internal, err, "Failed to sponsor challenge"))
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Challenge sponsored successfully",
	})
}

// SponsorCause godoc
// @Summary Sponsor a cause
// @Description Sponsor a cause with amount per kilometer
// @Tags sponsorship
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param sponsorship body dto.SponsorCauseRequest true "Sponsorship details"
// @Success 200 {object} dto.MessageResponse "Cause sponsored successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request body"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /causes/sponsor [post]
func (h *ChallengeHandler) SponsorCause(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		respondError(c, apperr.E("SponsorCause", apperr.Unauthorized, nil, "User not authenticated"))
		return
	}

	var req dto.SponsorCauseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperr.E("SponsorCause", apperr.InvalidInput, err, "Invalid request body"))
		return
	}

	err := h.challengeService.SponsorCause(
		req.CauseID,
		userID.(string),
		req.Distance,
		req.AmountPerKm,
	)
	if err != nil {
		respondError(c, apperr.E("SponsorCause", apperr.Internal, err, "Failed to sponsor cause"))
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Cause sponsored successfully",
	})
}

// BuyCause godoc
// @Summary Buy a cause
// @Description Purchase a cause with specified amount
// @Tags purchase
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param purchase body dto.BuyCauseRequest true "Purchase details"
// @Success 200 {object} dto.MessageResponse "Cause purchased successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request body"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /causes/buy [post]
func (h *ChallengeHandler) BuyCause(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		respondError(c, apperr.E("BuyCause", apperr.Unauthorized, nil, "User not authenticated"))
		return
	}

	var req dto.BuyCauseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperr.E("BuyCause", apperr.InvalidInput, err, "Invalid request body"))
		return
	}

	err := h.challengeService.BuyCause(req.CauseID, userID.(string), req.Amount)
	if err != nil {
		respondError(c, apperr.E("BuyCause", apperr.Internal, err, "Failed to buy cause"))
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Cause purchased successfully",
	})
}

// GetLeaderboard godoc
// @Summary Get leaderboard
// @Description Get the leaderboard of top performers
// @Tags leaderboard
// @Accept json
// @Produce json
// @Success 200 {object} dto.LeaderboardResponse "Leaderboard retrieved successfully"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /challenges/leaderboard [get]
func (h *ChallengeHandler) GetLeaderboard(c *gin.Context) {
	runners, err := h.challengeService.GetLeaderboard()
	if err != nil {
		respondError(c, apperr.E("GetLeaderboard", apperr.Internal, err, "Failed to retrieve leaderboard"))
		return
	}

	response := dto.LeaderboardResponse{
		Runners: make([]dto.CauseRunnerResponse, 0, len(runners)),
	}

	for _, runner := range runners {
		// Get user details for each runner
		user, err := h.userService.GetUserByID(runner.OwnerID)
		if err != nil {
			continue // Skip if user not found
		}
		response.Runners = append(response.Runners, h.causeRunnerToResponse(runner, user))
	}

	c.JSON(http.StatusOK, response)
}

// Helper functions for response conversion
func (h *ChallengeHandler) challengeToResponse(challenge *challengeModel.Challenge, owner *userModel.User) dto.ChallengeResponse {
	return dto.ChallengeResponse{
		ID:                challenge.ID,
		Name:              challenge.Name,
		Description:       challenge.Description,
		Mode:              string(challenge.Mode),
		Condition:         challenge.Condition,
		Goal:              challenge.Goal,
		Location:          challenge.Location,
		DistanceToCover:   challenge.DistanceToCover,
		TargetAmount:      challenge.TargetAmount,
		TargetAmountPerKm: challenge.TargetAmountPerKm,
		StartDuration:     challenge.StartDuration,
		EndDuration:       challenge.EndDuration,
		NoOfWinner:        challenge.NoOfWinner,
		WinningPrice:      challenge.WinningPrice,
		CausePrice:        challenge.CausePrice,
		CoverImage:        challenge.CoverImage,
		VideoUrl:          challenge.VideoUrl,
		Slug:              challenge.Slug,
		Owner:             h.userToChallengeOwnerInfo(owner),
		Members:           h.convertToChallengeMemberInfo(challenge.Members),
		Sponsors:          h.convertToChallengeSponsorInfo(challenge.Sponsors),
		DateCreated:       challenge.CreatedAt,
		DateUpdated:       challenge.UpdatedAt,
	}
}

func (h *ChallengeHandler) causeToResponse(cause *challengeModel.Cause, owner *userModel.User) dto.CauseResponse {
	return dto.CauseResponse{
		ID:                 cause.ID,
		ChallengeID:        cause.ChallengeID,
		Name:               cause.Name,
		Problem:            cause.Problem,
		Solution:           cause.Solution,
		ProductDescription: cause.ProductDescription,
		Activity:           string(cause.Activity),
		Location:           cause.Location,
		Description:        cause.Description,
		IsCommercial:       cause.IsCommercial,
		WhoIdeaImpact:      cause.WhoIdeaImpact,
		BuyerUser:          cause.BuyerUser,
		DistanceCovered:    cause.DistanceCovered,
		AmountPerPiece:     cause.AmountPerPiece,
		Duration:           cause.Duration,
		FundCause:          cause.FundCause,
		FundAmount:         cause.FundAmount,
		WillingAmount:      cause.WillingAmount,
		UnitPrice:          cause.UnitPrice,
		CostToLaunch:       cause.CostToLaunch,
		BenefitDesc:        cause.BenefitDesc,
		WorkoutImg:         cause.WorkoutImg,
		VideoUrl:           cause.VideoUrl,
		Slug:               cause.Slug,
		Owner:              h.userToCauseOwnerInfo(owner),
		Members:            h.convertToCauseMemberInfo(cause.Members),
		Sponsors:           h.convertToCauseSponsorInfo(cause.Sponsors),
		DateCreated:        cause.CreatedAt,
		DateUpdated:        cause.UpdatedAt,
	}
}

func (h *ChallengeHandler) causeRunnerToResponse(runner *challengeModel.CauseRunner, owner *userModel.User) dto.CauseRunnerResponse {
	return dto.CauseRunnerResponse{
		ID:              runner.ID,
		CauseID:         runner.CauseID,
		UserID:          runner.OwnerID,
		RunnerID:        runner.ID,
		Username:        owner.Username,
		DistanceToCover: runner.DistanceToCover,
		DistanceCovered: runner.DistanceCovered,
		Duration:        runner.Duration,
		MoneyRaised:     runner.MoneyRaised,
		CoverImage:      runner.CoverImage,
		Activity:        runner.Activity,
		DateJoined:      runner.DateJoined,
	}
}

func (h *ChallengeHandler) userToResponse(user *userModel.User) dto.UserResponse {
	return dto.UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		FullName: user.FirstName + " " + user.LastName,
	}
}

// Convert user to challenge owner info
func (h *ChallengeHandler) userToChallengeOwnerInfo(user *userModel.User) dto.ChallengeOwnerInfo {
	return dto.ChallengeOwnerInfo{
		ID:       user.ID,
		FullName: user.FirstName + " " + user.LastName,
		Username: user.Username,
	}
}

// Convert user to cause owner info
func (h *ChallengeHandler) userToCauseOwnerInfo(user *userModel.User) dto.CauseOwnerInfo {
	return dto.CauseOwnerInfo{
		ID:       user.ID,
		FullName: user.FirstName + " " + user.LastName,
		Username: user.Username,
	}
}

// Convert interface{} to ChallengeMemberInfo slice (placeholder implementation)
func (h *ChallengeHandler) convertToChallengeMemberInfo(members []interface{}) []dto.ChallengeMemberInfo {
	// In production, you would properly convert the interface{} slice to actual user objects
	// For now, return empty slice as the junction table queries are not implemented
	return []dto.ChallengeMemberInfo{}
}

// Convert interface{} to ChallengeSponsorInfo slice (placeholder implementation)
func (h *ChallengeHandler) convertToChallengeSponsorInfo(sponsors []interface{}) []dto.ChallengeSponsorInfo {
	// In production, you would properly convert the interface{} slice to actual user objects
	// For now, return empty slice as the junction table queries are not implemented
	return []dto.ChallengeSponsorInfo{}
}

// Convert interface{} to CauseMemberInfo slice (placeholder implementation)
func (h *ChallengeHandler) convertToCauseMemberInfo(members []interface{}) []dto.CauseMemberInfo {
	// In production, you would properly convert the interface{} slice to actual user objects
	// For now, return empty slice as the junction table queries are not implemented
	return []dto.CauseMemberInfo{}
}

// Convert interface{} to CauseSponsorInfo slice (placeholder implementation)
func (h *ChallengeHandler) convertToCauseSponsorInfo(sponsors []interface{}) []dto.CauseSponsorInfo {
	// In production, you would properly convert the interface{} slice to actual user objects
	// For now, return empty slice as the junction table queries are not implemented
	return []dto.CauseSponsorInfo{}
}
