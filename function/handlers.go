package function

import (
	"github.com/gin-gonic/gin"
	"github.com/prokhorind/nextcloud/function/calendar"
	"github.com/prokhorind/nextcloud/function/install"
	"github.com/prokhorind/nextcloud/function/oauth"
	"github.com/prokhorind/nextcloud/function/search"
)

func InitHandlers(r *gin.Engine) {
	r.GET("/manifest.json", install.GetManifest)
	r.GET("/static/icon.png", install.GetIcon)
	r.POST("/bindings", install.Bindings)
	r.POST("/configure", oauth.Configure)
	r.POST("/connect", oauth.HandleConnect)
	r.POST("/disconnect", oauth.Disconnect)
	r.POST("/oauth2/complete", oauth.Oauth2Complete)
	r.POST("/oauth2/connect", oauth.Oauth2Connect)
	r.POST("/file/search", search.FileSearch)
	r.POST("send", search.FileSearch)
	r.POST("/create-calendar-event", calendar.HandleCreateEvent)
	r.POST("/create-calendar-event-form", calendar.HandleCreateEventForm)
}
