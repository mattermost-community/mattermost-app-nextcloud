package main

import (
	"flag"
	"github.com/prokhorind/nextcloud/function/install"
	"github.com/prokhorind/nextcloud/http-server/config"
	"log"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/prokhorind/nextcloud/function"
)

func main() {
	r := gin.Default()

	configuration := getConfiguration()
	function.InitHandlers(r)
	r.Use(setAppConfig(configuration))

	r.GET("/manifest.json", install.GetManifest)
	r.GET("/static/icon.png", install.GetIcon)

	portStr := getPort(configuration)
	r.Run(":" + portStr)

}

func getConfiguration() config.Config {
	cpath := getEnv("CONFIGPATH", "config")
	flag.Parse()

	configuration, err := config.LoadConfig(cpath)
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	return configuration
}

func getPort(configuration config.Config) string {

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

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

func setAppConfig(conf config.Config) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		ctx.Set("config", conf)
	}
}
