package repository_order

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/Napat/golang-testcontainers-demo/pkg/model"
	"github.com/elastic/go-elasticsearch/v8"
)

const indexName = "orders"

type OrderRepository struct {
	client *elasticsearch.Client
}

func NewOrderRepository(client *elasticsearch.Client) *OrderRepository {
	return &OrderRepository{client: client}
}

func (r *OrderRepository) CreateOrder(ctx context.Context, order *model.Order) error {
	body, err := json.Marshal(order)
	if err != nil {
		return err
	}

	_, err = r.client.Index(
		indexName,
		bytes.NewReader(body),
		r.client.Index.WithDocumentID(order.ID),
		r.client.Index.WithContext(ctx),
	)
	return err
}

func (r *OrderRepository) SearchOrders(ctx context.Context, query map[string]interface{}) ([]model.Order, error) {
	// Marshal query map to JSON bytes
	queryBytes, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	res, err := r.client.Search(
		r.client.Search.WithIndex(indexName),
		r.client.Search.WithBody(bytes.NewReader(queryBytes)),
		r.client.Search.WithContext(ctx),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var result struct {
		Hits struct {
			Hits []struct {
				Source model.Order `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	orders := make([]model.Order, len(result.Hits.Hits))
	for i, hit := range result.Hits.Hits {
		orders[i] = hit.Source
	}

	return orders, nil
}
