package user

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
)

type BotService interface {
	AddBot()
}

type BotServiceImpl struct {
	Creq apps.CallRequest
}

func (s BotServiceImpl) AddBot() {
	s.addBotToTeam()
	s.addBotToChannel()
}

func (s BotServiceImpl) addBotToTeam() {
	teamId := s.Creq.Context.Channel.TeamId
	botId := s.Creq.Context.BotUserID
	asActingUser := appclient.AsActingUser(s.Creq.Context)
	_, _, err := asActingUser.GetTeamMember(teamId, botId, "")

	if err != nil {
		asActingUser.AddTeamMember(teamId, botId)
	}
}

func (s BotServiceImpl) addBotToChannel() {
	channelId := s.Creq.Context.Channel.Id
	botId := s.Creq.Context.BotUserID

	asActingUser := appclient.AsActingUser(s.Creq.Context)

	_, _, err := asActingUser.GetChannelMember(channelId, botId, "")

	if err != nil {
		asActingUser.AddChannelMember(channelId, botId)
	}

}
