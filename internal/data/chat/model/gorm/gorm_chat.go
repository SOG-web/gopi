package gorm

import (
	"encoding/json"
	"time"

	chatModel "gopi.com/internal/domain/chat/model"
	"gopi.com/internal/domain/model"
	"gopi.com/internal/lib/id"
	"gorm.io/gorm"
)

type Group struct {
	ID        string `gorm:"type:varchar(255);primary_key"`
	Name      string `gorm:"not null"`
	MemberIDs string `gorm:"type:text"` // JSON array stored as text
	CreatorID string `gorm:"not null"`
	Slug      string `gorm:"unique"`
	Image     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Message struct {
	ID        string `gorm:"type:varchar(255);primary_key"`
	SenderID  string `gorm:"not null"`
	Content   string `gorm:"type:text;not null"`
	GroupID   string `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (g *Group) BeforeCreate(tx *gorm.DB) (err error) {
	if g.ID == "" {
		g.ID = id.New()
	}
	return
}

func (m *Message) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = id.New()
	}
	return
}

// Conversion functions
func FromDomainGroup(g *chatModel.Group) *Group {
	memberIDs := ""
	if len(g.MemberIDs) > 0 {
		// Marshal member IDs to JSON
		if jsonData, err := json.Marshal(g.MemberIDs); err == nil {
			memberIDs = string(jsonData)
		}
	}

	return &Group{
		ID:        g.ID,
		Name:      g.Name,
		MemberIDs: memberIDs,
		CreatorID: g.CreatorID,
		Slug:      g.Slug,
		Image:     g.Image,
		CreatedAt: g.CreatedAt,
		UpdatedAt: g.UpdatedAt,
	}
}

func ToDomainGroup(g *Group) *chatModel.Group {
	var memberIDs []string
	if g.MemberIDs != "" {
		// Unmarshal member IDs from JSON
		if err := json.Unmarshal([]byte(g.MemberIDs), &memberIDs); err != nil {
			// If unmarshaling fails, leave memberIDs empty
			memberIDs = []string{}
		}
	}

	return &chatModel.Group{
		Base: model.Base{
			ID:        g.ID,
			CreatedAt: g.CreatedAt,
			UpdatedAt: g.UpdatedAt,
		},
		Name:      g.Name,
		MemberIDs: memberIDs,
		CreatorID: g.CreatorID,
		Slug:      g.Slug,
		Image:     g.Image,
	}
}

func FromDomainMessage(m *chatModel.Message) *Message {
	return &Message{
		ID:        m.ID,
		SenderID:  m.SenderID,
		Content:   m.Content,
		GroupID:   m.GroupID,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func ToDomainMessage(m *Message) *chatModel.Message {
	return &chatModel.Message{
		Base: model.Base{
			ID:        m.ID,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		},
		SenderID: m.SenderID,
		Content:  m.Content,
		GroupID:  m.GroupID,
	}
}
