package dto

// Post DTOs

type CreatePostRequest struct {
	Title         string `json:"title" binding:"required,max=200"`
	Content       string `json:"content" binding:"required"`
	CoverImageURL string `json:"cover_image_url"`
	Publish       bool   `json:"publish"`
}

type UpdatePostRequest struct {
	Title         string `json:"title" binding:"max=200"`
	Content       string `json:"content"`
	CoverImageURL string `json:"cover_image_url"`
}

type PostResponse struct {
	Success    bool        `json:"success"`
	StatusCode int         `json:"status_code"`
	Data       interface{} `json:"data,omitempty"`
	Message    string      `json:"message,omitempty"`
}

type ListPostsResponse struct {
	Success    bool        `json:"success"`
	StatusCode int         `json:"status_code"`
	Data       interface{} `json:"data,omitempty"`
	Count      int         `json:"count"`
}

// Comment DTOs

type CreateCommentRequest struct {
	TargetType string  `json:"target_type" binding:"required"`
	TargetID   string  `json:"target_id" binding:"required"`
	Content    string  `json:"content" binding:"required"`
	ParentID   *string `json:"parent_id"`
}

type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required"`
}

type CommentResponse struct {
	Success    bool        `json:"success"`
	StatusCode int         `json:"status_code"`
	Data       interface{} `json:"data,omitempty"`
	Message    string      `json:"message,omitempty"`
}

type ListCommentsResponse struct {
	Success    bool        `json:"success"`
	StatusCode int         `json:"status_code"`
	Data       interface{} `json:"data,omitempty"`
	Count      int         `json:"count"`
}
