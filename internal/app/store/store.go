package store

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Client struct {
	aws   *dynamodb.Client
	table *string
}

func NewClient(config *Config) *Client {
	return &Client{
		aws:   dynamodb.NewFromConfig(config.aws),
		table: config.table,
	}
}

func (client *Client) PutItem(item *ItemData) error {
	marshalItem, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}
	if _, err = client.aws.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: client.table,
		Item:      marshalItem,
	}); err != nil {
		return err
	}
	return nil
}

func (client *Client) GetItem(keyValues *ItemKey) (*ItemData, error) {
	marshalKeys, err := attributevalue.MarshalMap(keyValues)
	if err != nil {
		return nil, err
	}
	marshalItem, err := client.aws.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: client.table,
		Key:       marshalKeys,
	})
	if err != nil {
		return nil, err
	}
	item := &ItemData{}
	if err = attributevalue.UnmarshalMap(marshalItem.Item, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (client *Client) DeleteItem(keyValues *ItemKey) error {
	marshalKeys, err := attributevalue.MarshalMap(keyValues)
	if err != nil {
		return err
	}
	if _, err = client.aws.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: client.table,
		Key:       marshalKeys,
	}); err != nil {
		return err
	}
	return nil
}

func (client *Client) WriteBatch(items []*ItemData) error {
	requests := make([]types.WriteRequest, 0)
	for _, item := range items {
		marshalItem, err := attributevalue.MarshalMap(item)
		if err != nil {
			return err
		}
		requests = append(requests, types.WriteRequest{
			PutRequest: &types.PutRequest{
				Item: marshalItem,
			},
		})
	}
	if _, err := client.aws.BatchWriteItem(context.TODO(), &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			*client.table: requests,
		},
	},
	); err != nil {
		return err
	}
	return nil
}

func (client *Client) PrintDataTables() {
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
