package oauth

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"testing"
)

func TestShouldGenerateUrl(t *testing.T) {

	creq := apps.CallRequest{}
	creq.Context.ExpandedContext.OAuth2.OAuth2App.RemoteRootURL = "http://localhost:8082"
	creq.Context.ExpandedContext.OAuth2.OAuth2App.ClientID = "CLIENT_ID"
	m := make(map[string]interface{})
	m["state"] = "state"
	creq.Values = m

	expected := "http://localhost:8082/index.php/apps/oauth2/authorize?response_type=code&state=state&client_id=CLIENT_ID"
	actual := buildConnectUrl(&creq)

	if expected != actual {
		t.Errorf(" expected %q, actual %q", expected, actual)
	}

}
