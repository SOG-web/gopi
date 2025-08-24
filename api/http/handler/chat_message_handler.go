package handler

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "gopi.com/api/http/dto"
    "gopi.com/internal/apperr"
    chatModel "gopi.com/internal/domain/chat/model"
)

// This file contains message-related handlers split out to keep files under 500 lines.

// SendMessage godoc
// @Summary Send a message to a group
// @Description Send a message to a specific group (only members can send)
// @Tags chat
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param slug path string true "Group slug"
// @Param message body dto.SendMessageRequest true "Message details"
// @Success 201 {object} dto.ChatMessageResponse "Message sent successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request body"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - not a member"
// @Failure 404 {object} dto.ErrorResponse "Group not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /chat/groups/{slug}/messages [post]
func (h *ChatHandler) SendMessage(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        respondError(c, apperr.E("SendMessage", apperr.Unauthorized, nil, "User not authenticated"))
        return
    }

    slug := c.Param("slug")
    group, err := h.chatService.GetGroupBySlug(slug)
    if err != nil {
        respondError(c, apperr.E("SendMessage", apperr.NotFound, err, "Group not found"))
        return
    }

    var req dto.SendMessageRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        respondError(c, apperr.E("SendMessage", apperr.InvalidInput, err, "Invalid request body"))
        return
    }

    message, err := h.chatService.SendMessage(userID.(string), group.ID, req.Content)
    if err != nil {
        respondError(c, apperr.E("SendMessage", apperr.Internal, err, "Failed to send message"))
        return
    }

    response := h.messageToResponse(message)
    c.JSON(http.StatusCreated, response)
}

// GetMessages godoc
// @Summary Get messages from a group
// @Description Get a paginated list of messages from a specific group
// @Tags chat
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param slug path string true "Group slug"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(20)
// @Success 200 {object} dto.MessageListResponse "Messages retrieved successfully"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - not a member"
// @Failure 404 {object} dto.ErrorResponse "Group not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /chat/groups/{slug}/messages [get]
func (h *ChatHandler) GetMessages(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        respondError(c, apperr.E("GetMessages", apperr.Unauthorized, nil, "User not authenticated"))
        return
    }

    slug := c.Param("slug")
    group, err := h.chatService.GetGroupBySlug(slug)
    if err != nil {
        respondError(c, apperr.E("GetMessages", apperr.NotFound, err, "Group not found"))
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
        respondError(c, apperr.E("GetMessages", apperr.Forbidden, nil, "You must be a member of this group"))
        return
    }

    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

    if page < 1 {
        page = 1
    }
    if limit < 1 || limit > 100 {
        limit = 20
    }

    offset := (page - 1) * limit

    messages, err := h.chatService.GetMessagesByGroup(group.ID, limit, offset)
    if err != nil {
        respondError(c, apperr.E("GetMessages", apperr.Internal, err, "Failed to fetch messages"))
        return
    }

    // Cache sender image URLs to avoid N+1 lookups
    senderImageCache := make(map[string]string, len(messages))
    responses := make([]dto.ChatMessageResponse, 0, len(messages))
    for _, message := range messages {
        img, ok := senderImageCache[message.SenderID]
        if !ok {
            if u, err := h.userService.GetUserByID(message.SenderID); err == nil && u != nil {
                img = u.ProfileImageURL
            }
            senderImageCache[message.SenderID] = img
        }
        responses = append(responses, dto.ChatMessageResponse{
            ID:             message.ID,
            SenderID:       message.SenderID,
            SenderImageURL: img,
            Content:        message.Content,
            GroupID:        message.GroupID,
            CreatedAt:      message.CreatedAt,
            UpdatedAt:      message.UpdatedAt,
        })
    }

    response := dto.MessageListResponse{
        Messages: responses,
        Total:    len(responses),
        Page:     page,
        Limit:    limit,
    }

    c.JSON(http.StatusOK, response)
}

// Helper function to convert message model to response DTO with sender profile image URL
func (h *ChatHandler) messageToResponse(message *chatModel.Message) dto.ChatMessageResponse {
    var senderImage string
    if user, err := h.userService.GetUserByID(message.SenderID); err == nil && user != nil {
        senderImage = user.ProfileImageURL
    }

    return dto.ChatMessageResponse{
        ID:             message.ID,
        SenderID:       message.SenderID,
        SenderImageURL: senderImage,
        Content:        message.Content,
        GroupID:        message.GroupID,
        CreatedAt:      message.CreatedAt,
        UpdatedAt:      message.UpdatedAt,
    }
}
