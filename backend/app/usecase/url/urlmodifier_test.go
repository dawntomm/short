package url

import (
	"short/app/entity"
	"short/app/usecase/repository"
	"short/app/usecase/validator"
	"testing"

	"github.com/byliuyang/app/mdtest"
)

func TestUrlModifier_UpdateURL(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		urls          urlMap
		user          entity.User
		oldAlias      string
		newAlias      string
		hasErr        bool
		expectedURL   entity.URL
		relationUsers []entity.User
		relationURLs  []entity.URL
	}{
		{
			name: "alias not found",
			urls: urlMap{},
			user: entity.User{
				Email: "alpha@example.com",
			},
			oldAlias:      "220uFicCJj",
			newAlias:      "3029jasdjnc",
			hasErr:        true,
			expectedURL:   entity.URL{},
			relationUsers: []entity.User{},
			relationURLs:  []entity.URL{},
		},
		{
			name: "new alias not valid",
			urls: urlMap{
				"220uFicCJj": entity.URL{
					Alias: "220uFicCJj",
				},
			},
			user: entity.User{
				Email: "alpha@example.com",
			},
			oldAlias:      "220uFicCJj",
			newAlias:      "3029jasdjncasdasldnlkahfkhsaklhfkwehkjasddasdhsdsakl",
			hasErr:        true,
			expectedURL:   entity.URL{},
			relationUsers: []entity.User{},
			relationURLs:  []entity.URL{},
		},
		{
			name: "success",
			urls: urlMap{
				"220uFicCJj": entity.URL{
					Alias: "220uFicCJj",
				},
			},
			user: entity.User{
				Email: "alpha@example.com",
			},
			oldAlias: "220uFicCJj",
			newAlias: "3029jasdjnc",
			hasErr:   false,
			expectedURL: entity.URL{
				Alias: "3029jasdjnc",
			},
			relationUsers: []entity.User{
				entity.User{
					Email: "alpha@example.com",
				},
			},
			relationURLs: []entity.URL{
				entity.URL{
					Alias: "220uFicCJj",
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			fakeRepo := repository.NewURLFake(testCase.urls)
			userURLRepo := repository.NewUserURLRepoFake(
				testCase.relationUsers,
				testCase.relationURLs,
			)
			aliasValidator := validator.NewCustomAlias()
			modifier := NewModifierPersist(&fakeRepo, &userURLRepo, aliasValidator)
			url, err := modifier.UpdateURL(&testCase.oldAlias, &testCase.newAlias, testCase.user)

			if testCase.hasErr {
				mdtest.NotEqual(t, nil, err)
				return
			}
			mdtest.Equal(t, nil, err)
			mdtest.Equal(t, testCase.expectedURL, url)
		})
	}
}
