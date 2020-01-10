// +build integration all

package google

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/short-d/app/mdtest"
	"github.com/short-d/short/app/entity"
)

func TestAccount_GetSingleSignOnUser(t *testing.T) {
	testCases := []struct {
		name            string
		httpResponse    *http.Response
		httpErr         error
		expectHasErr    bool
		expectedSSOUser entity.SSOUser
	}{
		{
			name:         "invalid access token",
			httpResponse: nil,
			httpErr:      errors.New("invalid access token"),
			expectHasErr: true,
		},
		{
			name: "user has id, email, and name",
			httpResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body: ioutil.NopCloser(bytes.NewReader([]byte(`
{
      "sub": "bcBi3AMeOV3Zg3AlOPyn",
      "name": "Google User",
      "email": "googleUser@gmail.com"
}
`,
				)))},
			expectHasErr: false,
			expectedSSOUser: entity.SSOUser{
				ID:    "bcBi3AMeOV3Zg3AlOPyn",
				Name:  "Google User",
				Email: "googleUser@gmail.com",
			},
		},
		{
			name: "user doesn't have email",
			httpResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body: ioutil.NopCloser(bytes.NewReader([]byte(`
{
      "sub": "bcBi3AMeOV3Zg3AlOPyn",
      "name": "Google User",
      "email": ""
}
`,
				)))},
			expectHasErr: false,
			expectedSSOUser: entity.SSOUser{
				ID:    "bcBi3AMeOV3Zg3AlOPyn",
				Name:  "Google User",
				Email: "",
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			httpRequest := mdtest.NewHTTPRequestFake(
				func(req *http.Request) (response *http.Response, e error) {
					mdtest.Equal(t, "https://openidconnect.googleapis.com/v1/userinfo", req.URL.String())
					mdtest.Equal(t, "GET", req.Method)
					mdtest.Equal(t, "application/json", req.Header.Get("Accept"))
					mdtest.Equal(t, "Bearer access_token", req.Header.Get("Authorization"))

					return testCase.httpResponse, testCase.httpErr
				})
			googleAccount := NewAccount(httpRequest)

			gotSSOUser, err := googleAccount.GetSingleSignOnUser("access_token")

			if testCase.expectHasErr {
				mdtest.NotEqual(t, nil, err)
				return
			}
			mdtest.Equal(t, nil, err)
			mdtest.Equal(t, testCase.expectedSSOUser, gotSSOUser)
		})
	}
}
