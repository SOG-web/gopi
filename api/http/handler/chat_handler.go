package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gopi.com/api/http/dto"
	"gopi.com/internal/app/chat"
	"gopi.com/internal/app/user"
	"gopi.com/internal/apperr"
	chatModel "gopi.com/internal/domain/chat/model"
)

type ChatHandler struct {
	chatService *chat.ChatService
	userService *user.UserService
}

func NewChatHandler(chatService *chat.ChatService, userService *user.UserService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
		userService: userService,
	}
}

// CreateGroup godoc
// @Summary Create a new chat group
// @Description Create a new chat group with the provided details
// @Tags chat
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param group body dto.CreateGroupRequest true "Group details"
// @Success 201 {object} dto.GroupResponse "Group created successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request body"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /chat/groups [post]
func (h *ChatHandler) CreateGroup(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		respondError(c, apperr.E("CreateGroup", apperr.Unauthorized, nil, "User not authenticated"))
		return
	}

	var req dto.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperr.E("CreateGroup", apperr.InvalidInput, err, "Invalid request body"))
		return
	}

	// Create group
	group, err := h.chatService.CreateGroup(userID.(string), req.Name, req.Image, req.MemberIDs)
	if err != nil {
		respondError(c, apperr.E("CreateGroup", apperr.Internal, err, "Failed to create group"))
		return
	}

	response := h.groupToResponse(group)
	c.JSON(http.StatusCreated, response)
}

// GetGroups godoc
// @Summary List all groups for current user
// @Description Get a paginated list of all groups where the user is a member or creator
// @Tags chat
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(10)
// @Success 200 {object} dto.GroupListResponse "Groups retrieved successfully"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /chat/groups [get]
func (h *ChatHandler) GetGroups(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		respondError(c, apperr.E("GetGroups", apperr.Unauthorized, nil, "User not authenticated"))
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

	// Get groups where user is a member
	groups, err := h.chatService.GetGroupsByMember(userID.(string))
	if err != nil {
		respondError(c, apperr.E("GetGroups", apperr.Internal, err, "Failed to fetch groups"))
		return
	}

	// Apply pagination
	offset := (page - 1) * limit
	var paginatedGroups []*chatModel.Group
	total := len(groups)

	if offset < total {
		end := offset + limit
		if end > total {
			end = total
		}
		paginatedGroups = groups[offset:end]
	}

	var responses []dto.GroupResponse
	for _, group := range paginatedGroups {
		responses = append(responses, h.groupToResponse(group))
	}

	response := dto.GroupListResponse{
		Groups: responses,
		Total:  total,
		Page:   page,
		Limit:  limit,
	}

	c.JSON(http.StatusOK, response)
}

// GetGroupBySlug godoc
// @Summary Get group by slug
// @Description Get a specific group by its slug
// @Tags chat
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param slug path string true "Group slug"
// @Success 200 {object} dto.GroupResponse "Group retrieved successfully"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - not a member"
// @Failure 404 {object} dto.ErrorResponse "Group not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /chat/groups/{slug} [get]
func (h *ChatHandler) GetGroupBySlug(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		respondError(c, apperr.E("GetGroupBySlug", apperr.Unauthorized, nil, "User not authenticated"))
		return
	}

	slug := c.Param("slug")

	group, err := h.chatService.GetGroupBySlug(slug)
	if err != nil {
		respondError(c, apperr.E("GetGroupBySlug", apperr.NotFound, err, "Group not found"))
		return
	}

	// Check if user is a member of the group
	isMember := false
	for _, memberID := range group.MemberIDs {
		if memberID == userID.(string) {
			isMember = true
			break
		}
	}

	if !isMember && group.CreatorID != userID.(string) {
		respondError(c, apperr.E("GetGroupBySlug", apperr.Forbidden, nil, "You must be a member of this group"))
		return
	}

	response := h.groupToResponse(group)
	c.JSON(http.StatusOK, response)
}

// UpdateGroup godoc
// @Summary Update group
// @Description Update a group by its slug (only creator can update)
// @Tags chat
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param slug path string true "Group slug"
// @Param group body dto.UpdateGroupRequest true "Group update details"
// @Success 200 {object} dto.GroupResponse "Group updated successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request body"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - not group creator"
// @Failure 404 {object} dto.ErrorResponse "Group not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /chat/groups/{slug} [put]
func (h *ChatHandler) UpdateGroup(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		respondError(c, apperr.E("UpdateGroup", apperr.Unauthorized, nil, "User not authenticated"))
		return
	}

	slug := c.Param("slug")
	group, err := h.chatService.GetGroupBySlug(slug)
	if err != nil {
		respondError(c, apperr.E("UpdateGroup", apperr.NotFound, err, "Group not found"))
		return
	}

	// Check if user is the creator
	if group.CreatorID != userID.(string) {
		respondError(c, apperr.E("UpdateGroup", apperr.Forbidden, nil, "You must be the group creator"))
		return
	}

	var req dto.UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperr.E("UpdateGroup", apperr.InvalidInput, err, "Invalid request body"))
		return
	}

	// Update group fields
	if req.Name != "" {
		group.Name = req.Name
	}
	if req.Image != "" {
		group.Image = req.Image
	}

	err = h.chatService.UpdateGroup(group)
	if err != nil {
		respondError(c, apperr.E("UpdateGroup", apperr.Internal, err, "Failed to update group"))
		return
	}

	response := h.groupToResponse(group)
	c.JSON(http.StatusOK, response)
}

// DeleteGroup godoc
// @Summary Delete group
// @Description Delete a group by its slug (only creator can delete)
// @Tags chat
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param slug path string true "Group slug"
// @Success 204 "Group deleted successfully"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - not group creator"
// @Failure 404 {object} dto.ErrorResponse "Group not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /chat/groups/{slug} [delete]
func (h *ChatHandler) DeleteGroup(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		respondError(c, apperr.E("DeleteGroup", apperr.Unauthorized, nil, "User not authenticated"))
		return
	}

	slug := c.Param("slug")
	group, err := h.chatService.GetGroupBySlug(slug)
	if err != nil {
		respondError(c, apperr.E("DeleteGroup", apperr.NotFound, err, "Group not found"))
		return
	}

	// Check if user is the creator
	if group.CreatorID != userID.(string) {
		respondError(c, apperr.E("DeleteGroup", apperr.Forbidden, nil, "You must be the group creator"))
		return
	}

	err = h.chatService.DeleteGroup(group.ID, userID.(string))
	if err != nil {
		respondError(c, apperr.E("DeleteGroup", apperr.Internal, err, "Failed to delete group"))
		return
	}

	c.Status(http.StatusNoContent)
}

// JoinGroup godoc
// @Summary Join a chat group
// @Description Add the authenticated user as a member of the group
// @Tags chat
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param slug path string true "Group slug"
// @Success 200 {object} dto.GroupMemberResponse "Successfully joined group"
// @Failure 400 {object} dto.ErrorResponse "Already a member"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Group not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /chat/groups/{slug}/join [post]
func (h *ChatHandler) JoinGroup(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		respondError(c, apperr.E("JoinGroup", apperr.Unauthorized, nil, "User not authenticated"))
		return
	}

	slug := c.Param("slug")
	group, err := h.chatService.GetGroupBySlug(slug)
	if err != nil {
		respondError(c, apperr.E("JoinGroup", apperr.NotFound, err, "Group not found"))
		return
	}

	// Check if user is already a member
	userIDStr := userID.(string)
	for _, memberID := range group.MemberIDs {
		if memberID == userIDStr {
			c.JSON(http.StatusBadRequest, dto.GroupMemberResponse{
				Message: "You are already a member of " + group.Name + " group",
				Success: false,
			})
			return
		}
	}

	err = h.chatService.AddMemberToGroup(group.ID, userIDStr, userIDStr)
	if err != nil {
		respondError(c, apperr.E("JoinGroup", apperr.Internal, err, "Failed to join group"))
		return
	}

	c.JSON(http.StatusOK, dto.GroupMemberResponse{
		Message: "You joined " + group.Name + " group!",
		Success: true,
	})
}

// LeaveGroup godoc
// @Summary Leave a chat group
// @Description Remove the authenticated user from the group
// @Tags chat
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param slug path string true "Group slug"
// @Success 200 {object} dto.GroupMemberResponse "Successfully left group"
// @Failure 400 {object} dto.ErrorResponse "Not a member"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Group not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /chat/groups/{slug}/leave [post]
func (h *ChatHandler) LeaveGroup(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		respondError(c, apperr.E("LeaveGroup", apperr.Unauthorized, nil, "User not authenticated"))
		return
	}

	slug := c.Param("slug")
	group, err := h.chatService.GetGroupBySlug(slug)
	if err != nil {
		respondError(c, apperr.E("LeaveGroup", apperr.NotFound, err, "Group not found"))
		return
	}

	// Check if user is a member
	isMember := false
	for _, memberID := range group.MemberIDs {
		if memberID == userID.(string) {
			isMember = true
			break
		}
	}

	if !isMember {
		c.JSON(http.StatusBadRequest, dto.GroupMemberResponse{
			Message: "You are not a member of " + group.Name + " group",
			Success: false,
		})
		return
	}

	err = h.chatService.RemoveMemberFromGroup(group.ID, userID.(string), userID.(string))
	if err != nil {
		respondError(c, apperr.E("LeaveGroup", apperr.Internal, err, "Failed to leave group"))
		return
	}

	c.JSON(http.StatusOK, dto.GroupMemberResponse{
		Message: "You left " + group.Name + " group!",
		Success: true,
	})
}

// Helper function to convert group model to response DTO
func (h *ChatHandler) groupToResponse(group *chatModel.Group) dto.GroupResponse {
	return dto.GroupResponse{
		ID:        group.ID,
		Name:      group.Name,
		MemberIDs: group.MemberIDs,
		CreatorID: group.CreatorID,
		Slug:      group.Slug,
		Image:     group.Image,
		CreatedAt: group.CreatedAt,
		UpdatedAt: group.UpdatedAt,
	}
}
