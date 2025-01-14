# Challenge 5

Modify the simple GenAI chat application to use embeddings.

## Selection of Vector Store

There are several vector stores that can be used to store embeddings in Google Cloud:

* Vertex AI Search
* AlloyDB (in Cloud SQL)
* BigQuery
* Spanner

The selected choice is BigQuery.

**Pros:**

* One of the well known brands outside of Google Cloud
* Provides multiple additional functionalities
* Integrate with Cloud Logging to enable further experiments and to be used in Observability focused Hackathons
* Simpler and cheaper (relatively) compared to some of other choices

**Cons:**

* UI is not vector store oriented
* Querying requires additional interface

## Loading embeddings to BigQuery

SQL in this section will assume that the BigQuery dataset has a name `genai_upskilling`.
For generalization `PROJECT_ID` will be used as a placeholder for real project ID.

The embeddings are created from the data stored in a BG table. The table is created from a Newline delimitered JSON with the following schema:

```json
{
    "hotel_name": "the name of the fictional hotel",
    "hotel_address": "a fictional hotel address",
    "hotel_description": "a description of the hotel including who might want to stay here and what to expect",
    "nearest_attractions": "local attractions near the hotel tourists might want to see"
}
```

I used BigQuery Explorer to do that as described in [documentation][bqdoc4].
The table was named `hotels_fictional_data`. Then I added a new column to store embeddings:

```sql
ALTER TABLE `genai_upskilling.hotels_fictional_data` ADD COLUMN embeddings ARRAY<FLOAT64>
```

After that I've followed BigQuery [documentation]][bqdoc1] to generate text embeddings.
Following the documentation I create a connection to Vertex AI remote model with the fully qualified name:

```text
projects/PROJECT_ID/locations/us/connections/vertexai-to-bq-connection
```

Then I ran SQL to create a BigQuery model using built-in "text_embedding_004" as a remote model, selecting it from [supported models list][models]:

```sql
CREATE OR REPLACE MODEL `genai_upskilling.text_embedding_004`
REMOTE WITH CONNECTION `projects/PROJECT_ID/locations/us/connections/vertexai-to-bq-connection`
OPTIONS (ENDPOINT = 'text_embedding_004');
```

Finally, I populated the "embeddings" column using the hotel description, address and near attractions:

```sql
UPDATE `genai_upskilling.hotels_fictional_data`
SET embeddings = (
  SELECT ml_generate_embedding_result
  FROM ML.GENERATE_EMBEDDING(
      MODEL `genai_upskilling.text_embedding_004`,
      (SELECT CONCAT(hotel_description, ". Hotel is located at ", hotel_address, ". Nearest attractions include ", nearest_attractions) AS content),
      STRUCT(
            TRUE AS flatten_json_output, 
            'RETRIEVAL_QUERY' AS task_type)
))
WHERE TRUE
```

See documentation about [ML.GENERATE_EMBEDDING function][bqdoc3] for more details.

[bqdoc1]: https://cloud.google.com/bigquery/docs/generate-text-embedding
[bqdoc3]: https://cloud.google.com/bigquery/docs/reference/standard-sql/bigqueryml-syntax-generate-embedding
[bqdoc4]: https://cloud.google.com/bigquery/docs/samples/bigquery-load-table-gcs-json
[models]: https://cloud.google.com/vertex-ai/generative-ai/docs/embeddings/get-text-embeddings#supported-models

### (Optionally) Add Vector Index to improve the search performance

Use [CREATE VECTOR INDEX](https://cloud.google.com/bigquery/docs/reference/standard-sql/data-definition-language#create_vector_index_statement) statement.

## BigQuery Vector search

Executing vector search in BigQuery requires:

1. [Create an embedding][embed_text] from a text / prompt
1. Execute a query that uses [VECTOR_SEARCH function][bq_vector_search] to find "closest" matching from the embeddings [generated previously](#Loading_embeddings_to_BigQuery)

Creating embedding is straightforward. Currently it requires the use of the auto-generated version of the client library (like in Challenge 1). Code can be found in [pkg/agents/embedding.go](https://github.com/minherz/aichallenges/blob/challenge5/challenge5/pkg/agents/embedding.go).

Executing the BigQuery query requires google/bigquery package. It supports "pushing" the embedding as a parameter into the following query:

```sql
SELECT base.hotel_name AS hotel_name,
	base.hotel_address AS hotel_address,
	base.hotel_description AS hotel_description,
	base.nearest_attractions AS nearest_attractions
FROM VECTOR_SEARCH(TABLE genai_upskilling.hotels_fictional_data,
	'embeddings',
	(SELECT @embeddings),
	 top_k => 5,
	distance_type => 'COSINE',
	options => '{"use_brute_force":true}')
```

Note the following things about the query:

- Uses "brute force" option to search for matching embeddings; for large data it is recommended to [create an index][bq_vector_index] and use the different options: `'{"fraction_lists_to_search": 0.005}'`.
- Uses BigQuery select statement to select the parameterized embedding vector as a value; alternatively it could be stored in a different table or derived using a sub-query.
- Uses `base.` prefix to retrieve other fields from the table; if the value is queried from another table, other fields from that table can be retrieved using `query.` prefix.

[embed_text]: https://cloud.google.com/vertex-ai/generative-ai/docs/embeddings/get-text-embeddings#get_text_embeddings_for_a_snippet_of_text
[bq_vector_search]: https://cloud.google.com/bigquery/docs/vector-search#use_the_vector_search_function_with_brute_force
[bq_vector_index]: https://cloud.google.com/bigquery/docs/vector-search#create_a_vector_index

### Call BigQuery from Cloud Run

There is no special configuration to call BigQuery API from Cloud Run.
Communication between the Cloud Run service and BigQuery and between the service and Vertex are secured because the grpc packets are [encrypted in-transit](https://cloud.google.com/docs/security/encryption-in-transit) and [do not leave Google internal network](https://cloud.google.com/run/docs/securing/private-networking#to-other-services).
