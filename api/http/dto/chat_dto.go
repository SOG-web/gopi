package dto

import (
	"time"
)

// Group DTOs
type CreateGroupRequest struct {
	Name      string   `json:"name" binding:"required,max=20"`
	Image     string   `json:"image,omitempty"`
	MemberIDs []string `json:"member_ids,omitempty"`
}

type UpdateGroupRequest struct {
	Name  string `json:"name,omitempty" binding:"omitempty,max=20"`
	Image string `json:"image,omitempty"`
}

type GroupResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	MemberIDs []string  `json:"member_ids"`
	CreatorID string    `json:"creator_id"`
	Slug      string    `json:"slug"`
	Image     string    `json:"image"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GroupListResponse struct {
	Groups []GroupResponse `json:"groups"`
	Total  int             `json:"total"`
	Page   int             `json:"page"`
	Limit  int             `json:"limit"`
}

// Message DTOs
type SendMessageRequest struct {
	Content string `json:"content" binding:"required,max=1000"`
}

type UpdateMessageRequest struct {
	Content string `json:"content" binding:"required,max=1000"`
}

type ChatMessageResponse struct {
	ID           string    `json:"id"`
	SenderID     string    `json:"sender_id"`
	SenderImageURL string `json:"sender_image_url,omitempty"`
	Content      string    `json:"content"`
	GroupID      string    `json:"group_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type MessageListResponse struct {
	Messages []ChatMessageResponse `json:"messages"`
	Total    int                   `json:"total"`
	Page     int                   `json:"page"`
	Limit    int                   `json:"limit"`
}

// Group Member Management DTOs
type AddMemberRequest struct {
	MemberID string `json:"member_id" binding:"required"`
}

type RemoveMemberRequest struct {
	MemberID string `json:"member_id" binding:"required"`
}

type GroupMemberResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// WebSocket DTOs
type WebSocketMessage struct {
	Type            string `json:"type"`
	Message         string `json:"message,omitempty"`
	Username        string `json:"username,omitempty"`
	UserID          string `json:"user_id,omitempty"`
	GroupSlug       string `json:"group_slug,omitempty"`
	UserImage       string `json:"user_image,omitempty"`
	IsAuthenticated bool   `json:"is_authenticated,omitempty"`
}

// Search DTOs
type SearchGroupsRequest struct {
	Query string `form:"q" binding:"required"`
	Page  int    `form:"page,default=1" binding:"min=1"`
	Limit int    `form:"limit,default=10" binding:"min=1,max=100"`
}
