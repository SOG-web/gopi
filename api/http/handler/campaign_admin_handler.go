package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gopi.com/api/http/dto"
	"gopi.com/internal/app/campaign"
	"gopi.com/internal/app/user"
	"gopi.com/internal/apperr"
	campaignModel "gopi.com/internal/domain/campaign/model"
	userModel "gopi.com/internal/domain/user/model"
)

type CampaignAdminHandler struct {
	campaignService *campaign.CampaignService
	userService     *user.UserService
}

func NewCampaignAdminHandler(campaignService *campaign.CampaignService, userService *user.UserService) *CampaignAdminHandler {
	return &CampaignAdminHandler{
		campaignService: campaignService,
		userService:     userService,
	}
}

// CAMPAIGN RUNNER ADMIN ENDPOINTS

// CreateCampaignRunner godoc
// @Summary Create a new campaign runner (Admin)
// @Description Create a new campaign runner entry for admin purposes
// @Tags admin-campaign-runners
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param runner body dto.CreateCampaignRunnerRequest true "Campaign runner details"
// @Success 201 {object} dto.CampaignRunnerResponse "Campaign runner created successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request body"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Campaign or user not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /admin/campaign-runners [post]
func (h *CampaignAdminHandler) CreateCampaignRunner(c *gin.Context) {
	var req dto.CreateCampaignRunnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperr.E("CreateCampaignRunner", apperr.InvalidInput, err, "Invalid request body"))
		return
	}

	// Verify campaign exists
	campaign, err := h.campaignService.GetCampaignByID(req.CampaignID)
	if err != nil {
		respondError(c, apperr.E("CreateCampaignRunner", apperr.NotFound, err, "Campaign not found"))
		return
	}

	// Verify user exists
	user, err := h.userService.GetUserByID(req.UserID)
	if err != nil {
		respondError(c, apperr.E("CreateCampaignRunner", apperr.NotFound, err, "User not found"))
		return
	}

	// Create campaign runner via service
	runner, err := h.campaignService.ParticipateCampaign(campaign.Slug, req.UserID, req.Activity)
	if err != nil {
		respondError(c, apperr.E("CreateCampaignRunner", apperr.Internal, err, "Failed to create campaign runner"))
		return
	}

	// Update runner with additional details if provided
	if req.DistanceCovered > 0 || req.MoneyRaised > 0 || req.Duration != "" {
		err = h.campaignService.FinishActivity(runner.ID, req.DistanceCovered, req.Duration, req.MoneyRaised)
		if err != nil {
			respondError(c, apperr.E("CreateCampaignRunner", apperr.Internal, err, "Failed to update campaign runner"))
			return
		}
	}

	response := h.campaignRunnerToResponse(runner, user)
	c.JSON(http.StatusCreated, response)
}

// GetCampaignRunners godoc
// @Summary Get all campaign runners (Admin)
// @Description Get a paginated list of all campaign runners
// @Tags admin-campaign-runners
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(10)
// @Param campaign_id query string false "Filter by campaign ID"
// @Param user_id query string false "Filter by user ID"
// @Success 200 {object} dto.CampaignRunnerListResponse "Campaign runners retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid query parameters"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /admin/campaign-runners [get]
func (h *CampaignAdminHandler) GetCampaignRunners(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	campaignID := c.Query("campaign_id")
	userID := c.Query("user_id")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	var runners []*campaignModel.CampaignRunner
	var err error

	// Implement filtering logic
	if campaignID != "" {
		// Get runners by campaign - first get campaign by ID to get slug
		campaign, err := h.campaignService.GetCampaignByID(campaignID)
		if err != nil {
			respondError(c, apperr.E("GetCampaignRunners", apperr.NotFound, err, "Campaign not found"))
			return
		}
		runners, err = h.campaignService.GetLeaderboard(campaign.Slug)
		if err != nil {
			respondError(c, apperr.E("GetCampaignRunners", apperr.Internal, err, "Failed to fetch campaign runners"))
			return
		}
	} else if userID != "" {
		runners, err = h.campaignService.GetRunnersByUser(userID)
		if err != nil {
			respondError(c, apperr.E("GetCampaignRunners", apperr.Internal, err, "Failed to fetch campaign runners"))
			return
		}
	} else {
		// Get all runners - we'll need to implement this in service
		// For now, get a representative sample by getting runners from recent campaigns
		campaigns, campErr := h.campaignService.ListCampaigns(10, 0)
		if campErr != nil {
			respondError(c, apperr.E("GetCampaignRunners", apperr.Internal, campErr, "Failed to fetch campaigns"))
			return
		}

		// Get runners from all recent campaigns
		for _, campaign := range campaigns {
			campaignRunners, _ := h.campaignService.GetLeaderboard(campaign.Slug)
			runners = append(runners, campaignRunners...)
		}
	}

	if err != nil {
		respondError(c, apperr.E("GetCampaignRunners", apperr.Internal, err, "Failed to fetch campaign runners"))
		return
	}

	var responses []dto.CampaignRunnerResponse
	for _, runner := range runners {
		user, _ := h.userService.GetUserByID(runner.OwnerID)
		responses = append(responses, h.campaignRunnerToResponse(runner, user))
	}

	response := dto.CampaignRunnerListResponse{
		Runners: responses,
		Total:   len(responses),
		Page:    page,
		Limit:   limit,
	}

	c.JSON(http.StatusOK, response)
}

// GetCampaignRunnerByID godoc
// @Summary Get campaign runner by ID (Admin)
// @Description Get a specific campaign runner by its ID
// @Tags admin-campaign-runners
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Campaign runner ID"
// @Success 200 {object} dto.CampaignRunnerResponse "Campaign runner retrieved successfully"
// @Failure 404 {object} dto.ErrorResponse "Campaign runner not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /admin/campaign-runners/{id} [get]
func (h *CampaignAdminHandler) GetCampaignRunnerByID(c *gin.Context) {
	id := c.Param("id")

	// Implement GetRunnerByID in service
	runner, err := h.campaignService.GetRunnerByID(id)
	if err != nil {
		// Check if it's a not found error
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "no rows") {
			respondError(c, apperr.E("GetCampaignRunnerByID", apperr.NotFound, err, "Campaign runner not found"))
		} else {
			respondError(c, apperr.E("GetCampaignRunnerByID", apperr.Internal, err, "Failed to get campaign runner"))
		}
		return
	}

	user, _ := h.userService.GetUserByID(runner.OwnerID)
	response := h.campaignRunnerToResponse(runner, user)

	c.JSON(http.StatusOK, response)
}

// UpdateCampaignRunner godoc
// @Summary Update campaign runner (Admin)
// @Description Update a campaign runner by its ID
// @Tags admin-campaign-runners
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Campaign runner ID"
// @Param runner body dto.UpdateCampaignRunnerRequest true "Campaign runner update details"
// @Success 200 {object} dto.CampaignRunnerResponse "Campaign runner updated successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request body"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Campaign runner not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /admin/campaign-runners/{id} [put]
func (h *CampaignAdminHandler) UpdateCampaignRunner(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateCampaignRunnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperr.E("UpdateCampaignRunner", apperr.InvalidInput, err, "Invalid request body"))
		return
	}

	// Get existing runner
	runner, err := h.campaignService.GetRunnerByID(id)
	if err != nil {
		respondError(c, apperr.E("UpdateCampaignRunner", apperr.NotFound, err, "Campaign runner not found"))
		return
	}

	// Update fields if provided
	if req.Activity != "" {
		runner.Activity = req.Activity
	}
	if req.DistanceCovered != nil {
		runner.DistanceCovered = *req.DistanceCovered
	}
	if req.Duration != "" {
		runner.Duration = req.Duration
	}
	if req.MoneyRaised != nil {
		runner.MoneyRaised = *req.MoneyRaised
	}
	if req.CoverImage != "" {
		runner.CoverImage = req.CoverImage
	}

	// Update the runner
	err = h.campaignService.UpdateRunner(runner)
	if err != nil {
		respondError(c, apperr.E("UpdateCampaignRunner", apperr.Internal, err, "Failed to update campaign runner"))
		return
	}

	user, _ := h.userService.GetUserByID(runner.OwnerID)
	response := h.campaignRunnerToResponse(runner, user)
	c.JSON(http.StatusOK, response)
}

// DeleteCampaignRunner godoc
// @Summary Delete campaign runner (Admin)
// @Description Delete a campaign runner by its ID
// @Tags admin-campaign-runners
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Campaign runner ID"
// @Success 204 "Campaign runner deleted successfully"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Campaign runner not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /admin/campaign-runners/{id} [delete]
func (h *CampaignAdminHandler) DeleteCampaignRunner(c *gin.Context) {
	id := c.Param("id")

	// Check if runner exists first
	_, err := h.campaignService.GetRunnerByID(id)
	if err != nil {
		respondError(c, apperr.E("DeleteCampaignRunner", apperr.NotFound, err, "Campaign runner not found"))
		return
	}

	// Delete the runner
	err = h.campaignService.DeleteRunner(id)
	if err != nil {
		respondError(c, apperr.E("DeleteCampaignRunner", apperr.Internal, err, "Failed to delete campaign runner"))
		return
	}

	c.Status(http.StatusNoContent)
}

// SPONSOR CAMPAIGN ADMIN ENDPOINTS

// CreateSponsorCampaign godoc
// @Summary Create a new sponsor campaign (Admin)
// @Description Create a new sponsor campaign entry for admin purposes
// @Tags admin-sponsor-campaigns
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param sponsor body dto.CreateSponsorCampaignRequest true "Sponsor campaign details"
// @Success 201 {object} dto.SponsorCampaignResponse "Sponsor campaign created successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request body"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Campaign not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /admin/sponsor-campaigns [post]
func (h *CampaignAdminHandler) CreateSponsorCampaign(c *gin.Context) {
	var req dto.CreateSponsorCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperr.E("CreateSponsorCampaign", apperr.InvalidInput, err, "Invalid request body"))
		return
	}

	// Verify campaign exists
	campaign, err := h.campaignService.GetCampaignByID(req.CampaignID)
	if err != nil {
		respondError(c, apperr.E("CreateSponsorCampaign", apperr.NotFound, err, "Campaign not found"))
		return
	}

	// Verify sponsor user exists
	_, err = h.userService.GetUserByID(req.SponsorID)
	if err != nil {
		respondError(c, apperr.E("CreateSponsorCampaign", apperr.NotFound, err, "Sponsor user not found"))
		return
	}

	// Prepare sponsors as []interface{} (for compatibility with new domain model)
	var sponsors []interface{}
	if len(req.SponsorIDs) > 0 {
		for _, id := range req.SponsorIDs {
			sponsors = append(sponsors, id)
		}
	} else if req.SponsorID != "" {
		sponsors = append(sponsors, req.SponsorID)
	}

	// Implement sponsor campaign creation
	sponsor, err := h.campaignService.CreateSponsorCampaign(
		req.CampaignID,
		sponsors,
		req.Distance,
		req.AmountPerKm,
		req.BrandImg,
		req.VideoUrl,
	)
	if err != nil {
		respondError(c, apperr.E("CreateSponsorCampaign", apperr.Internal, err, "Failed to create sponsor campaign"))
		return
	}

	// Add sponsor to campaign sponsors
	err = h.campaignService.AddSponsor(req.CampaignID, req.SponsorID)
	if err != nil {
		respondError(c, apperr.E("CreateSponsorCampaign", apperr.Internal, err, "Failed to add sponsor to campaign"))
		return
	}

	response := h.sponsorCampaignToResponse(sponsor)
	response.Campaign = campaign.Name
	response.Sponsor = req.SponsorID

	c.JSON(http.StatusCreated, response)
}

// GetSponsorCampaigns godoc
// @Summary Get all sponsor campaigns (Admin)
// @Description Get a paginated list of all sponsor campaigns
// @Tags admin-sponsor-campaigns
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(10)
// @Param campaign_id query string false "Filter by campaign ID"
// @Success 200 {object} dto.SponsorCampaignListResponse "Sponsor campaigns retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid query parameters"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /admin/sponsor-campaigns [get]
func (h *CampaignAdminHandler) GetSponsorCampaigns(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	campaignID := c.Query("campaign_id")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Implement fetching logic
	var sponsors []*campaignModel.SponsorCampaign
	var err error

	if campaignID != "" {
		sponsors, err = h.campaignService.GetSponsorCampaignsByCampaign(campaignID)
	} else {
		// Get sponsors from all recent campaigns
		campaigns, campErr := h.campaignService.ListCampaigns(10, 0)
		if campErr != nil {
			respondError(c, apperr.E("GetSponsorCampaigns", apperr.Internal, campErr, "Failed to fetch campaigns"))
			return
		}

		// Get sponsors from all recent campaigns
		for _, campaign := range campaigns {
			campaignSponsors, _ := h.campaignService.GetSponsorCampaignsByCampaign(campaign.ID)
			sponsors = append(sponsors, campaignSponsors...)
		}
	}

	if err != nil {
		respondError(c, apperr.E("GetSponsorCampaigns", apperr.Internal, err, "Failed to fetch sponsor campaigns"))
		return
	}

	var responses []dto.SponsorCampaignResponse
	for _, sponsor := range sponsors {
		responses = append(responses, h.sponsorCampaignToResponse(sponsor))
	}

	response := dto.SponsorCampaignListResponse{
		Sponsors: responses,
		Total:    len(responses),
		Page:     page,
		Limit:    limit,
	}

	c.JSON(http.StatusOK, response)
}

// GetSponsorCampaignByID godoc
// @Summary Get sponsor campaign by ID (Admin)
// @Description Get a specific sponsor campaign by its ID
// @Tags admin-sponsor-campaigns
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Sponsor campaign ID"
// @Success 200 {object} dto.SponsorCampaignResponse "Sponsor campaign retrieved successfully"
// @Failure 404 {object} dto.ErrorResponse "Sponsor campaign not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /admin/sponsor-campaigns/{id} [get]
func (h *CampaignAdminHandler) GetSponsorCampaignByID(c *gin.Context) {
	id := c.Param("id")

	sponsor, err := h.campaignService.GetSponsorCampaignByID(id)
	if err != nil {
		// Check if it's a not found error
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "no rows") {
			respondError(c, apperr.E("GetSponsorCampaignByID", apperr.NotFound, err, "Sponsor campaign not found"))
		} else {
			respondError(c, apperr.E("GetSponsorCampaignByID", apperr.Internal, err, "Failed to get sponsor campaign"))
		}
		return
	}

	response := h.sponsorCampaignToResponse(sponsor)

	// Get campaign and sponsor names
	campaign, _ := h.campaignService.GetCampaignByID(sponsor.CampaignID)
	if campaign != nil {
		response.Campaign = campaign.Name
	}

	if len(sponsor.Sponsors) > 0 {
		if id, ok := sponsor.Sponsors[0].(string); ok {
			response.Sponsor = id // Use first sponsor ID
		} else {
			response.Sponsor = "" // fallback if not string
		}
	}

	c.JSON(http.StatusOK, response)
}

// UpdateSponsorCampaign godoc
// @Summary Update sponsor campaign (Admin)
// @Description Update a sponsor campaign by its ID
// @Tags admin-sponsor-campaigns
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Sponsor campaign ID"
// @Param sponsor body dto.UpdateSponsorCampaignRequest true "Sponsor campaign update details"
// @Success 200 {object} dto.SponsorCampaignResponse "Sponsor campaign updated successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request body"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Sponsor campaign not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /admin/sponsor-campaigns/{id} [put]
func (h *CampaignAdminHandler) UpdateSponsorCampaign(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateSponsorCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperr.E("UpdateSponsorCampaign", apperr.InvalidInput, err, "Invalid request body"))
		return
	}

	// Get existing sponsor campaign
	sponsor, err := h.campaignService.GetSponsorCampaignByID(id)
	if err != nil {
		respondError(c, apperr.E("UpdateSponsorCampaign", apperr.NotFound, err, "Sponsor campaign not found"))
		return
	}

	// Update fields if provided
	if req.Distance != nil {
		sponsor.Distance = *req.Distance
	}
	if req.AmountPerKm != nil {
		sponsor.AmountPerKm = *req.AmountPerKm
	}
	if req.BrandImg != "" {
		sponsor.BrandImg = req.BrandImg
	}
	if req.VideoUrl != "" {
		sponsor.VideoUrl = req.VideoUrl
	}

	// Recalculate total amount
	sponsor.CalculateTotalAmount()

	// Update the sponsor campaign
	err = h.campaignService.UpdateSponsorCampaign(sponsor)
	if err != nil {
		respondError(c, apperr.E("UpdateSponsorCampaign", apperr.Internal, err, "Failed to update sponsor campaign"))
		return
	}

	response := h.sponsorCampaignToResponse(sponsor)
	c.JSON(http.StatusOK, response)
}

// DeleteSponsorCampaign godoc
// @Summary Delete sponsor campaign (Admin)
// @Description Delete a sponsor campaign by its ID
// @Tags admin-sponsor-campaigns
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Sponsor campaign ID"
// @Success 204 "Sponsor campaign deleted successfully"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Sponsor campaign not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /admin/sponsor-campaigns/{id} [delete]
func (h *CampaignAdminHandler) DeleteSponsorCampaign(c *gin.Context) {
	id := c.Param("id")

	// Check if sponsor campaign exists first
	_, err := h.campaignService.GetSponsorCampaignByID(id)
	if err != nil {
		respondError(c, apperr.E("DeleteSponsorCampaign", apperr.NotFound, err, "Sponsor campaign not found"))
		return
	}

	// Delete the sponsor campaign
	err = h.campaignService.DeleteSponsorCampaign(id)
	if err != nil {
		respondError(c, apperr.E("DeleteSponsorCampaign", apperr.Internal, err, "Failed to delete sponsor campaign"))
		return
	}

	c.Status(http.StatusNoContent)
}

// Helper functions to convert models to response DTOs
func (h *CampaignAdminHandler) campaignRunnerToResponse(runner *campaignModel.CampaignRunner, user *userModel.User) dto.CampaignRunnerResponse {
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

func (h *CampaignAdminHandler) sponsorCampaignToResponse(sponsor *campaignModel.SponsorCampaign) dto.SponsorCampaignResponse {
	return dto.SponsorCampaignResponse{
		ID:          sponsor.ID,
		Distance:    sponsor.Distance,
		Campaign:    sponsor.CampaignID, // Will be overridden with campaign name in calling function
		Sponsor:     "",                 // Will be set in calling function
		AmountPerKm: sponsor.AmountPerKm,
		TotalAmount: sponsor.TotalAmount,
		BrandImg:    sponsor.BrandImg,
		VideoUrl:    sponsor.VideoUrl,
		DateCreated: sponsor.CreatedAt,
		Paystack:    nil, // Payment integration would go here
	}
}
