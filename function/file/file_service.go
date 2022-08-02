package file

import (
	"bytes"
	"fmt"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"net/http"
)

type FileServiceImpl struct {
	Url   string
	Token string
}

func (s FileServiceImpl) UploadFile(file []byte) {
	req, _ := http.NewRequest("PUT", s.Url, bytes.NewBuffer(file))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.Token))

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()
}

func AddBot(creq apps.CallRequest) {
	addBotToTeam(creq)
	addBotToChannel(creq)
}

func addBotToTeam(creq apps.CallRequest) {
	teamId := creq.Context.Channel.TeamId
	botId := creq.Context.BotUserID
	asActingUser := appclient.AsActingUser(creq.Context)
	_, _, err := asActingUser.GetTeamMember(teamId, botId, "")

	if err != nil {
		asActingUser.AddTeamMember(teamId, botId)
	}
}

func addBotToChannel(creq apps.CallRequest) {
	channelId := creq.Context.Channel.Id
	botId := creq.Context.BotUserID

	asActingUser := appclient.AsActingUser(creq.Context)

	_, _, err := asActingUser.GetChannelMember(channelId, botId, "")

	if err != nil {
		asActingUser.AddChannelMember(channelId, botId)
	}

}
