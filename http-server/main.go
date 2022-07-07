package main

import (
	"log"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/prokhorind/nextcloud/config"
	"github.com/prokhorind/nextcloud/function"
)

func main() {
	r := gin.Default()

	function.InitHandlers(r)

	portStr := getPort()
	r.Run(":" + portStr)

}

func getPort() string {
	configuration, err := config.LoadConfig("../config")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	portStr := configuration.Port

	if portStr == "" {
		u, err := url.Parse(configuration.APPURL)
		if err != nil {
			panic(err)
		}
		portStr = u.Port()
	}

	return portStr
}
