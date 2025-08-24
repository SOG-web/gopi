package repo

import (
	"errors"
	"strings"

	"gorm.io/gorm"

	gormmodel "gopi.com/internal/data/chat/model/gorm"
	chatModel "gopi.com/internal/domain/chat/model"
	chatRepo "gopi.com/internal/domain/chat/repo"
)

type GormGroupRepository struct {
	db *gorm.DB
}

func NewGormGroupRepository(db *gorm.DB) chatRepo.GroupRepository {
	return &GormGroupRepository{db: db}
}

func (r *GormGroupRepository) Create(group *chatModel.Group) error {
	dbGroup := gormmodel.FromDomainGroup(group)
	if err := r.db.Create(&dbGroup).Error; err != nil {
		return err
	}
	*group = *gormmodel.ToDomainGroup(dbGroup)
	return nil
}

func (r *GormGroupRepository) GetByID(id string) (*chatModel.Group, error) {
	var g gormmodel.Group
	if err := r.db.First(&g, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	dg := gormmodel.ToDomainGroup(&g)
	return dg, nil
}

func (r *GormGroupRepository) GetBySlug(slug string) (*chatModel.Group, error) {
	var g gormmodel.Group
	if err := r.db.Where("slug = ?", slug).First(&g).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	dg := gormmodel.ToDomainGroup(&g)
	return dg, nil
}

func (r *GormGroupRepository) GetByCreatorID(creatorID string) ([]*chatModel.Group, error) {
	var groups []gormmodel.Group
	if err := r.db.Where("creator_id = ?", creatorID).Find(&groups).Error; err != nil {
		return nil, err
	}

	var result []*chatModel.Group
	for _, g := range groups {
		result = append(result, gormmodel.ToDomainGroup(&g))
	}
	return result, nil
}

func (r *GormGroupRepository) GetByMemberID(memberID string) ([]*chatModel.Group, error) {
	var groups []gormmodel.Group
	// Query all groups and filter by member ID in application code
	if err := r.db.Find(&groups).Error; err != nil {
		return nil, err
	}

	var result []*chatModel.Group
	for _, g := range groups {
		domainGroup := gormmodel.ToDomainGroup(&g)
		// Check if memberID is in the member IDs slice
		for _, id := range domainGroup.MemberIDs {
			if id == memberID {
				result = append(result, domainGroup)
				break
			}
		}
	}
	return result, nil
}

// SearchByName returns groups whose names contain the query (case-insensitive), paginated
func (r *GormGroupRepository) SearchByName(query string, limit, offset int) ([]*chatModel.Group, error) {
	var groups []gormmodel.Group
	q := "%" + strings.ToLower(query) + "%"
	if err := r.db.Where("LOWER(name) LIKE ?", q).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&groups).Error; err != nil {
		return nil, err
	}

	var result []*chatModel.Group
	for _, g := range groups {
		result = append(result, gormmodel.ToDomainGroup(&g))
	}
	return result, nil
}

func (r *GormGroupRepository) Update(group *chatModel.Group) error {
	dbGroup := gormmodel.FromDomainGroup(group)
	if err := r.db.Save(&dbGroup).Error; err != nil {
		return err
	}
	*group = *gormmodel.ToDomainGroup(dbGroup)
	return nil
}

func (r *GormGroupRepository) Delete(id string) error {
	return r.db.Delete(&gormmodel.Group{}, "id = ?", id).Error
}

func (r *GormGroupRepository) AddMember(groupID, memberID string) error {
	group, err := r.GetByID(groupID)
	if err != nil {
		return err
	}

	// Check if member already exists
	for _, id := range group.MemberIDs {
		if id == memberID {
			return errors.New("member already exists in group")
		}
	}

	group.MemberIDs = append(group.MemberIDs, memberID)
	return r.Update(group)
}

func (r *GormGroupRepository) RemoveMember(groupID, memberID string) error {
	group, err := r.GetByID(groupID)
	if err != nil {
		return err
	}

	// Remove member from slice
	for i, id := range group.MemberIDs {
		if id == memberID {
			group.MemberIDs = append(group.MemberIDs[:i], group.MemberIDs[i+1:]...)
			break
		}
	}

	return r.Update(group)
}

// Message Repository
type GormMessageRepository struct {
	db *gorm.DB
}

func NewGormMessageRepository(db *gorm.DB) chatRepo.MessageRepository {
	return &GormMessageRepository{db: db}
}

func (r *GormMessageRepository) Create(message *chatModel.Message) error {
	dbMessage := gormmodel.FromDomainMessage(message)
	if err := r.db.Create(&dbMessage).Error; err != nil {
		return err
	}
	*message = *gormmodel.ToDomainMessage(dbMessage)
	return nil
}

func (r *GormMessageRepository) GetByID(id string) (*chatModel.Message, error) {
	var m gormmodel.Message
	if err := r.db.First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	dm := gormmodel.ToDomainMessage(&m)
	return dm, nil
}

func (r *GormMessageRepository) GetByGroupID(groupID string, limit, offset int) ([]*chatModel.Message, error) {
	var messages []gormmodel.Message
	if err := r.db.Where("group_id = ?", groupID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error; err != nil {
		return nil, err
	}

	var result []*chatModel.Message
	for _, m := range messages {
		result = append(result, gormmodel.ToDomainMessage(&m))
	}
	return result, nil
}

func (r *GormMessageRepository) GetBySenderID(senderID string) ([]*chatModel.Message, error) {
	var messages []gormmodel.Message
	if err := r.db.Where("sender_id = ?", senderID).Find(&messages).Error; err != nil {
		return nil, err
	}

	var result []*chatModel.Message
	for _, m := range messages {
		result = append(result, gormmodel.ToDomainMessage(&m))
	}
	return result, nil
}

func (r *GormMessageRepository) Update(message *chatModel.Message) error {
	dbMessage := gormmodel.FromDomainMessage(message)
	if err := r.db.Save(&dbMessage).Error; err != nil {
		return err
	}
	*message = *gormmodel.ToDomainMessage(dbMessage)
	return nil
}

func (r *GormMessageRepository) Delete(id string) error {
	return r.db.Delete(&gormmodel.Message{}, "id = ?", id).Error
}
