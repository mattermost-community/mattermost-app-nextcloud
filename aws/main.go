package aws

import (
	"context"
	"flag"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	"github.com/prokhorind/nextcloud/config"
	"github.com/prokhorind/nextcloud/function"
	"log"
	"os"
)

var ginLambda *ginadapter.GinLambda

func init() {
	r := gin.Default()

	configuration := getConfiguration()
	configuration.APPTYPE = "AWS"
	function.InitHandlers(r, configuration)
	ginLambda = ginadapter.New(r)
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// If no name is provided in the HTTP request body, throw an error
	return ginLambda.ProxyWithContext(ctx, req)
}

func getConfiguration() config.Config {
	cpath := getEnv("CONFIGPATH", "../config")
	flag.Parse()

	configuration, err := config.LoadConfig(cpath)
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	return configuration
}

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

func main() {
	lambda.Start(Handler)
}
