package providers

import (
	"net/url"
)

type ProviderData struct {
	ProviderName      string
	ClientID          string
	ClientSecret      string
	LoginURL          *url.URL
	RedeemURL         *url.URL
	ProfileURL        *url.URL
	ProtectedResource *url.URL
	ValidateURL       *url.URL
	Scope             string
	Prompt            string
}

func (p *ProviderData) Data() *ProviderData { return p }
