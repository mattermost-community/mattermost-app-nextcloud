package help

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/prokhorind/nextcloud/function/locales"
)

type HelpService interface {
	HandleHelpCommand(c *gin.Context)
}

type HelpServiceImpl struct {
	c       *gin.Context
	request apps.CallRequest
}

func (h HelpServiceImpl) getSingleHelpMessage(message string) string {
	locale := h.request.Context.ActingUser.Locale
	messageSource := locales.MessageSource{h.c, locale}
	return messageSource.GetMessage("help." + message)
}

func (h HelpServiceImpl) createHelpForSingleCommand(command string) string {
	locale := h.request.Context.ActingUser.Locale
	messageSource := locales.MessageSource{h.c, locale}
	description := messageSource.GetMessage(fmt.Sprintf("help.%s", command))
	return fmt.Sprintf("/nextcloud %s - %s", command, description)
}
