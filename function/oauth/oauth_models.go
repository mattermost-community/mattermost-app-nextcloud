package oauth

type Connect struct {
	Poll struct {
		Token    string `json:"token"`
		Endpoint string `json:"endpoint"`
	} `json:"poll"`
	Login string `json:"login"`
}

type UserTokenV2 struct {
	Server      string `json:"server"`
	LoginName   string `json:"loginName"`
	AppPassword string `json:"appPassword"`
}

type Token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	UserID       string `json:"user_id"`
}

type RequestTokenBody struct {
	Code      string `json:"code"`
	GrantType string `json:"grant_type"`
}

type RefreshTokenBody struct {
	RefreshToken string `json:"refresh_token"`
	GrantType    string `json:"grant_type"`
}

type CreateWebhooksBody struct {
	Enabled                 bool   `json:"enabled"`
	CalendarEventCreatedURL string `json:"calendar_event_created_url"`
	CalendarEventUpdatedURL string `json:"calendar_event_updated_url"`
	WebhookSecret           string `json:"webhook_secret"`
}
