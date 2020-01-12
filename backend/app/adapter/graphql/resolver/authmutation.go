package resolver

import (
	"errors"
	"fmt"
	"short/app/entity"
	"short/app/usecase/url"
	"time"
)

// AuthMutation represents GraphQL mutation resolver that acts differently based
// on the identify of the user
type AuthMutation struct {
	user       *entity.User
	urlCreator url.Creator
	urlModifier url.Modifier
}

// URLInput represents possible URL attributes
type URLInput struct {
	OriginalURL string
	CustomAlias *string
	ExpireAt    *time.Time
}

// CreateURLArgs represents the possible parameters for CreateURL endpoint
type CreateURLArgs struct {
	URL      URLInput
	IsPublic bool
}

// ModifyURLArgs represents the possible parameters for ModifyURL endpoint
type ModifyURLArgs struct {
	OldAlias      string
	NewAlias      string
}

// CreateURL creates mapping between an alias and a long link for a given user
func (a AuthMutation) CreateURL(args *CreateURLArgs) (*URL, error) {
	if a.user == nil {
		return nil, errors.New("unauthorized request")
	}

	customAlias := args.URL.CustomAlias
	u := entity.URL{
		OriginalURL: args.URL.OriginalURL,
		ExpireAt:    args.URL.ExpireAt,
	}

	isPublic := args.IsPublic

	newURL, err := a.urlCreator.CreateURL(u, customAlias, *a.user, isPublic)
	if err == nil {
		return &URL{url: newURL}, nil
	}
	switch err.(type) {
	case url.ErrAliasExist:
		return nil, ErrURLAliasExist(*customAlias)
	case url.ErrInvalidLongLink:
		return nil, ErrInvalidLongLink(u.OriginalURL)
	case url.ErrInvalidCustomAlias:
		return nil, ErrInvalidCustomAlias(*customAlias)
	default:
		return nil, ErrUnknown{}
	}
}

// ModifyURL updates mapping between an alias and a long link for a given user
func (a AuthMutation) ModifyURL(args *ModifyURLArgs) (*URL, error) {
	if a.user == nil {
		return nil, errors.New("unauthorized request")
	}

	newURL, err := a.urlModifier.UpdateURL(&args.OldAlias, &args.NewAlias, *a.user)
	if err == nil {
		return &URL{url: newURL}, nil
	}
	fmt.Println(err)
	switch err.(type) {
	case url.ErrAliasExist:
		return nil, ErrURLAliasExist(args.NewAlias)
	case url.ErrInvalidCustomAlias:
		return nil, ErrInvalidCustomAlias(args.NewAlias)
	default:
		return nil, ErrUnknown{}
	}
}

func newAuthMutation(user *entity.User, urlCreator url.Creator, urlModifier url.Modifier) AuthMutation {
	return AuthMutation{
		user:       user,
		urlCreator: urlCreator,
		urlModifier: urlModifier,
	}
}
