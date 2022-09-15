package user

import "github.com/mattermost/mattermost-plugin-apps/apps/appclient"

const UserSettingsKvKey = "user-settings-"

type UserSettingsServiceImpl struct {
	AsBot *appclient.Client
}

func (userSettings *UserSettingsServiceImpl) GetUserSettingsById(id string) UserSettings {
	us := UserSettings{}
	userSettings.AsBot.KVGet("", UserSettingsKvKey+id, &us)
	return us
}

func (userSettings *UserSettingsServiceImpl) SetUserSettingsById(id string, us UserSettings) UserSettings {
	userSettings.AsBot.KVSet("", UserSettingsKvKey+id, us)
	return us
}
