package repo

import "gopi.com/internal/domain/chat/model"

type GroupRepository interface {
	Create(group *model.Group) error
	GetByID(id string) (*model.Group, error)
	GetBySlug(slug string) (*model.Group, error)
	GetByCreatorID(creatorID string) ([]*model.Group, error)
	GetByMemberID(memberID string) ([]*model.Group, error)
	// SearchByName returns groups whose names contain the query (case-insensitive), paginated by limit/offset
	SearchByName(query string, limit, offset int) ([]*model.Group, error)
	Update(group *model.Group) error
	Delete(id string) error
	AddMember(groupID, memberID string) error
	RemoveMember(groupID, memberID string) error
}

type MessageRepository interface {
	Create(message *model.Message) error
	GetByID(id string) (*model.Message, error)
	GetByGroupID(groupID string, limit, offset int) ([]*model.Message, error)
	GetBySenderID(senderID string) ([]*model.Message, error)
	Update(message *model.Message) error
	Delete(id string) error
}
