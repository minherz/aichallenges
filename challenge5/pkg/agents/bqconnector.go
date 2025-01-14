package agents

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	"github.com/minherz/aichallenges/challenge1/pkg/utils"
	"google.golang.org/api/iterator"
)

type BQConnector struct {
	client *bigquery.Client
}

type HotelRecord struct {
	Name            string `json:"name"`
	Address         string `json:"address"`
	Description     string `json:"description"`
	NearAttractions string `json:"attractions"`
}

func NewBigQueryConnector(ctx context.Context) (*BQConnector, error) {
	var err error
	projectID := utils.GetEnvOrDefault("PROJECT_ID", utils.GetEnvOrDefault("GOOGLE_CLOUD_PROJECT", ""))
	if projectID == "" {
		if projectID, err = utils.ProjectID(ctx); err != nil || projectID == "" {
			return nil, fmt.Errorf("project ID is missing: %w", err)
		}
	}
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return &BQConnector{client: client}, nil
}

func (c *BQConnector) Close() {
	if c.client != nil {
		c.client.Close()
	}
}

func (c *BQConnector) matchEmbedding(ctx context.Context, vector []float32) ([]HotelRecord, error) {
	var hotels []HotelRecord
	q := c.client.Query("SELECT base.hotel_name AS hotel_name," +
		" base.hotel_address AS hotel_address," +
		" base.hotel_description AS hotel_description," +
		" base.nearest_attractions AS nearest_attractions" +
		" FROM VECTOR_SEARCH(TABLE genai_upskilling.hotels_fictional_data," +
		" 'embeddings'," +
		" (SELECT @embeddings)," +
		"  top_k => 5," +
		" distance_type => 'COSINE'," +
		" options => '{\"use_brute_force\":true}');")
	q.Parameters = []bigquery.QueryParameter{
		{
			Name:  "embeddings",
			Value: vector,
		},
	}
	it, err := q.Read(ctx)
	if err != nil {
		return hotels, err
	}
	for {
		var row []bigquery.Value
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return hotels, err
		}
		hotels = append(hotels, HotelRecord{
			Name:            row[0].(string),
			Address:         row[1].(string),
			Description:     row[2].(string),
			NearAttractions: row[3].(string),
		})
	}
	return hotels, nil
}
