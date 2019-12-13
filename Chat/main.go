package main

import (
	"Chat/websocketchat"
	"github.com/joho/godotenv"
	"os"
)

func main() {
	if v,found := os.LookupEnv("ENV_INITIALIZED"); !found || v == "false" {
		err := godotenv.Load()
		if err != nil {
			websocketchat.Logger().Fatal(err.Error())
		}
	}

	port := os.Getenv("PORT")
	enableSLL := os.Getenv("ENABLE_SSL")
	apiEndpoint := os.Getenv("API_URL")
	authEndpoint := os.Getenv("AUTH_URL")
	websocketchat.GetFlankiChecker.SetEndpoints(apiEndpoint, authEndpoint)

	websocketchat.Logger().Fatal(websocketchat.NewChatServer().Run(":" + port, enableSLL))
}