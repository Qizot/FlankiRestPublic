package main

import (
	"FlankiRest/app"
	"FlankiRest/config"
	"FlankiRest/logger"
	"golang.org/x/oauth2"
	"log"
	"os"
)


func main() {

	appConfig := config.GetAppConfig()
	authConfig := config.GetAuthServerConfig()
	cfg := &oauth2.Config{
		ClientID:     "8245d6e94963dcf75fd285958721341e",
		ClientSecret: "e9943005b6c456250d3b771783b941af",
		Scopes:       []string{"all"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  authConfig.Domain + ":" + authConfig.Port + "/authorize",
			TokenURL: authConfig.Domain + ":" +  authConfig.Port + "/token",
		},
	}

	application := app.NewApp(cfg)
	application.SetLogger(logger.GetGlobalLogger())
	application.Initialize(appConfig)

	enableSSL := false
	if val, ok := os.LookupEnv("ENABLE_SSL"); ok {
		if val == "true" {
			enableSSL = true
		}
	}
	log.Fatal(application.Run(enableSSL))			// application server
}