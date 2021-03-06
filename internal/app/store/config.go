package store

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

type ItemKey struct {
	Brand   string `dynamodbav:"brand"`
	Article string `dynamodbav:"id"`
}

type ItemData struct {
	Brand        string  `dynamodbav:"brand"`
	Article      string  `dynamodbav:"id"`
	Label        string  `dynamodbav:"item_name"`
	Description  string  `dynamodbav:"item_description"`
	Manufacturer string  `dynamodbav:"brand_description"`
	Price        float64 `dynamodbav:"price"`
	Available    bool    `dynamodbav:"available"`
	CreatedAt    string  `dynamodbav:"created_at"`
}

type Config struct {
	aws   aws.Config
	table *string
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
		aws:   awsConfig,
		table: aws.String("fishersci_items"),
	}
}
