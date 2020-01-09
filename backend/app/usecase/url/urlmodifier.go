package url

import (
	"short/app/entity"
	"short/app/usecase/repository"
	"short/app/usecase/validator"
)

var _ Modifier = (*ModifierPersist)(nil)

// Modifier represents URL modifier
type Modifier interface {
	UpdateURL(oldAlias *string, newAlias *string, user entity.User) (entity.URL, error)
}

// ModifierPersist represents URL modifier that modifies URL from persistent
// storage, such as database
type ModifierPersist struct {
	urlRepo             repository.URL
	userURLRelationRepo repository.UserURLRelation
	aliasValidator      validator.CustomAlias
}

// UpdateURL updates URL matching oldAlias with newAlias
func (m ModifierPersist) UpdateURL(oldAlias *string, newAlias *string, user entity.User) (entity.URL, error) {
	url, err := m.getURL(*oldAlias)
	if err != nil {
		return entity.URL{}, err
	}

	if !m.aliasValidator.IsValid(newAlias) {
		return entity.URL{}, ErrInvalidCustomAlias(*newAlias)
	}

	url.Alias = *newAlias

	err = m.urlRepo.Update(url)
	if err != nil {
		return entity.URL{}, err
	}

	err = m.userURLRelationRepo.UpdateRelation(user, url)
	if err != nil {
		return entity.URL{}, err
	}

	return url, nil
}

func (m ModifierPersist) getURL(alias string) (entity.URL, error) {
	url, err := m.urlRepo.GetByAlias(alias)
	if err != nil {
		return entity.URL{}, err
	}

	return url, nil
}

// NewModifierPersist creates persistent URL modifier
func NewModifierPersist(
	urlRepo repository.URL,
	userURLRelationRepo repository.UserURLRelation,
	aliasValidator validator.CustomAlias,
) ModifierPersist {
	return ModifierPersist{
		urlRepo:             urlRepo,
		userURLRelationRepo: userURLRelationRepo,
		aliasValidator:      aliasValidator,
	}
}
