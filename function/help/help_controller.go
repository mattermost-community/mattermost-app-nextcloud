package help

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"net/http"
	"strings"
)

func HandleHelpCommand(c *gin.Context) {
	creq := apps.CallRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)
	helpService := HelpServiceImpl{c: c, request: creq}
	builder := strings.Builder{}
	builder.WriteString(helpService.getSingleHelpMessage("title"))
	builder.WriteString("\n")
	if creq.Context.ActingUser.IsSystemAdmin() {
		builder.WriteString(helpService.createHelpForSingleCommand("configure"))
	}
	builder.WriteString("\n")
	builder.WriteString(helpService.createHelpForSingleCommand("connect"))
	builder.WriteString("\n")
	builder.WriteString(helpService.createHelpForSingleCommand("share"))
	builder.WriteString("\n")
	builder.WriteString(helpService.createHelpForSingleCommand("calendars"))
	builder.WriteString("\n")
	builder.WriteString(helpService.createHelpForSingleCommand("disconnect"))
	builder.WriteString("\n")
	builder.WriteString("\n")
	builder.WriteString(helpService.getSingleHelpMessage("tips"))
	c.JSON(http.StatusOK, apps.NewTextResponse(builder.String()))
}
