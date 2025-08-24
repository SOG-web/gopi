package model

import "gopi.com/internal/domain/model"

type Group struct {
	model.Base
	Name      string   `json:"name"`
	MemberIDs []string `json:"member_ids"`
	CreatorID string   `json:"creator_id"`
	Slug      string   `json:"slug"`
	Image     string   `json:"image"`
}

type Message struct {
	model.Base
	SenderID string `json:"sender_id"`
	Content  string `json:"content"`
	GroupID  string `json:"group_id"`
}
