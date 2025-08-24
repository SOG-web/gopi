package repo

import (
	"errors"

	"gorm.io/gorm"

	gormmodel "gopi.com/internal/data/challenge/model/gorm"
	challengeModel "gopi.com/internal/domain/challenge/model"
	challengeRepo "gopi.com/internal/domain/challenge/repo"
)

type GormChallengeRepository struct {
	db *gorm.DB
}

func NewGormChallengeRepository(db *gorm.DB) challengeRepo.ChallengeRepository {
	return &GormChallengeRepository{db: db}
}

func (r *GormChallengeRepository) Create(challenge *challengeModel.Challenge) error {
	dbChallenge := gormmodel.FromDomainChallenge(challenge)
	if err := r.db.Create(&dbChallenge).Error; err != nil {
		return err
	}
	*challenge = *gormmodel.ToDomainChallenge(dbChallenge)
	return nil
}

func (r *GormChallengeRepository) GetByID(id string) (*challengeModel.Challenge, error) {
	var c gormmodel.Challenge
	if err := r.db.First(&c, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	dc := gormmodel.ToDomainChallenge(&c)
	return dc, nil
}

func (r *GormChallengeRepository) GetBySlug(slug string) (*challengeModel.Challenge, error) {
	var c gormmodel.Challenge
	if err := r.db.Where("slug = ?", slug).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	dc := gormmodel.ToDomainChallenge(&c)
	return dc, nil
}

func (r *GormChallengeRepository) GetByOwnerID(ownerID string) ([]*challengeModel.Challenge, error) {
	var challenges []gormmodel.Challenge
	if err := r.db.Where("owner_id = ?", ownerID).Find(&challenges).Error; err != nil {
		return nil, err
	}

	var result []*challengeModel.Challenge
	for _, c := range challenges {
		result = append(result, gormmodel.ToDomainChallenge(&c))
	}
	return result, nil
}

func (r *GormChallengeRepository) Update(challenge *challengeModel.Challenge) error {
	dbChallenge := gormmodel.FromDomainChallenge(challenge)
	if err := r.db.Save(&dbChallenge).Error; err != nil {
		return err
	}
	*challenge = *gormmodel.ToDomainChallenge(dbChallenge)
	return nil
}

func (r *GormChallengeRepository) Delete(id string) error {
	return r.db.Delete(&gormmodel.Challenge{}, "id = ?", id).Error
}

func (r *GormChallengeRepository) List(limit, offset int) ([]*challengeModel.Challenge, error) {
	var challenges []gormmodel.Challenge
	if err := r.db.Limit(limit).Offset(offset).Order("created_at DESC").Find(&challenges).Error; err != nil {
		return nil, err
	}

	var result []*challengeModel.Challenge
	for _, c := range challenges {
		result = append(result, gormmodel.ToDomainChallenge(&c))
	}
	return result, nil
}

// Cause Repository
type GormCauseRepository struct {
	db *gorm.DB
}

func NewGormCauseRepository(db *gorm.DB) challengeRepo.CauseRepository {
	return &GormCauseRepository{db: db}
}

func (r *GormCauseRepository) Create(cause *challengeModel.Cause) error {
	dbCause := gormmodel.FromDomainCause(cause)
	if err := r.db.Create(&dbCause).Error; err != nil {
		return err
	}
	*cause = *gormmodel.ToDomainCause(dbCause)
	return nil
}

func (r *GormCauseRepository) GetByID(id string) (*challengeModel.Cause, error) {
	var c gormmodel.Cause
	if err := r.db.First(&c, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	dc := gormmodel.ToDomainCause(&c)
	return dc, nil
}

func (r *GormCauseRepository) GetBySlug(slug string) (*challengeModel.Cause, error) {
	var c gormmodel.Cause
	if err := r.db.Where("slug = ?", slug).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	dc := gormmodel.ToDomainCause(&c)
	return dc, nil
}

func (r *GormCauseRepository) GetByChallengeID(challengeID string) ([]*challengeModel.Cause, error) {
	var causes []gormmodel.Cause
	if err := r.db.Where("challenge_id = ?", challengeID).Find(&causes).Error; err != nil {
		return nil, err
	}

	var result []*challengeModel.Cause
	for _, c := range causes {
		result = append(result, gormmodel.ToDomainCause(&c))
	}
	return result, nil
}

func (r *GormCauseRepository) GetByOwnerID(ownerID string) ([]*challengeModel.Cause, error) {
	var causes []gormmodel.Cause
	if err := r.db.Where("owner_id = ?", ownerID).Find(&causes).Error; err != nil {
		return nil, err
	}

	var result []*challengeModel.Cause
	for _, c := range causes {
		result = append(result, gormmodel.ToDomainCause(&c))
	}
	return result, nil
}

func (r *GormCauseRepository) Update(cause *challengeModel.Cause) error {
	dbCause := gormmodel.FromDomainCause(cause)
	if err := r.db.Save(&dbCause).Error; err != nil {
		return err
	}
	*cause = *gormmodel.ToDomainCause(dbCause)
	return nil
}

func (r *GormCauseRepository) Delete(id string) error {
	return r.db.Delete(&gormmodel.Cause{}, "id = ?", id).Error
}

// CauseRunner Repository
type GormCauseRunnerRepository struct {
	db *gorm.DB
}

func NewGormCauseRunnerRepository(db *gorm.DB) challengeRepo.CauseRunnerRepository {
	return &GormCauseRunnerRepository{db: db}
}

func (r *GormCauseRunnerRepository) Create(runner *challengeModel.CauseRunner) error {
	dbRunner := gormmodel.FromDomainCauseRunner(runner)
	if err := r.db.Create(&dbRunner).Error; err != nil {
		return err
	}
	*runner = *gormmodel.ToDomainCauseRunner(dbRunner)
	return nil
}

func (r *GormCauseRunnerRepository) GetByID(id string) (*challengeModel.CauseRunner, error) {
	var cr gormmodel.CauseRunner
	if err := r.db.First(&cr, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	dcr := gormmodel.ToDomainCauseRunner(&cr)
	return dcr, nil
}

func (r *GormCauseRunnerRepository) GetByCauseID(causeID string) ([]*challengeModel.CauseRunner, error) {
	var runners []gormmodel.CauseRunner
	if err := r.db.Where("cause_id = ?", causeID).Find(&runners).Error; err != nil {
		return nil, err
	}

	var result []*challengeModel.CauseRunner
	for _, cr := range runners {
		result = append(result, gormmodel.ToDomainCauseRunner(&cr))
	}
	return result, nil
}

func (r *GormCauseRunnerRepository) GetByOwnerID(ownerID string) ([]*challengeModel.CauseRunner, error) {
	var runners []gormmodel.CauseRunner
	if err := r.db.Where("owner_id = ?", ownerID).Find(&runners).Error; err != nil {
		return nil, err
	}

	var result []*challengeModel.CauseRunner
	for _, cr := range runners {
		result = append(result, gormmodel.ToDomainCauseRunner(&cr))
	}
	return result, nil
}

func (r *GormCauseRunnerRepository) Update(runner *challengeModel.CauseRunner) error {
	dbRunner := gormmodel.FromDomainCauseRunner(runner)
	if err := r.db.Save(&dbRunner).Error; err != nil {
		return err
	}
	*runner = *gormmodel.ToDomainCauseRunner(dbRunner)
	return nil
}

func (r *GormCauseRunnerRepository) Delete(id string) error {
	return r.db.Delete(&gormmodel.CauseRunner{}, "id = ?", id).Error
}

func (r *GormCauseRunnerRepository) GetLeaderboard() ([]*challengeModel.CauseRunner, error) {
	var runners []gormmodel.CauseRunner
	if err := r.db.Order("distance_covered DESC").Limit(10).Find(&runners).Error; err != nil {
		return nil, err
	}

	var result []*challengeModel.CauseRunner
	for _, cr := range runners {
		result = append(result, gormmodel.ToDomainCauseRunner(&cr))
	}
	return result, nil
}

// SponsorChallenge Repository
type GormSponsorChallengeRepository struct {
	db *gorm.DB
}

func NewGormSponsorChallengeRepository(db *gorm.DB) challengeRepo.SponsorChallengeRepository {
	return &GormSponsorChallengeRepository{db: db}
}

func (r *GormSponsorChallengeRepository) Create(sponsor *challengeModel.SponsorChallenge) error {
	dbSponsor := gormmodel.FromDomainSponsorChallenge(sponsor)
	if err := r.db.Create(&dbSponsor).Error; err != nil {
		return err
	}
	*sponsor = *gormmodel.ToDomainSponsorChallenge(dbSponsor)
	return nil
}

func (r *GormSponsorChallengeRepository) GetByID(id string) (*challengeModel.SponsorChallenge, error) {
	var sc gormmodel.SponsorChallenge
	if err := r.db.First(&sc, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	dsc := gormmodel.ToDomainSponsorChallenge(&sc)
	return dsc, nil
}

func (r *GormSponsorChallengeRepository) GetBySponsorID(sponsorID string) ([]*challengeModel.SponsorChallenge, error) {
	var sponsors []gormmodel.SponsorChallenge
	if err := r.db.Where("sponsor_id = ?", sponsorID).Find(&sponsors).Error; err != nil {
		return nil, err
	}

	var result []*challengeModel.SponsorChallenge
	for _, sc := range sponsors {
		result = append(result, gormmodel.ToDomainSponsorChallenge(&sc))
	}
	return result, nil
}

func (r *GormSponsorChallengeRepository) GetByChallengeID(challengeID string) ([]*challengeModel.SponsorChallenge, error) {
	var sponsors []gormmodel.SponsorChallenge
	if err := r.db.Where("challenge_id = ?", challengeID).Find(&sponsors).Error; err != nil {
		return nil, err
	}

	var result []*challengeModel.SponsorChallenge
	for _, sc := range sponsors {
		result = append(result, gormmodel.ToDomainSponsorChallenge(&sc))
	}
	return result, nil
}

func (r *GormSponsorChallengeRepository) Update(sponsor *challengeModel.SponsorChallenge) error {
	dbSponsor := gormmodel.FromDomainSponsorChallenge(sponsor)
	if err := r.db.Save(&dbSponsor).Error; err != nil {
		return err
	}
	*sponsor = *gormmodel.ToDomainSponsorChallenge(dbSponsor)
	return nil
}

func (r *GormSponsorChallengeRepository) Delete(id string) error {
	return r.db.Delete(&gormmodel.SponsorChallenge{}, "id = ?", id).Error
}

// SponsorCause Repository
type GormSponsorCauseRepository struct {
	db *gorm.DB
}

func NewGormSponsorCauseRepository(db *gorm.DB) challengeRepo.SponsorCauseRepository {
	return &GormSponsorCauseRepository{db: db}
}

func (r *GormSponsorCauseRepository) Create(sponsor *challengeModel.SponsorCause) error {
	dbSponsor := gormmodel.FromDomainSponsorCause(sponsor)
	if err := r.db.Create(&dbSponsor).Error; err != nil {
		return err
	}
	*sponsor = *gormmodel.ToDomainSponsorCause(dbSponsor)
	return nil
}

func (r *GormSponsorCauseRepository) GetByID(id string) (*challengeModel.SponsorCause, error) {
	var sc gormmodel.SponsorCause
	if err := r.db.First(&sc, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	dsc := gormmodel.ToDomainSponsorCause(&sc)
	return dsc, nil
}

func (r *GormSponsorCauseRepository) GetBySponsorID(sponsorID string) ([]*challengeModel.SponsorCause, error) {
	var sponsors []gormmodel.SponsorCause
	if err := r.db.Where("sponsor_id = ?", sponsorID).Find(&sponsors).Error; err != nil {
		return nil, err
	}

	var result []*challengeModel.SponsorCause
	for _, sc := range sponsors {
		result = append(result, gormmodel.ToDomainSponsorCause(&sc))
	}
	return result, nil
}

func (r *GormSponsorCauseRepository) GetByCauseID(causeID string) ([]*challengeModel.SponsorCause, error) {
	var sponsors []gormmodel.SponsorCause
	if err := r.db.Where("cause_id = ?", causeID).Find(&sponsors).Error; err != nil {
		return nil, err
	}

	var result []*challengeModel.SponsorCause
	for _, sc := range sponsors {
		result = append(result, gormmodel.ToDomainSponsorCause(&sc))
	}
	return result, nil
}

func (r *GormSponsorCauseRepository) Update(sponsor *challengeModel.SponsorCause) error {
	dbSponsor := gormmodel.FromDomainSponsorCause(sponsor)
	if err := r.db.Save(&dbSponsor).Error; err != nil {
		return err
	}
	*sponsor = *gormmodel.ToDomainSponsorCause(dbSponsor)
	return nil
}

func (r *GormSponsorCauseRepository) Delete(id string) error {
	return r.db.Delete(&gormmodel.SponsorCause{}, "id = ?", id).Error
}

// CauseBuyer Repository
type GormCauseBuyerRepository struct {
	db *gorm.DB
}

func NewGormCauseBuyerRepository(db *gorm.DB) challengeRepo.CauseBuyerRepository {
	return &GormCauseBuyerRepository{db: db}
}

func (r *GormCauseBuyerRepository) Create(buyer *challengeModel.CauseBuyer) error {
	dbBuyer := gormmodel.FromDomainCauseBuyer(buyer)
	if err := r.db.Create(&dbBuyer).Error; err != nil {
		return err
	}
	*buyer = *gormmodel.ToDomainCauseBuyer(dbBuyer)
	return nil
}

func (r *GormCauseBuyerRepository) GetByID(id string) (*challengeModel.CauseBuyer, error) {
	var cb gormmodel.CauseBuyer
	if err := r.db.First(&cb, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	dcb := gormmodel.ToDomainCauseBuyer(&cb)
	return dcb, nil
}

func (r *GormCauseBuyerRepository) GetByCauseID(causeID string) ([]*challengeModel.CauseBuyer, error) {
	var buyers []gormmodel.CauseBuyer
	if err := r.db.Where("cause_id = ?", causeID).Find(&buyers).Error; err != nil {
		return nil, err
	}

	var result []*challengeModel.CauseBuyer
	for _, cb := range buyers {
		result = append(result, gormmodel.ToDomainCauseBuyer(&cb))
	}
	return result, nil
}

func (r *GormCauseBuyerRepository) GetByBuyerID(buyerID string) ([]*challengeModel.CauseBuyer, error) {
	var buyers []gormmodel.CauseBuyer
	if err := r.db.Where("buyer_id = ?", buyerID).Find(&buyers).Error; err != nil {
		return nil, err
	}

	var result []*challengeModel.CauseBuyer
	for _, cb := range buyers {
		result = append(result, gormmodel.ToDomainCauseBuyer(&cb))
	}
	return result, nil
}

func (r *GormCauseBuyerRepository) Update(buyer *challengeModel.CauseBuyer) error {
	dbBuyer := gormmodel.FromDomainCauseBuyer(buyer)
	if err := r.db.Save(&dbBuyer).Error; err != nil {
		return err
	}
	*buyer = *gormmodel.ToDomainCauseBuyer(dbBuyer)
	return nil
}

func (r *GormCauseBuyerRepository) Delete(id string) error {
	return r.db.Delete(&gormmodel.CauseBuyer{}, "id = ?", id).Error
}
