package store

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

type Config struct {
	aws aws.Config
}

func NewConfig() *Config {
	awsConfig, err := config.LoadDefaultConfig(context.TODO(),
		func(options *config.LoadOptions) error {
			options.Region = "us-east-1"
			return nil
		})
	if err != nil {
		log.Fatal(err)
	}

	return &Config{
		aws: awsConfig,
	}
}
