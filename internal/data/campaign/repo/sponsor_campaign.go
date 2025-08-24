package repo

import (
	"errors"

	"gorm.io/gorm"

	gormmodel "gopi.com/internal/data/campaign/model/gorm"
	campaignModel "gopi.com/internal/domain/campaign/model"
	campaignRepo "gopi.com/internal/domain/campaign/repo"
)

type GormSponsorCampaignRepository struct {
	db *gorm.DB
}

func NewGormSponsorCampaignRepository(db *gorm.DB) campaignRepo.SponsorCampaignRepository {
	return &GormSponsorCampaignRepository{db: db}
}

func (r *GormSponsorCampaignRepository) Create(sponsor *campaignModel.SponsorCampaign) error {
	dbSponsor := gormmodel.FromDomainSponsorCampaign(sponsor)
	if err := r.db.Create(&dbSponsor).Error; err != nil {
		return err
	}
	*sponsor = *gormmodel.ToDomainSponsorCampaign(dbSponsor)
	return nil
}

func (r *GormSponsorCampaignRepository) GetByID(id string) (*campaignModel.SponsorCampaign, error) {
	var sc gormmodel.SponsorCampaign
	if err := r.db.First(&sc, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	dsc := gormmodel.ToDomainSponsorCampaign(&sc)
	return dsc, nil
}

func (r *GormSponsorCampaignRepository) GetByCampaignID(campaignID string) ([]*campaignModel.SponsorCampaign, error) {
	var sponsors []gormmodel.SponsorCampaign
	if err := r.db.Where("campaign_id = ?", campaignID).Order("created_at DESC").Find(&sponsors).Error; err != nil {
		return nil, err
	}

	var result []*campaignModel.SponsorCampaign
	for _, sc := range sponsors {
		result = append(result, gormmodel.ToDomainSponsorCampaign(&sc))
	}
	return result, nil
}

func (r *GormSponsorCampaignRepository) Update(sponsor *campaignModel.SponsorCampaign) error {
	dbSponsor := gormmodel.FromDomainSponsorCampaign(sponsor)
	if err := r.db.Save(&dbSponsor).Error; err != nil {
		return err
	}
	*sponsor = *gormmodel.ToDomainSponsorCampaign(dbSponsor)
	return nil
}

func (r *GormSponsorCampaignRepository) Delete(id string) error {
	return r.db.Delete(&gormmodel.SponsorCampaign{}, "id = ?", id).Error
}
