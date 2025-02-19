package repository_order

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Napat/golang-testcontainers-demo/pkg/metrics"
	"github.com/Napat/golang-testcontainers-demo/pkg/model"
	"github.com/elastic/go-elasticsearch/v8"
)

const indexName = "orders"

type OrderRepository struct {
	client  *elasticsearch.Client
	metrics *metrics.SearchMetrics
}

func NewOrderRepository(client *elasticsearch.Client) *OrderRepository {
	return &OrderRepository{
		client:  client,
		metrics: metrics.NewSearchMetrics(),
	}
}

func (r *OrderRepository) CreateOrder(ctx context.Context, order *model.Order) error {
	timer := time.Now()
	defer func() {
		r.metrics.SearchDuration.WithLabelValues("orders").Observe(time.Since(timer).Seconds())
	}()

	body, err := json.Marshal(order)
	if err != nil {
		r.metrics.SearchesTotal.WithLabelValues("orders", "error").Inc()
		return err
	}

	_, err = r.client.Index(
		indexName,
		bytes.NewReader(body),
		r.client.Index.WithDocumentID(order.ID),
		r.client.Index.WithContext(ctx),
		r.client.Index.WithRefresh("true"), // Force immediate refresh
	)

	if err != nil {
		r.metrics.SearchesTotal.WithLabelValues("orders", "error").Inc()
		return err
	}

	r.metrics.SearchesTotal.WithLabelValues("orders", "success").Inc()
	return nil
}

func (r *OrderRepository) SearchOrders(ctx context.Context, params map[string]interface{}) ([]model.Order, error) {
	timer := time.Now()
	defer func() {
		r.metrics.SearchDuration.WithLabelValues("orders").Observe(time.Since(timer).Seconds())
	}()

	var buf bytes.Buffer
	var searchQuery map[string]interface{}

	if query, ok := params["query"].(string); ok {
		// Simple text search
		searchQuery = map[string]interface{}{
			"query": map[string]interface{}{
				"bool": map[string]interface{}{
					"should": []map[string]interface{}{
						{
							"term": map[string]interface{}{
								"customer_id": query,
							},
						},
						{
							"term": map[string]interface{}{
								"status": query,
							},
						},
						{
							"term": map[string]interface{}{
								"payment_method": query,
							},
						},
						{
							"nested": map[string]interface{}{
								"path": "items",
								"query": map[string]interface{}{
									"bool": map[string]interface{}{
										"should": []map[string]interface{}{
											{
												"term": map[string]interface{}{
													"items.product_id": query,
												},
											},
											{
												"match": map[string]interface{}{
													"items.product_name": query,
												},
											},
										},
									},
								},
								"score_mode": "max",
							},
						},
					},
					"minimum_should_match": 1,
				},
			},
		}
	} else if customerID, ok := params["customer_id"].(string); ok {
		// Exact customer ID match
		searchQuery = map[string]interface{}{
			"query": map[string]interface{}{
				"term": map[string]interface{}{
					"customer_id": customerID,
				},
			},
		}
	} else if rawQuery, ok := params["query"].(map[string]interface{}); ok {
		// Direct query (used by ListOrders)
		searchQuery = map[string]interface{}{
			"query": rawQuery,
		}
	} else {
		// Default to match all
		searchQuery = map[string]interface{}{
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
		}
	}

	// // Add debug logging for query
	// if queryJSON, err := json.Marshal(searchQuery); err == nil {
	// 	fmt.Printf("Search query: %s\n", string(queryJSON))
	// }

	if err := json.NewEncoder(&buf).Encode(searchQuery); err != nil {
		r.metrics.SearchesTotal.WithLabelValues("orders", "error").Inc()
		return nil, fmt.Errorf("error encoding query: %w", err)
	}

	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex(indexName),
		r.client.Search.WithBody(&buf),
		// r.client.Search.WithPretty(), // Add pretty printing for debugging
	)
	if err != nil {
		r.metrics.SearchesTotal.WithLabelValues("orders", "error").Inc()
		return nil, fmt.Errorf("error getting response: %w", err)
	}
	defer res.Body.Close()

	// // Debug: Print raw response
	// if debugResp, err := io.ReadAll(res.Body); err == nil {
	// 	fmt.Printf("Raw Elasticsearch response: %s\n", string(debugResp))
	// 	// Re-create reader for further processing
	// 	res.Body = io.NopCloser(bytes.NewReader(debugResp))
	// }

	var response map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		r.metrics.SearchesTotal.WithLabelValues("orders", "error").Inc()
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	// // Debug: Print total hits
	// if hits, ok := response["hits"].(map[string]interface{}); ok {
	// 	if total, ok := hits["total"].(map[string]interface{}); ok {
	// 		fmt.Printf("Total hits: %v\n", total)
	// 	}
	// }

	var orders []model.Order
	hits, _ := response["hits"].(map[string]interface{})
	if hits != nil {
		if total, ok := hits["total"].(map[string]interface{}); ok {
			if value, ok := total["value"].(float64); ok {
				r.metrics.ResultsReturned.WithLabelValues("orders").Observe(value)
			}
		}
		if hitList, ok := hits["hits"].([]interface{}); ok {
			for _, hit := range hitList {
				hitMap, ok := hit.(map[string]interface{})
				if !ok {
					continue
				}
				source, ok := hitMap["_source"].(map[string]interface{})
				if !ok {
					continue
				}
				var order model.Order
				sourceBytes, err := json.Marshal(source)
				if err != nil {
					continue
				}
				if err := json.Unmarshal(sourceBytes, &order); err != nil {
					continue
				}
				orders = append(orders, order)
			}
		}
	}

	r.metrics.SearchesTotal.WithLabelValues("orders", "success").Inc()
	return orders, nil
}
