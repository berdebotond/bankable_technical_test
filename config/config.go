package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	DatabaseHost     string `json:"databaseHost" default:"localhost"`
	DatabasePort     string `json:"databasePort", default:"5432"`
	DatabaseUser     string `json:"databaseUser", default:"username"`
	DatabasePassword string `json:"databasePassword", default:"password"`
	DatabaseName     string `json:"databaseName", default:"test"`
	// Add more fields as needed
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile("./config/config.json") // Specify the configuration file path
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
