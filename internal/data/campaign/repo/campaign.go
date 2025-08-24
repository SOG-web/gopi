package repo

import (
	"errors"

	"gorm.io/gorm"

	gormmodel "gopi.com/internal/data/campaign/model/gorm"
	campaignModel "gopi.com/internal/domain/campaign/model"
	campaignRepo "gopi.com/internal/domain/campaign/repo"
)

type GormCampaignRepository struct {
	db *gorm.DB
}

func NewGormCampaignRepository(db *gorm.DB) campaignRepo.CampaignRepository {
	return &GormCampaignRepository{db: db}
}

func (r *GormCampaignRepository) Create(campaign *campaignModel.Campaign) error {
	dbCampaign := gormmodel.FromDomainCampaign(campaign)
	if err := r.db.Create(&dbCampaign).Error; err != nil {
		return err
	}
	*campaign = *gormmodel.ToDomainCampaign(dbCampaign)
	return nil
}

func (r *GormCampaignRepository) GetByID(id string) (*campaignModel.Campaign, error) {
	var c gormmodel.Campaign
	if err := r.db.Preload("Members").Preload("Sponsors").First(&c, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	dc := gormmodel.ToDomainCampaign(&c)
	return dc, nil
}

func (r *GormCampaignRepository) GetBySlug(slug string) (*campaignModel.Campaign, error) {
	var c gormmodel.Campaign
	if err := r.db.Preload("Members").Preload("Sponsors").Where("slug = ?", slug).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	dc := gormmodel.ToDomainCampaign(&c)
	return dc, nil
}

func (r *GormCampaignRepository) GetByOwnerID(ownerID string) ([]*campaignModel.Campaign, error) {
	var campaigns []gormmodel.Campaign
	if err := r.db.Preload("Members").Preload("Sponsors").Where("owner_id = ?", ownerID).Find(&campaigns).Error; err != nil {
		return nil, err
	}

	var result []*campaignModel.Campaign
	for _, c := range campaigns {
		result = append(result, gormmodel.ToDomainCampaign(&c))
	}
	return result, nil
}

func (r *GormCampaignRepository) Update(campaign *campaignModel.Campaign) error {
	dbCampaign := gormmodel.FromDomainCampaign(campaign)
	if err := r.db.Save(&dbCampaign).Error; err != nil {
		return err
	}
	*campaign = *gormmodel.ToDomainCampaign(dbCampaign)
	return nil
}

func (r *GormCampaignRepository) Delete(id string) error {
	return r.db.Delete(&gormmodel.Campaign{}, "id = ?", id).Error
}

func (r *GormCampaignRepository) List(limit, offset int) ([]*campaignModel.Campaign, error) {
	var campaigns []gormmodel.Campaign
	if err := r.db.Preload("Members").Preload("Sponsors").Limit(limit).Offset(offset).Order("created_at DESC").Find(&campaigns).Error; err != nil {
		return nil, err
	}

	var result []*campaignModel.Campaign
	for _, c := range campaigns {
		result = append(result, gormmodel.ToDomainCampaign(&c))
	}
	return result, nil
}

func (r *GormCampaignRepository) Search(query string, limit, offset int) ([]*campaignModel.Campaign, error) {
	var campaigns []gormmodel.Campaign
	searchPattern := "%" + query + "%"
	
	if err := r.db.Preload("Members").Preload("Sponsors").
		Where("name ILIKE ? OR description ILIKE ? OR location ILIKE ?", searchPattern, searchPattern, searchPattern).
		Limit(limit).Offset(offset).Order("created_at DESC").Find(&campaigns).Error; err != nil {
		return nil, err
	}

	var result []*campaignModel.Campaign
	for _, c := range campaigns {
		result = append(result, gormmodel.ToDomainCampaign(&c))
	}
	return result, nil
}

// Many-to-many relationship methods
func (r *GormCampaignRepository) AddMember(campaignID, userID string) error {
	member := &gormmodel.CampaignMember{
		CampaignID: campaignID,
		UserID:     userID,
	}
	return r.db.Create(member).Error
}

func (r *GormCampaignRepository) RemoveMember(campaignID, userID string) error {
	return r.db.Where("campaign_id = ? AND user_id = ?", campaignID, userID).
		Delete(&gormmodel.CampaignMember{}).Error
}

func (r *GormCampaignRepository) IsMember(campaignID, userID string) (bool, error) {
	var count int64
	err := r.db.Model(&gormmodel.CampaignMember{}).
		Where("campaign_id = ? AND user_id = ?", campaignID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *GormCampaignRepository) AddSponsor(campaignID, userID string) error {
	sponsor := &gormmodel.CampaignSponsor{
		CampaignID: campaignID,
		UserID:     userID,
	}
	return r.db.Create(sponsor).Error
}

func (r *GormCampaignRepository) RemoveSponsor(campaignID, userID string) error {
	return r.db.Where("campaign_id = ? AND user_id = ?", campaignID, userID).
		Delete(&gormmodel.CampaignSponsor{}).Error
}

func (r *GormCampaignRepository) IsSponsor(campaignID, userID string) (bool, error) {
	var count int64
	err := r.db.Model(&gormmodel.CampaignSponsor{}).
		Where("campaign_id = ? AND user_id = ?", campaignID, userID).
		Count(&count).Error
	return count > 0, err
}
