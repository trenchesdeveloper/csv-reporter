package config

import (
	"github.com/spf13/viper"
)

type AppConfig struct {
	DBSOURCE                string `mapstructure:"DB_SOURCE"`
	DBDRIVER                string `mapstructure:"DB_DRIVER"`
	DB_SOURCE_TEST          string `mapstructure:"DB_SOURCE_TEST"`
	SERVER_PORT             string `mapstructure:"SERVER_PORT"`
	ENVIRONMENT             string `mapstructure:"ENVIRONMENT"`
	JWT_SECRET              string `mapstructure:"JWT_SECRET"`
	AppName                 string `mapstructure:"APP_NAME"`
	AWS_ACCESS_KEY_ID       string `mapstructure:"AWS_ACCESS_KEY_ID"`
	AWS_SECRET_ACCESS_KEY   string `mapstructure:"AWS_SECRET_ACCESS_KEY"`
	AWS_DEFAULT_REGION      string `mapstructure:"AWS_DEFAULT_REGION"`
	SQS_QUEUE               string `mapstructure:"SQS_QUEUE"`
	S3_BUCKET               string `mapstructure:"S3_BUCKET"`
	S3_LOCALSTACK_ENDPOINT  string `mapstructure:"S3_LOCALSTACK_ENDPOINT"`
	SQS_LOCALSTACK_ENDPOINT string `mapstructure:"SQS_LOCALSTACK_ENDPOINT"`

	TF_VAR_aws_access_key_id       string `mapstructure:"TF_VAR_aws_access_key_id"`
	TF_VAR_aws_secret_access_key   string `mapstructure:"TF_VAR_aws_secret_access_key"`
	TF_VAR_aws_default_region      string `mapstructure:"TF_VAR_aws_default_region"`
	TF_VAR_sqs_queue               string `mapstructure:"TF_VAR_sqs_queue"`
	TF_VAR_s3_bucket               string `mapstructure:"TF_VAR_s3_bucket"`
	TF_VAR_s3_localstack_endpoint  string `mapstructure:"TF_VAR_s3_localstack_endpoint"`
	TF_VAR_sqs_localstack_endpoint string `mapstructure:"TF_VAR_sqs_localstack_endpoint"`
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
	viper.BindEnv("AWS_ACCESS_KEY_ID", "AWS_ACCESS_KEY_ID")
	viper.BindEnv("AWS_SECRET_ACCESS_KEY", "AWS_SECRET_ACCESS_KEY")
	viper.BindEnv("AWS_DEFAULT_REGION", "AWS_DEFAULT_REGION")
	viper.BindEnv("SQS_QUEUE", "SQS_QUEUE")
	viper.BindEnv("S3_BUCKET", "S3_BUCKET")
	viper.BindEnv("TF_VAR_aws_access_key_id", "TF_VAR_aws_access_key_id")
	viper.BindEnv("TF_VAR_aws_secret_access_key", "TF_VAR_aws_secret_access_key")
	viper.BindEnv("TF_VAR_aws_default_region", "TF_VAR_aws_default_region")
	viper.BindEnv("TF_VAR_sqs_queue", "TF_VAR_sqs_queue")
	viper.BindEnv("TF_VAR_s3_bucket", "TF_VAR_s3_bucket")
	viper.BindEnv("S3_LOCALSTACK_ENDPOINT", "S3_LOCALSTACK_ENDPOINT")
	viper.BindEnv("SQS_LOCALSTACK_ENDPOINT", "SQS_LOCALSTACK_ENDPOINT")

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
