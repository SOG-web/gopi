package chat

import (
	"errors"
	"time"
	"unicode/utf8"

	chatModel "gopi.com/internal/domain/chat/model"
	"gopi.com/internal/domain/chat/repo"
	"gopi.com/internal/domain/model"
	"gopi.com/internal/lib/id"
)

type ChatService struct {
	groupRepo   repo.GroupRepository
	messageRepo repo.MessageRepository
}

func NewChatService(
	groupRepo repo.GroupRepository,
	messageRepo repo.MessageRepository,
) *ChatService {
	return &ChatService{
		groupRepo:   groupRepo,
		messageRepo: messageRepo,
	}
}

func (s *ChatService) CreateGroup(creatorID, name, image string, memberIDs []string) (*chatModel.Group, error) {
	// Enforce max 20 characters for group name
	if utf8.RuneCountInString(name) > 20 {
		return nil, errors.New("group name must be at most 20 characters")
	}
	// Add creator to members if not already included
	found := false
	for _, memberID := range memberIDs {
		if memberID == creatorID {
			found = true
			break
		}
	}
	if !found {
		memberIDs = append(memberIDs, creatorID)
	}

	group := &chatModel.Group{
		Base: model.Base{
			ID:        id.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:      name,
		MemberIDs: memberIDs,
		CreatorID: creatorID,
		Image:     image,
		Slug:      generateSlug(name),
	}

	err := s.groupRepo.Create(group)
	if err != nil {
		return nil, err
	}

	return group, nil
}

func (s *ChatService) GetGroupByID(id string) (*chatModel.Group, error) {
	return s.groupRepo.GetByID(id)
}

func (s *ChatService) GetGroupBySlug(slug string) (*chatModel.Group, error) {
	return s.groupRepo.GetBySlug(slug)
}

func (s *ChatService) GetGroupsByCreator(creatorID string) ([]*chatModel.Group, error) {
	return s.groupRepo.GetByCreatorID(creatorID)
}

func (s *ChatService) GetGroupsByMember(memberID string) ([]*chatModel.Group, error) {
	return s.groupRepo.GetByMemberID(memberID)
}

func (s *ChatService) UpdateGroup(group *chatModel.Group) error {
	// Enforce max 20 characters for group name
	if utf8.RuneCountInString(group.Name) > 20 {
		return errors.New("group name must be at most 20 characters")
	}
	return s.groupRepo.Update(group)
}

func (s *ChatService) AddMemberToGroup(groupID, memberID, requesterID string) error {
	group, err := s.groupRepo.GetByID(groupID)
	if err != nil {
		return err
	}

	// Check if requester is the creator or a member
	if group.CreatorID != requesterID {
		isMember := false
		for _, id := range group.MemberIDs {
			if id == requesterID {
				isMember = true
				break
			}
		}
		if !isMember {
			return errors.New("only group members can add new members")
		}
	}

	return s.groupRepo.AddMember(groupID, memberID)
}

func (s *ChatService) RemoveMemberFromGroup(groupID, memberID, requesterID string) error {
	group, err := s.groupRepo.GetByID(groupID)
	if err != nil {
		return err
	}

	// Check if requester is the creator or the member themselves
	if group.CreatorID != requesterID && memberID != requesterID {
		return errors.New("only group creator or the member themselves can remove a member")
	}

	return s.groupRepo.RemoveMember(groupID, memberID)
}

func (s *ChatService) SendMessage(senderID, groupID, content string) (*chatModel.Message, error) {
	// Check if sender is a member of the group
	group, err := s.groupRepo.GetByID(groupID)
	if err != nil {
		return nil, err
	}

	isMember := false
	for _, memberID := range group.MemberIDs {
		if memberID == senderID {
			isMember = true
			break
		}
	}
	if !isMember {
		return nil, errors.New("sender is not a member of this group")
	}

	message := &chatModel.Message{
		Base: model.Base{
			ID:        id.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		SenderID: senderID,
		Content:  content,
		GroupID:  groupID,
	}

	err = s.messageRepo.Create(message)
	if err != nil {
		return nil, err
	}

	return message, nil
}

func (s *ChatService) GetMessageByID(id string) (*chatModel.Message, error) {
	return s.messageRepo.GetByID(id)
}

func (s *ChatService) GetMessagesByGroup(groupID string, limit, offset int) ([]*chatModel.Message, error) {
	return s.messageRepo.GetByGroupID(groupID, limit, offset)
}

func (s *ChatService) GetMessagesBySender(senderID string) ([]*chatModel.Message, error) {
	return s.messageRepo.GetBySenderID(senderID)
}

func (s *ChatService) UpdateMessage(messageID, content, requesterID string) error {
	message, err := s.messageRepo.GetByID(messageID)
	if err != nil {
		return err
	}

	// Only the sender can update their message
	if message.SenderID != requesterID {
		return errors.New("only the sender can update their message")
	}

	message.Content = content
	message.UpdatedAt = time.Now()

	return s.messageRepo.Update(message)
}

func (s *ChatService) DeleteMessage(messageID, requesterID string) error {
	message, err := s.messageRepo.GetByID(messageID)
	if err != nil {
		return err
	}

	// Check if requester is the sender or group creator
	group, err := s.groupRepo.GetByID(message.GroupID)
	if err != nil {
		return err
	}

	if message.SenderID != requesterID && group.CreatorID != requesterID {
		return errors.New("only the sender or group creator can delete a message")
	}

	return s.messageRepo.Delete(messageID)
}

func (s *ChatService) DeleteGroup(groupID, requesterID string) error {
	group, err := s.groupRepo.GetByID(groupID)
	if err != nil {
		return err
	}

	// Only the creator can delete the group
	if group.CreatorID != requesterID {
		return errors.New("only the group creator can delete the group")
	}

	return s.groupRepo.Delete(groupID)
}

// Helper function to generate slug from name
func generateSlug(name string) string {
	// Simplified implementation - in real app you'd use a proper slug library
	return name + "-" + id.New()[:8]
}

// SearchGroupsByName performs case-insensitive paginated search by group name
func (s *ChatService) SearchGroupsByName(query string, limit, offset int) ([]*chatModel.Group, error) {
	return s.groupRepo.SearchByName(query, limit, offset)
}
