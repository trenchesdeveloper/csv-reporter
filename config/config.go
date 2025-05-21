package config

import (
	"github.com/spf13/viper"
)

type AppConfig struct {
	DBSOURCE       string `mapstructure:"DB_SOURCE"`
	DBDRIVER       string `mapstructure:"DB_DRIVER"`
	DB_SOURCE_TEST string `mapstructure:"DB_SOURCE_TEST"`
	SERVER_PORT    string `mapstructure:"SERVER_PORT"`
	ENVIRONMENT    string `mapstructure:"ENVIRONMENT"`
	JWT_SECRET     string `mapstructure:"JWT_SECRET"`
	AppName        string `mapstructure:"APP_NAME"`
}

func LoadConfig(path string) (*AppConfig, error) {
	// Always load environment variables from the environment
	viper.AutomaticEnv()

	// bind environment variables
	viper.BindEnv("DB_SOURCE", "DB_SOURCE")
	viper.BindEnv("HTTP_PORT", "HTTP_PORT")
	viper.BindEnv("DB_DRIVER", "DB_DRIVER")
	viper.BindEnv("DB_SOURCE_TEST", "DB_SOURCE_TEST")
	viper.BindEnv("SERVER_PORT", "SERVER_PORT")
	viper.BindEnv("ENVIRONMENT", "ENVIRONMENT")
	viper.BindEnv("JWT_SECRET", "JWT_SECRET")
	viper.BindEnv("APP_NAME", "APP_NAME")

	// Check if the environment is set to production
	if viper.GetString("ENVIRONMENT") != "production" {
		viper.AddConfigPath(path)
		viper.SetConfigName("app")
		viper.SetConfigType("env")

		err := viper.ReadInConfig()
		if err != nil {
			return nil, err
		}
	}

	var config AppConfig
	err := viper.Unmarshal(&config)

	if err != nil {
		return nil, err
	}

	return &config, nil
}
