package authorization

import (
	"AuthorizationServer/utils"
	"github.com/joeshaw/envdecode"
	"github.com/joho/godotenv"
	"gopkg.in/oauth2.v3/models"
	"os"
)

type AuthServerConfig struct {
	Domain  string `env:"AUTHORIZATION_SERVER_DOMAIN,required"`
	Port    string `env:"AUTHORIZATION_SERVER_PORT,required"`
	DBUsername string `env:"DB_USER,required"`
	DBPassword string `env:"DB_PASS,required"`
	DBName     string `env:"DB_NAME,required"`
	DBHost     string `env:"DB_HOST,required"`
	Clients []models.Client
}

type Client struct {
	ID     string `env:"CLIENT_ID,required"`
	Secret string `env:"CLIENT_SECRET,required"`
	Domain string `env:"CLIENT_DOMAIN,required"`
	UserID string
}

func (client *Client) GetOauthClient() models.Client {
	return models.Client{client.ID,client.Secret, client.Domain,""}
}

var authInstance *AuthServerConfig

func init() {
	entry := utils.AuthLoggerEntry()
	if value, exists := os.LookupEnv("ENV_INITIALIZED"); !exists || value == "false" {
		err := godotenv.Load() //Load .env file
		if err != nil {
			entry.Fatal("Error while loading .env file: " + err.Error())
		}
		entry.Info("Using .env file")
	} else {
		entry.Info("Using environmental variables")
	}

	authInstance = new(AuthServerConfig)

	entry.Info("Initializing authentication server config")
	authInstance.LoadFromEnvVariables()

}

func (cfg *AuthServerConfig) LoadFromEnvVariables() {
	entry := utils.AuthLogger()
	err := envdecode.Decode(cfg)
	if err != nil {
		entry.Fatal("Failed to load authorization server config: ", err.Error())
	}
	client := Client{}
	err = envdecode.Decode(&client)
	if err != nil {
		entry.Fatal("Failed to load app client oauth config: ", err.Error())
	}
	cfg.Clients = []models.Client{client.GetOauthClient()}
}

func GetAuthServerConfig() *AuthServerConfig {
	return authInstance
}