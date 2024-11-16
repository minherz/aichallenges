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

TO-BE-DEFINED: How to search the model for more details

[bqdoc1]: https://cloud.google.com/bigquery/docs/generate-text-embedding
[bqdoc2]: https://cloud.google.com/bigquery/docs/reference/standard-sql/bigqueryml-syntax-create-remote-model
[bqdoc3]: https://cloud.google.com/bigquery/docs/reference/standard-sql/bigqueryml-syntax-generate-embedding
[bqdoc4]: https://cloud.google.com/bigquery/docs/samples/bigquery-load-table-gcs-json
[models]: https://cloud.google.com/vertex-ai/generative-ai/docs/embeddings/get-text-embeddings#supported-models