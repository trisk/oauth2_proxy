package providers

import (
	"log"
	"net/http"
	"net/url"

	"github.com/bitly/oauth2_proxy/api"
)

type SlackProvider struct {
	*ProviderData
	Team string
}

func NewSlackProvider(p *ProviderData) *SlackProvider {
	p.ProviderName = "Slack"
	if p.LoginURL == nil || p.LoginURL.String() == "" {
		p.LoginURL = &url.URL{
			Scheme: "https",
			Host:   "slack.com",
			Path:   "/oauth/authorize",
		}
	}
	if p.RedeemURL == nil || p.RedeemURL.String() == "" {
		p.RedeemURL = &url.URL{
			Scheme: "https",
			Host:   "slack.com",
			Path:   "/api/oauth.access",
		}
	}
	if p.ValidateURL == nil || p.ValidateURL.String() == "" {
		p.ValidateURL = &url.URL{
			Scheme: "https",
			Host:   "slack.com",
			Path:   "/api/users.identity",
		}
	}
	if p.Scope == "" {
		p.Scope = "identity.basic,identity.email"
	}
	return &SlackProvider{ProviderData: p}
}

func (p *SlackProvider) Configure(team string) {
	p.Team = team
}

// GetLoginURL with Slack-accepted oauth parameters
func (p *SlackProvider) GetLoginURL(redirectURI, state string) string {
	var a url.URL
	a = *p.LoginURL
	params, _ := url.ParseQuery(a.RawQuery)
	params.Set("redirect_uri", redirectURI)
	if p.Team != "" {
		params.Add("team", p.Team)
	}
	params.Add("scope", p.Scope)
	params.Set("client_id", p.ClientID)
	params.Set("response_type", "code")
	params.Add("state", state)
	a.RawQuery = params.Encode()
	return a.String()
}

func (p *SlackProvider) GetEmailAddress(s *SessionState) (string, error) {

	req, err := http.NewRequest("GET",
		p.ValidateURL.String()+"?token="+s.AccessToken, nil)
	if err != nil {
		log.Printf("failed building request %s", err)
		return "", err
	}
	json, err := api.Request(req)
	if err != nil {
		log.Printf("failed making request %s", err)
		return "", err
	}
	if email, ok := json.Get("user").CheckGet("email"); ok {
		return email.String()
	}
	return json.Get("user").Get("id").String()
}
