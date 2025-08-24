package repo

import "gopi.com/internal/domain/challenge/model"

type ChallengeRepository interface {
	Create(challenge *model.Challenge) error
	GetByID(id string) (*model.Challenge, error)
	GetBySlug(slug string) (*model.Challenge, error)
	GetByOwnerID(ownerID string) ([]*model.Challenge, error)
	Update(challenge *model.Challenge) error
	Delete(id string) error
	List(limit, offset int) ([]*model.Challenge, error)
}

type CauseRepository interface {
	Create(cause *model.Cause) error
	GetByID(id string) (*model.Cause, error)
	GetBySlug(slug string) (*model.Cause, error)
	GetByChallengeID(challengeID string) ([]*model.Cause, error)
	GetByOwnerID(ownerID string) ([]*model.Cause, error)
	Update(cause *model.Cause) error
	Delete(id string) error
}

type CauseRunnerRepository interface {
	Create(runner *model.CauseRunner) error
	GetByID(id string) (*model.CauseRunner, error)
	GetByCauseID(causeID string) ([]*model.CauseRunner, error)
	GetByOwnerID(ownerID string) ([]*model.CauseRunner, error)
	Update(runner *model.CauseRunner) error
	Delete(id string) error
	GetLeaderboard() ([]*model.CauseRunner, error)
}

type SponsorChallengeRepository interface {
	Create(sponsor *model.SponsorChallenge) error
	GetByID(id string) (*model.SponsorChallenge, error)
	GetByChallengeID(challengeID string) ([]*model.SponsorChallenge, error)
	GetBySponsorID(sponsorID string) ([]*model.SponsorChallenge, error)
	Update(sponsor *model.SponsorChallenge) error
	Delete(id string) error
}

type SponsorCauseRepository interface {
	Create(sponsor *model.SponsorCause) error
	GetByID(id string) (*model.SponsorCause, error)
	GetByCauseID(causeID string) ([]*model.SponsorCause, error)
	GetBySponsorID(sponsorID string) ([]*model.SponsorCause, error)
	Update(sponsor *model.SponsorCause) error
	Delete(id string) error
}

type CauseBuyerRepository interface {
	Create(buyer *model.CauseBuyer) error
	GetByID(id string) (*model.CauseBuyer, error)
	GetByCauseID(causeID string) ([]*model.CauseBuyer, error)
	GetByBuyerID(buyerID string) ([]*model.CauseBuyer, error)
	Update(buyer *model.CauseBuyer) error
	Delete(id string) error
}
