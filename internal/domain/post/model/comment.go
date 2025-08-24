package model

import (
	"gopi.com/internal/domain/model"
)

// Comment is a generic comment that can attach to multiple target types
// via TargetType and TargetID (e.g., "cause", "challenge", "campaign", "post").
// It supports flat or threaded comments via optional ParentID.
type Comment struct {
	model.Base
	AuthorID   string  `json:"author_id"`
	Content    string  `json:"content"`
	TargetType string  `json:"target_type"`
	TargetID   string  `json:"target_id"`
	ParentID   *string `json:"parent_id"`
	IsDeleted  bool    `json:"is_deleted"`
}
