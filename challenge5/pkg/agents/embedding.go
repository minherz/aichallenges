package agents

import (
	"context"
	"fmt"
	"log/slog"

	aiplatform "cloud.google.com/go/aiplatform/apiv1"
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"github.com/minherz/aichallenges/challenge1/pkg/utils"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/structpb"
)

type Embedding struct {
	client   *aiplatform.PredictionClient
	endpoint string
}

func NewEmbedding(ctx context.Context) (*Embedding, error) {
	region := utils.GetEnvOrDefault("REGION_NAME", "")
	if region == "" {
		v, err := utils.Region(ctx)
		if err != nil || v == "" {
			return nil, fmt.Errorf("location is missing: %w", err)
		}
		region = v
	}
	endpoint := fmt.Sprintf("%s-aiplatform.googleapis.com:443", region)
	client, err := aiplatform.NewPredictionClient(ctx, option.WithEndpoint(endpoint))
	if err != nil {
		return nil, err
	}
	projectID := utils.GetEnvOrDefault("PROJECT_ID", utils.GetEnvOrDefault("GOOGLE_CLOUD_PROJECT", ""))
	if projectID == "" {
		v, err := utils.ProjectID(ctx)
		if err != nil || v == "" {
			return nil, fmt.Errorf("project ID is missing: %w", err)
		}
		projectID = v
	}
	embeddingModel := utils.GetEnvOrDefault("EMBEDDING_MODEL", "text-embedding-004")
	endpoint = fmt.Sprintf("projects/%s/locations/%s/publishers/google/models/%s", projectID, region, embeddingModel)
	slog.Debug("embedding is initialized", slog.String("endpoint", endpoint))
	return &Embedding{client: client, endpoint: endpoint}, nil
}

func (e *Embedding) Close() {
	if e.client != nil {
		e.client.Close()
	}
}

func (e *Embedding) Embed(ctx context.Context, input string) ([]float32, error) {
	var vector []float32

	instances := []*structpb.Value{
		structpb.NewStructValue(&structpb.Struct{
			Fields: map[string]*structpb.Value{
				"content":   structpb.NewStringValue(input),
				"task_type": structpb.NewStringValue("QUESTION_ANSWERING"),
			},
		}),
	}
	params := structpb.NewStructValue(&structpb.Struct{
		Fields: map[string]*structpb.Value{
			"outputDimensionality": structpb.NewNumberValue(float64(dimensionality)),
		},
	})
	req := &aiplatformpb.PredictRequest{
		Endpoint:   e.endpoint,
		Instances:  instances,
		Parameters: params,
	}
	resp, err := e.client.Predict(ctx, req)
	if err != nil {
		return vector, err
	}
	if len(resp.Predictions) != 1 {
		return vector, fmt.Errorf("unexpected number of embeddings")
	}
	values := resp.Predictions[0].GetStructValue().Fields["embeddings"].GetStructValue().Fields["values"].GetListValue().Values
	vector = make([]float32, len(values))
	for i, value := range values {
		vector[i] = float32(value.GetNumberValue())
	}
	return vector, nil
}
