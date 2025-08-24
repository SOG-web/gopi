package repo

import "gopi.com/internal/domain/campaign/model"

type CampaignRepository interface {
	Create(campaign *model.Campaign) error
	GetByID(id string) (*model.Campaign, error)
	GetBySlug(slug string) (*model.Campaign, error)
	GetByOwnerID(ownerID string) ([]*model.Campaign, error)
	Update(campaign *model.Campaign) error
	Delete(id string) error
	List(limit, offset int) ([]*model.Campaign, error)
	Search(query string, limit, offset int) ([]*model.Campaign, error)
	
	// Many-to-many relationship methods
	AddMember(campaignID, userID string) error
	RemoveMember(campaignID, userID string) error
	IsMember(campaignID, userID string) (bool, error)
	AddSponsor(campaignID, userID string) error
	RemoveSponsor(campaignID, userID string) error
	IsSponsor(campaignID, userID string) (bool, error)
}

type CampaignRunnerRepository interface {
	Create(runner *model.CampaignRunner) error
	GetByID(id string) (*model.CampaignRunner, error)
	GetByCampaignID(campaignID string) ([]*model.CampaignRunner, error)
	GetByOwnerID(ownerID string) ([]*model.CampaignRunner, error)
	Update(runner *model.CampaignRunner) error
	Delete(id string) error
}

type SponsorCampaignRepository interface {
	Create(sponsor *model.SponsorCampaign) error
	GetByID(id string) (*model.SponsorCampaign, error)
	GetByCampaignID(campaignID string) ([]*model.SponsorCampaign, error)
	Update(sponsor *model.SponsorCampaign) error
	Delete(id string) error
}
