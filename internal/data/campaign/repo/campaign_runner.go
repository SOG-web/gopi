package repo

import (
	"errors"

	"gorm.io/gorm"

	gormmodel "gopi.com/internal/data/campaign/model/gorm"
	campaignModel "gopi.com/internal/domain/campaign/model"
	campaignRepo "gopi.com/internal/domain/campaign/repo"
)

type GormCampaignRunnerRepository struct {
	db *gorm.DB
}

func NewGormCampaignRunnerRepository(db *gorm.DB) campaignRepo.CampaignRunnerRepository {
	return &GormCampaignRunnerRepository{db: db}
}

func (r *GormCampaignRunnerRepository) Create(runner *campaignModel.CampaignRunner) error {
	dbRunner := gormmodel.FromDomainCampaignRunner(runner)
	if err := r.db.Create(&dbRunner).Error; err != nil {
		return err
	}
	*runner = *gormmodel.ToDomainCampaignRunner(dbRunner)
	return nil
}

func (r *GormCampaignRunnerRepository) GetByID(id string) (*campaignModel.CampaignRunner, error) {
	var cr gormmodel.CampaignRunner
	if err := r.db.First(&cr, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	dcr := gormmodel.ToDomainCampaignRunner(&cr)
	return dcr, nil
}

func (r *GormCampaignRunnerRepository) GetByCampaignID(campaignID string) ([]*campaignModel.CampaignRunner, error) {
	var runners []gormmodel.CampaignRunner
	if err := r.db.Where("campaign_id = ?", campaignID).Order("distance_covered DESC").Find(&runners).Error; err != nil {
		return nil, err
	}

	var result []*campaignModel.CampaignRunner
	for _, cr := range runners {
		result = append(result, gormmodel.ToDomainCampaignRunner(&cr))
	}
	return result, nil
}

func (r *GormCampaignRunnerRepository) GetByOwnerID(ownerID string) ([]*campaignModel.CampaignRunner, error) {
	var runners []gormmodel.CampaignRunner
	if err := r.db.Where("owner_id = ?", ownerID).Order("created_at DESC").Find(&runners).Error; err != nil {
		return nil, err
	}

	var result []*campaignModel.CampaignRunner
	for _, cr := range runners {
		result = append(result, gormmodel.ToDomainCampaignRunner(&cr))
	}
	return result, nil
}

func (r *GormCampaignRunnerRepository) Update(runner *campaignModel.CampaignRunner) error {
	dbRunner := gormmodel.FromDomainCampaignRunner(runner)
	if err := r.db.Save(&dbRunner).Error; err != nil {
		return err
	}
	*runner = *gormmodel.ToDomainCampaignRunner(dbRunner)
	return nil
}

func (r *GormCampaignRunnerRepository) Delete(id string) error {
	return r.db.Delete(&gormmodel.CampaignRunner{}, "id = ?", id).Error
}
