package store

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	// "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Client struct {
	aws *dynamodb.Client
}

func NewClient(config *Config) *Client {
	return &Client{
		aws: dynamodb.NewFromConfig(config.aws),
	}
}

func (client *Client) PrintDataTables(config *Config) {
	outPaginator := dynamodb.NewListTablesPaginator(client.aws, nil,
		func(options *dynamodb.ListTablesPaginatorOptions) {
			options.StopOnDuplicateToken = true
		})
	for outPaginator.HasMorePages() {
		outList, err := outPaginator.NextPage(context.TODO())
		if err != nil {
			log.Fatal(err)
		}
		for _, outTable := range outList.TableNames {
			fmt.Printf("%s\t", outTable)
		}
	}
}
