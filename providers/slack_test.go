package providers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testSlackProvider(hostname string) *SlackProvider {
	p := NewSlackProvider(
		&ProviderData{
			ProviderName: "",
			LoginURL:     &url.URL{},
			RedeemURL:    &url.URL{},
			ProfileURL:   &url.URL{},
			ValidateURL:  &url.URL{},
			Scope:        ""})
	if hostname != "" {
		updateURL(p.Data().LoginURL, hostname)
		updateURL(p.Data().RedeemURL, hostname)
		updateURL(p.Data().ProfileURL, hostname)
		updateURL(p.Data().ValidateURL, hostname)
	}
	return p
}

func testSlackBackend(payload string) *httptest.Server {
	path := "/api/users.identity"
	query := "token=imaginary_access_token"

	return httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			url := r.URL
			if url.Path != path || url.RawQuery != query {
				w.WriteHeader(404)
			} else {
				w.WriteHeader(200)
				w.Write([]byte(payload))
			}
		}))
}

func TestSlackProviderDefaults(t *testing.T) {
	p := testSlackProvider("")
	assert.NotEqual(t, nil, p)
	p.Configure("")
	assert.Equal(t, "Slack", p.Data().ProviderName)
	assert.Equal(t, "", p.Team)
	assert.Equal(t, "https://slack.com/oauth/authorize",
		p.Data().LoginURL.String())
	assert.Equal(t, "https://slack.com/api/oauth.access",
		p.Data().RedeemURL.String())
	assert.Equal(t, "https://slack.com/api/users.identity",
		p.Data().ValidateURL.String())
	assert.Equal(t, "identity.basic,identity.email", p.Data().Scope)
}

func TestSlackProviderOverrides(t *testing.T) {
	p := NewSlackProvider(
		&ProviderData{
			LoginURL: &url.URL{
				Scheme: "https",
				Host:   "example.com",
				Path:   "/oauth/auth"},
			RedeemURL: &url.URL{
				Scheme: "https",
				Host:   "example.com",
				Path:   "/api/oauth.access"},
			ValidateURL: &url.URL{
				Scheme: "https",
				Host:   "example.com",
				Path:   "/api/users.identity"},
			Scope: "identity.basic"})
	assert.NotEqual(t, nil, p)
	assert.Equal(t, "Slack", p.Data().ProviderName)
	assert.Equal(t, "https://example.com/oauth/auth",
		p.Data().LoginURL.String())
	assert.Equal(t, "https://example.com/api/oauth.access",
		p.Data().RedeemURL.String())
	assert.Equal(t, "https://example.com/api/users.identity",
		p.Data().ValidateURL.String())
	assert.Equal(t, "identity.basic", p.Data().Scope)
}

func TestSlackProviderSetTeam(t *testing.T) {
	p := testSlackProvider("")
	assert.NotEqual(t, nil, p)
	p.Configure("example")
	assert.Equal(t, "Slack", p.Data().ProviderName)
	assert.Equal(t, "example", p.Team)
	assert.Equal(t, "https://slack.com/oauth/authorize",
		p.Data().LoginURL.String())
	assert.Equal(t, "https://slack.com/api/oauth.access",
		p.Data().RedeemURL.String())
	assert.Equal(t, "https://slack.com/api/users.identity",
		p.Data().ValidateURL.String())
	assert.Equal(t, "identity.basic,identity.email", p.Data().Scope)
}

func TestSlackProviderGetEmailAddress(t *testing.T) {
	b := testSlackBackend("{\"user\": {\"id\": \"XYZZY\", " +
		"\"email\": \"michael.bland@gsa.gov\"}}")
	defer b.Close()

	b_url, _ := url.Parse(b.URL)
	p := testSlackProvider(b_url.Host)

	session := &SessionState{AccessToken: "imaginary_access_token"}
	email, err := p.GetEmailAddress(session)
	assert.Equal(t, nil, err)
	assert.Equal(t, "michael.bland@gsa.gov", email)
}

// Note that trying to trigger the "failed building request" case is not
// practical, since the only way it can fail is if the URL fails to parse.
func TestSlackProviderGetEmailAddressFailedRequest(t *testing.T) {
	b := testSlackBackend("unused payload")
	defer b.Close()

	b_url, _ := url.Parse(b.URL)
	p := testSlackProvider(b_url.Host)

	// We'll trigger a request failure by using an unexpected access
	// token. Alternatively, we could allow the parsing of the payload as
	// JSON to fail.
	session := &SessionState{AccessToken: "unexpected_access_token"}
	email, err := p.GetEmailAddress(session)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, "", email)
}

func TestSlackProviderGetEmailAddressEmailNotPresentInPayload(t *testing.T) {
	b := testSlackBackend("{\"user\": {\"id\": \"bar\"}}")
	defer b.Close()

	b_url, _ := url.Parse(b.URL)
	p := testSlackProvider(b_url.Host)

	session := &SessionState{AccessToken: "imaginary_access_token"}
	email, err := p.GetEmailAddress(session)
	assert.Equal(t, nil, err)
	assert.Equal(t, "bar", email)
}
