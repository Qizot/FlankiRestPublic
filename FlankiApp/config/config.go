package config

import (
	"FlankiRest/logger"
	"github.com/joeshaw/envdecode"
	"github.com/joho/godotenv"
	"gopkg.in/oauth2.v3/models"
	"os"
)

type AppConfig struct {
	DBUsername string `env:"DB_USER,required"`
	DBPassword string `env:"DB_PASS,required"`
	DBName     string `env:"DB_NAME,required"`
	DBHost     string `env:"DB_HOST,required"`
	Port       string `env:"SERVER_PORT,default=8080"`
}

type Client struct {
	ID     string `env:"CLIENT_ID,required"`
	Secret string `env:"CLIENT_SECRET,required"`
	Domain string `env:"CLIENT_DOMAIN,required"`
	UserID string
}

type AuthServerConfig struct {
	Domain  string `env:"AUTHORIZATION_SERVER_DOMAIN,required"`
	Port    string `env:"AUTHORIZATION_SERVER_PORT,required"`
	Clients []models.Client
}

type ImageServerConfig struct {
	Domain string `env:"IMAGE_SERVER_DOMAIN,required"`
	Port   string `env:"IMAGE_SERVER_PORT,required"`
}

var API_PREFIX string

var appInstance *AppConfig
var authInstance *AuthServerConfig
var imgInstance *ImageServerConfig



func init() {
	if value, exists := os.LookupEnv("ENV_INITIALIZED"); !exists || value == "false" {
		err := godotenv.Load() //Load .env file
		if err != nil {
			logger.GetGlobalLogger().Fatal("Error while loading .env file: " + err.Error())
		}
		logger.GetGlobalLogger().Info("Using .env file")
	} else {
		logger.GetGlobalLogger().Info("Using environmental variables")
	}
	appInstance  = new(AppConfig)
	authInstance = new(AuthServerConfig)
	imgInstance  = new(ImageServerConfig)
	logger.GetGlobalLogger().Info("Initializing configs")
	appInstance.LoadFromEnvVariables()
	authInstance.LoadFromEnvVariables()
	imgInstance.LoadFromEnvVariables()

	API_PREFIX = ""
}

func GetAuthServerConfig() *AuthServerConfig {
	return authInstance
}

func GetAppConfig() *AppConfig {
	return appInstance
}

func GetImageServerConfig() *ImageServerConfig {
	return imgInstance
}


func (client *Client) GetOauthClient() models.Client {
	return models.Client{client.ID,client.Secret, client.Domain,""}
}

func (cfg *AuthServerConfig) LoadFromEnvVariables() {
	logEntry := logger.GetGlobalLogger().WithField("prefix","[CONFIG]")
	err := envdecode.Decode(cfg)
	if err != nil {
		logEntry.Fatal("Failed to load authorization server config: ", err.Error())
	}
	client := Client{}
	err = envdecode.Decode(&client)
	if err != nil {
		logEntry.Fatal("Failed to load app client oauth config: ", err.Error())
	}
	cfg.Clients = []models.Client{client.GetOauthClient()}
}

func (cfg *AppConfig) LoadFromEnvVariables() {
	err := envdecode.Decode(cfg)
	if err != nil {
		logger.GetGlobalLogger().WithField("prefix","[CONFIG]").Fatal("Failed to load application config: ", err.Error())
	}
}

func (cfg *ImageServerConfig) LoadFromEnvVariables() {
	err := envdecode.Decode(cfg)
	if err != nil {
		logger.GetGlobalLogger().WithField("prefix","[CONFIG]").Fatal("Failed to image server config: ", err.Error())
	}
}


