package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gopi.com/api/http/dto"
	"gopi.com/internal/app/campaign"
	"gopi.com/internal/app/user"
	"gopi.com/internal/apperr"
	campaignModel "gopi.com/internal/domain/campaign/model"
	userModel "gopi.com/internal/domain/user/model"
)

type CampaignHandler struct {
	campaignService *campaign.CampaignService
	userService     *user.UserService
}

func NewCampaignHandler(campaignService *campaign.CampaignService, userService *user.UserService) *CampaignHandler {
	return &CampaignHandler{
		campaignService: campaignService,
		userService:     userService,
	}
}

// CreateCampaign godoc
// @Summary Create a new campaign
// @Description Create a new campaign with the provided details
// @Tags campaigns
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param campaign body dto.CreateCampaignRequest true "Campaign details"
// @Success 201 {object} dto.CampaignResponse "Campaign created successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request body"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized - not a member or sponsor"
// @Failure 404 {object} dto.ErrorResponse "Campaign not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /campaigns [post]
func (h *CampaignHandler) CreateCampaign(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		respondError(c, apperr.E("CreateCampaign", apperr.Unauthorized, nil, "User not authenticated"))
		return
	}

	var req dto.CreateCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperr.E("CreateCampaign", apperr.InvalidInput, err, "Invalid request body"))
		return
	}

	// Get user details for slug generation
	user, err := h.userService.GetUserByID(userID.(string))
	if err != nil {
		respondError(c, apperr.E("CreateCampaign", apperr.Internal, err, "Failed to get user details"))
		return
	}

	// Create campaign
	campaign, err := h.campaignService.CreateCampaign(
		userID.(string),
		user.Username,
		req.Name,
		req.Description,
		req.Condition,
		req.Goal,
		req.Location,
		campaignModel.CampaignMode(req.Mode),
		campaignModel.Activity(req.Activity),
		req.TargetAmount,
		req.TargetAmountPerKm,
		req.DistanceToCover,
		req.StartDuration,
		req.EndDuration,
	)
	if err != nil {
		respondError(c, apperr.E("CreateCampaign", apperr.Internal, err, "Failed to create campaign"))
		return
	}

	response := h.campaignToResponse(campaign, user)
	c.JSON(http.StatusCreated, response)
}

// GetCampaigns godoc
// @Summary List all campaigns
// @Description Get a paginated list of all campaigns
// @Tags campaigns
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(10)
// @Success 200 {object} dto.CampaignListResponse "Campaigns retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /campaigns [get]
func (h *CampaignHandler) GetCampaigns(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	
	offset := (page - 1) * limit

	campaigns, err := h.campaignService.ListCampaigns(limit, offset)
	if err != nil {
		respondError(c, apperr.E("GetCampaigns", apperr.Internal, err, "Failed to fetch campaigns"))
		return
	}

	var responses []dto.CampaignResponse
	for _, campaign := range campaigns {
		owner, _ := h.userService.GetUserByID(campaign.OwnerID)
		responses = append(responses, h.campaignToResponse(campaign, owner))
	}

	response := dto.CampaignListResponse{
		Campaigns: responses,
		Total:     len(responses),
		Page:      page,
		Limit:     limit,
	}

	c.JSON(http.StatusOK, response)
}

// GetCampaignBySlug godoc
// @Summary Get campaign by slug
// @Description Get a specific campaign by its slug
// @Tags campaigns
// @Accept json
// @Produce json
// @Param slug path string true "Campaign slug"
// @Success 200 {object} dto.CampaignResponse "Campaign retrieved successfully"
// @Failure 404 {object} dto.ErrorResponse "Campaign not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /campaigns/{slug} [get]
func (h *CampaignHandler) GetCampaignBySlug(c *gin.Context) {
	slug := c.Param("slug")

	campaign, err := h.campaignService.GetCampaignBySlug(slug)
	if err != nil {
		respondError(c, apperr.E("GetCampaignBySlug", apperr.NotFound, err, "Campaign not found"))
		return
	}

	owner, _ := h.userService.GetUserByID(campaign.OwnerID)
	response := h.campaignToResponse(campaign, owner)
	c.JSON(http.StatusOK, response)
}

// UpdateCampaign godoc
// @Summary Update campaign
// @Description Update a campaign by its slug (only owner can update)
// @Tags campaigns
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param slug path string true "Campaign slug"
// @Param campaign body dto.UpdateCampaignRequest true "Campaign update details"
// @Success 200 {object} dto.CampaignResponse "Campaign updated successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request body"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - not campaign owner"
// @Failure 404 {object} dto.ErrorResponse "Campaign not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /campaigns/{slug} [put]
func (h *CampaignHandler) UpdateCampaign(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		respondError(c, apperr.E("UpdateCampaign", apperr.Unauthorized, nil, "User not authenticated"))
		return
	}

	slug := c.Param("slug")
	campaign, err := h.campaignService.GetCampaignBySlug(slug)
	if err != nil {
		respondError(c, apperr.E("UpdateCampaign", apperr.NotFound, err, "Campaign not found"))
		return
	}

	// Check if user is the owner
	if campaign.OwnerID != userID.(string) {
		respondError(c, apperr.E("UpdateCampaign", apperr.Forbidden, nil, "You must be the owner of this campaign"))
		return
	}

	var req dto.UpdateCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperr.E("UpdateCampaign", apperr.InvalidInput, err, "Invalid request body"))
		return
	}

	// Update campaign fields
	if req.Name != "" {
		campaign.Name = req.Name
	}
	if req.Description != "" {
		campaign.Description = req.Description
	}
	if req.Condition != "" {
		campaign.Condition = req.Condition
	}
	if req.Mode != "" {
		campaign.Mode = campaignModel.CampaignMode(req.Mode)
	}
	if req.Goal != "" {
		campaign.Goal = req.Goal
	}
	if req.Activity != "" {
		campaign.Activity = campaignModel.Activity(req.Activity)
	}
	if req.Location != "" {
		campaign.Location = req.Location
	}
	if req.TargetAmount != nil {
		campaign.TargetAmount = *req.TargetAmount
	}
	if req.TargetAmountPerKm != nil {
		campaign.TargetAmountPerKm = *req.TargetAmountPerKm
	}
	if req.DistanceToCover != nil {
		campaign.DistanceToCover = *req.DistanceToCover
	}
	if req.StartDuration != "" {
		campaign.StartDuration = req.StartDuration
	}
	if req.EndDuration != "" {
		campaign.EndDuration = req.EndDuration
	}
	if req.WorkoutImg != "" {
		campaign.WorkoutImg = req.WorkoutImg
	}
	if req.AcceptTac != nil {
		campaign.AcceptTac = *req.AcceptTac
	}

	err = h.campaignService.UpdateCampaign(campaign)
	if err != nil {
		respondError(c, apperr.E("UpdateCampaign", apperr.Internal, err, "Failed to update campaign"))
		return
	}

	owner, _ := h.userService.GetUserByID(campaign.OwnerID)
	response := h.campaignToResponse(campaign, owner)
	c.JSON(http.StatusOK, response)
}

// DeleteCampaign godoc
// @Summary Delete campaign
// @Description Delete a campaign by its slug (only owner can delete)
// @Tags campaigns
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param slug path string true "Campaign slug"
// @Success 204 "Campaign deleted successfully"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - not campaign owner"
// @Failure 404 {object} dto.ErrorResponse "Campaign not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /campaigns/{slug} [delete]
func (h *CampaignHandler) DeleteCampaign(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		respondError(c, apperr.E("DeleteCampaign", apperr.Unauthorized, nil, "User not authenticated"))
		return
	}

	slug := c.Param("slug")
	campaign, err := h.campaignService.GetCampaignBySlug(slug)
	if err != nil {
		respondError(c, apperr.E("DeleteCampaign", apperr.NotFound, err, "Campaign not found"))
		return
	}

	// Check if user is the owner
	if campaign.OwnerID != userID.(string) {
		respondError(c, apperr.E("DeleteCampaign", apperr.Forbidden, nil, "You must be the owner of this campaign"))
		return
	}

	err = h.campaignService.DeleteCampaign(campaign.ID)
	if err != nil {
		respondError(c, apperr.E("DeleteCampaign", apperr.Internal, err, "Failed to delete campaign"))
		return
	}

	c.Status(http.StatusNoContent)
}

// GetCampaignsByUser godoc
// @Summary Get campaigns by current user
// @Description Get all campaigns created by the authenticated user
// @Tags campaigns
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} dto.CampaignListResponse "User campaigns retrieved successfully"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /campaigns/by_user [get]
func (h *CampaignHandler) GetCampaignsByUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		respondError(c, apperr.E("GetCampaignsByUser", apperr.Unauthorized, nil, "User not authenticated"))
		return
	}

	campaigns, err := h.campaignService.GetCampaignsByOwner(userID.(string))
	if err != nil {
		respondError(c, apperr.E("GetCampaignsByUser", apperr.Internal, err, "Failed to fetch user campaigns"))
		return
	}

	var responses []dto.CampaignResponse
	user, _ := h.userService.GetUserByID(userID.(string))
	for _, campaign := range campaigns {
		responses = append(responses, h.campaignToResponse(campaign, user))
	}

	response := dto.CampaignListResponse{
		Campaigns: responses,
		Total:     len(responses),
		Page:      1,
		Limit:     len(responses),
	}

	c.JSON(http.StatusOK, response)
}

// GetCampaignsByOthers godoc
// @Summary Get campaigns by other users
// @Description Get all campaigns created by users other than the authenticated user
// @Tags campaigns
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(10)
// @Success 200 {object} dto.CampaignListResponse "Other users' campaigns retrieved successfully"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /campaigns/by_others [get]
func (h *CampaignHandler) GetCampaignsByOthers(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		respondError(c, apperr.E("GetCampaignsByOthers", apperr.Unauthorized, nil, "User not authenticated"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	
	offset := (page - 1) * limit

	campaigns, err := h.campaignService.GetCampaignsByNonOwner(userID.(string), limit, offset)
	if err != nil {
		respondError(c, apperr.E("GetCampaignsByOthers", apperr.Internal, err, "Failed to fetch other users' campaigns"))
		return
	}

	var responses []dto.CampaignResponse
	for _, campaign := range campaigns {
		owner, _ := h.userService.GetUserByID(campaign.OwnerID)
		responses = append(responses, h.campaignToResponse(campaign, owner))
	}

	response := dto.CampaignListResponse{
		Campaigns: responses,
		Total:     len(responses),
		Page:      page,
		Limit:     limit,
	}

	c.JSON(http.StatusOK, response)
}

// JoinCampaign godoc
// @Summary Join a campaign
// @Description Add the authenticated user as a member of the campaign
// @Tags campaigns
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param slug path string true "Campaign slug"
// @Success 200 {object} dto.MessageResponse "Successfully joined campaign"
// @Failure 400 {object} dto.ErrorResponse "Already a member"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Campaign not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /campaigns/{slug}/join [put]
func (h *CampaignHandler) JoinCampaign(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		respondError(c, apperr.E("JoinCampaign", apperr.Unauthorized, nil, "User not authenticated"))
		return
	}

	slug := c.Param("slug")
	campaign, err := h.campaignService.GetCampaignBySlug(slug)
	if err != nil {
		respondError(c, apperr.E("JoinCampaign", apperr.NotFound, err, "Campaign not found"))
		return
	}

	// Check if user is already a member
	isMember, err := h.campaignService.IsMember(campaign.ID, userID.(string))
	if err != nil {
		respondError(c, apperr.E("JoinCampaign", apperr.Internal, err, "Failed to check membership"))
		return
	}

	if isMember {
		c.JSON(http.StatusBadRequest, dto.MessageResponse{
			Message: "You have already joined " + campaign.Name + " campaign",
			Success: false,
		})
		return
	}

	err = h.campaignService.AddMember(campaign.ID, userID.(string))
	if err != nil {
		respondError(c, apperr.E("JoinCampaign", apperr.Internal, err, "Failed to join campaign"))
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "You joined " + campaign.Name + " campaign!",
		Success: true,
	})
}

// ParticipateCampaign godoc
// @Summary Participate in a campaign
// @Description Start participating in a campaign by creating a runner entry
// @Tags campaigns
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param slug path string true "Campaign slug"
// @Param participation body dto.ParticipateCampaignRequest true "Participation details"
// @Success 200 {object} dto.ParticipateCampaignResponse "Successfully started participation"
// @Failure 400 {object} dto.ErrorResponse "Invalid request body"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Campaign not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /campaigns/{slug}/participate [post]
func (h *CampaignHandler) ParticipateCampaign(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		respondError(c, apperr.E("ParticipateCampaign", apperr.Unauthorized, nil, "User not authenticated"))
		return
	}

	slug := c.Param("slug")
	campaign, err := h.campaignService.GetCampaignBySlug(slug)
	if err != nil {
		respondError(c, apperr.E("ParticipateCampaign", apperr.NotFound, err, "Campaign not found"))
		return
	}

	var req dto.ParticipateCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperr.E("ParticipateCampaign", apperr.InvalidInput, err, "Invalid request body"))
		return
	}

	// Add user as member if not already
	isMember, _ := h.campaignService.IsMember(campaign.ID, userID.(string))
	if !isMember {
		err = h.campaignService.AddMember(campaign.ID, userID.(string))
		if err != nil {
			respondError(c, apperr.E("ParticipateCampaign", apperr.Internal, err, "Failed to add user as member"))
			return
		}
	}

	// Create campaign runner entry
	runner, err := h.campaignService.ParticipateCampaign(slug, userID.(string), req.Activity)
	if err != nil {
		respondError(c, apperr.E("ParticipateCampaign", apperr.Internal, err, "Failed to create runner"))
		return
	}

	// Create campaign post entry (simplified - would integrate with post service in real implementation)
	postID := "post-" + runner.ID // Simplified post ID generation

	response := dto.ParticipateCampaignResponse{
		Message:  req.Activity,
		Success:  true,
		PostID:   postID,
		RunnerID: runner.ID,
	}

	c.JSON(http.StatusOK, response)
}

// SponsorCampaign godoc
// @Summary Sponsor a campaign
// @Description Add sponsorship to a campaign
// @Tags campaigns
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param slug path string true "Campaign slug"
// @Param sponsorship body dto.SponsorCampaignRequest true "Sponsorship details"
// @Success 200 {object} dto.SponsorCampaignResponse "Successfully sponsored campaign"
// @Failure 400 {object} dto.ErrorResponse "Invalid request body"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Campaign not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /campaigns/{slug}/sponsor [post]
func (h *CampaignHandler) SponsorCampaign(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		respondError(c, apperr.E("SponsorCampaign", apperr.Unauthorized, nil, "User not authenticated"))
		return
	}

	slug := c.Param("slug")
	campaign, err := h.campaignService.GetCampaignBySlug(slug)
	if err != nil {
		respondError(c, apperr.E("SponsorCampaign", apperr.NotFound, err, "Campaign not found"))
		return
	}

	var req dto.SponsorCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperr.E("SponsorCampaign", apperr.InvalidInput, err, "Invalid request body"))
		return
	}

	// Calculate total amount
	totalAmount := req.Distance * req.AmountPerKm

	// Create sponsor campaign entry
	err = h.campaignService.SponsorCampaign(campaign.ID, []interface{}{userID.(string)}, req.Distance, req.AmountPerKm)
	if err != nil {
		respondError(c, apperr.E("SponsorCampaign", apperr.Internal, err, "Failed to create sponsorship"))
		return
	}

	// Add user as sponsor
	err = h.campaignService.AddSponsor(campaign.ID, userID.(string))
	if err != nil {
		respondError(c, apperr.E("SponsorCampaign", apperr.Internal, err, "Failed to add sponsor"))
		return
	}

	response := dto.SponsorCampaignResponse{
		ID:          campaign.ID + "-" + userID.(string), // Generate a composite ID
		Distance:    req.Distance,
		Campaign:    campaign.Name,
		Sponsor:     userID.(string),
		AmountPerKm: req.AmountPerKm,
		TotalAmount: totalAmount,
		BrandImg:    req.BrandImg,
		VideoUrl:    req.VideoUrl,
		DateCreated: campaign.CreatedAt,
		Paystack:    nil, // Payment integration would go here
	}

	c.JSON(http.StatusOK, response)
}

// GetCampaignLeaderboard godoc
// @Summary Get campaign leaderboard
// @Description Get leaderboard of campaign participants ordered by distance covered
// @Tags campaigns
// @Accept json
// @Produce json
// @Param slug path string true "Campaign slug"
// @Success 200 {object} dto.CampaignLeaderboardResponse "Leaderboard retrieved successfully"
// @Failure 404 {object} dto.ErrorResponse "Campaign not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /campaigns/{slug}/leaderboard [get]
func (h *CampaignHandler) GetCampaignLeaderboard(c *gin.Context) {
	slug := c.Param("slug")
	
	// Get campaign runners ordered by distance covered
	runners, err := h.campaignService.GetLeaderboard(slug)
	if err != nil {
		respondError(c, apperr.E("GetCampaignLeaderboard", apperr.Internal, err, "Failed to get leaderboard"))
		return
	}

	// Convert to leaderboard response
	var leaderboard []dto.CampaignLeaderboardEntry
	for _, runner := range runners {
		user, _ := h.userService.GetUserByID(runner.OwnerID)
		entry := dto.CampaignLeaderboardEntry{
			UserID:          runner.OwnerID,
			Username:        "",
			FullName:        "",
			DistanceCovered: runner.DistanceCovered,
			MoneyRaised:     runner.MoneyRaised,
			Duration:        runner.Duration,
			Activity:        runner.Activity,
			CoverImage:      runner.CoverImage,
		}
		
		if user != nil {
			entry.Username = user.Username
			entry.FullName = user.FirstName + " " + user.LastName
		}
		
		leaderboard = append(leaderboard, entry)
	}

	response := dto.CampaignLeaderboardResponse{
		CampaignSlug: slug,
		Leaderboard:  leaderboard,
	}

	c.JSON(http.StatusOK, response)
}

// Helper function to convert campaign model to response DTO
func (h *CampaignHandler) campaignToResponse(campaign *campaignModel.Campaign, owner *userModel.User) dto.CampaignResponse {
	var ownerInfo dto.CampaignOwnerInfo
	if owner != nil {
		ownerInfo = dto.CampaignOwnerInfo{
			ID:       owner.ID,
			FullName: owner.FirstName + " " + owner.LastName,
			Username: owner.Username,
		}
	}

	// Members and Sponsors are now []interface{} in domain model and DTO
	return dto.CampaignResponse{
		ID:                campaign.ID,
		Name:              campaign.Name,
		Description:       campaign.Description,
		Condition:         campaign.Condition,
		Mode:              string(campaign.Mode),
		Goal:              campaign.Goal,
		Activity:          string(campaign.Activity),
		AcceptTac:         campaign.AcceptTac,
		Location:          campaign.Location,
		MoneyRaised:       campaign.MoneyRaised,
		TargetAmount:      campaign.TargetAmount,
		TargetAmountPerKm: campaign.TargetAmountPerKm,
		DistanceToCover:   campaign.DistanceToCover,
		DistanceCovered:   campaign.DistanceCovered,
		StartDuration:     campaign.StartDuration,
		EndDuration:       campaign.EndDuration,
		Members:           campaign.Members,
		Sponsors:          campaign.Sponsors,
		Owner:             ownerInfo,
		Slug:              campaign.Slug,
		WorkoutImg:        campaign.WorkoutImg,
		DateCreated:       campaign.CreatedAt,
		DateUpdated:       campaign.UpdatedAt,
	}
}

// GetFinishCampaignDetails godoc
// @Summary Get campaign finish details
// @Description Get details for finishing a campaign run (simplified without post system)
// @Tags campaigns
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param slug path string true "Campaign slug"
// @Param runner_id path string true "Campaign runner ID"
// @Success 200 {object} dto.CampaignRunnerResponse "Campaign runner details retrieved successfully"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Campaign or runner not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /campaigns/{slug}/finish_campaign/{runner_id} [get]
func (h *CampaignHandler) GetFinishCampaignDetails(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		respondError(c, apperr.E("GetFinishCampaignDetails", apperr.Unauthorized, nil, "User not authenticated"))
		return
	}

	slug := c.Param("slug")
	runnerID := c.Param("runner_id")

	// Verify campaign exists
	campaign, err := h.campaignService.GetCampaignBySlug(slug)
	if err != nil {
		respondError(c, apperr.E("GetFinishCampaignDetails", apperr.NotFound, err, "Campaign not found"))
		return
	}

	// Get runner details
	runner, err := h.campaignService.GetRunnerByID(runnerID)
	if err != nil {
		respondError(c, apperr.E("GetFinishCampaignDetails", apperr.NotFound, err, "Campaign runner not found"))
		return
	}

	// Verify runner belongs to campaign and user
	if runner.CampaignID != campaign.ID || runner.OwnerID != userID.(string) {
		respondError(c, apperr.E("GetFinishCampaignDetails", apperr.Forbidden, nil, "Access denied to this runner"))
		return
	}

	user, _ := h.userService.GetUserByID(runner.OwnerID)
	response := h.campaignRunnerToResponse(runner, user)
	c.JSON(http.StatusOK, response)
}

// FinishCampaignRun godoc
// @Summary Finish a campaign run
// @Description Complete a campaign run by updating runner details
// @Tags campaigns
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param slug path string true "Campaign slug"
// @Param runner_id path string true "Campaign runner ID"
// @Param details body dto.FinishActivityRequest true "Activity completion details"
// @Success 200 {object} dto.CampaignRunnerResponse "Campaign run finished successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request body"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Campaign or runner not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /campaigns/{slug}/finish_campaign/{runner_id} [put]
func (h *CampaignHandler) FinishCampaignRun(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		respondError(c, apperr.E("FinishCampaignRun", apperr.Unauthorized, nil, "User not authenticated"))
		return
	}

	slug := c.Param("slug")
	runnerID := c.Param("runner_id")

	var req dto.FinishActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperr.E("FinishCampaignRun", apperr.InvalidInput, err, "Invalid request body"))
		return
	}

	// Verify campaign exists
	campaign, err := h.campaignService.GetCampaignBySlug(slug)
	if err != nil {
		respondError(c, apperr.E("FinishCampaignRun", apperr.NotFound, err, "Campaign not found"))
		return
	}

	// Get runner details
	runner, err := h.campaignService.GetRunnerByID(runnerID)
	if err != nil {
		respondError(c, apperr.E("FinishCampaignRun", apperr.NotFound, err, "Campaign runner not found"))
		return
	}

	// Verify runner belongs to campaign and user
	if runner.CampaignID != campaign.ID || runner.OwnerID != userID.(string) {
		respondError(c, apperr.E("FinishCampaignRun", apperr.Forbidden, nil, "Access denied to this runner"))
		return
	}

	// Update runner with completion details
	err = h.campaignService.FinishActivity(runnerID, req.DistanceCovered, req.Duration, req.MoneyRaised)
	if err != nil {
		respondError(c, apperr.E("FinishCampaignRun", apperr.Internal, err, "Failed to finish campaign run"))
		return
	}

	// Get updated runner
	updatedRunner, err := h.campaignService.GetRunnerByID(runnerID)
	if err != nil {
		respondError(c, apperr.E("FinishCampaignRun", apperr.Internal, err, "Failed to get updated runner"))
		return
	}

	user, _ := h.userService.GetUserByID(updatedRunner.OwnerID)
	response := h.campaignRunnerToResponse(updatedRunner, user)
	c.JSON(http.StatusOK, response)
}

// Helper function to convert campaign runner to response DTO
func (h *CampaignHandler) campaignRunnerToResponse(runner *campaignModel.CampaignRunner, user *userModel.User) dto.CampaignRunnerResponse {
	var username string
	if user != nil {
		username = user.Username
	}

	return dto.CampaignRunnerResponse{
		ID:              runner.ID,
		UserID:          runner.OwnerID,
		RunnerID:        runner.ID,
		Username:        username,
		DistanceCovered: runner.DistanceCovered,
		Duration:        runner.Duration,
		MoneyRaised:     runner.MoneyRaised,
		CampaignID:      runner.CampaignID,
		CoverImage:      runner.CoverImage,
		Activity:        runner.Activity,
		DateJoined:      runner.CreatedAt,
	}
}
