# Challenge 9

The journey:

1. Create a vector database with real data. Use the [RAG JSS](https://cloud.google.com/architecture/ai-ml/generative-ai-rag). What needed is the Cloud SQL database with embeddings but it is easy to provision all resources and then to use only the dabase.
   Additionally, create a [test function](https://github.com/minherz/aichallenges/tree/main/challenge9/test_function).
1. Explore upgrading options for the vector database data. Try to upgrade the embeddgings (which currently uses `textembedding-gecko@001`) to use `text-embedding-005`. The model version may be different due to availability constraints.
1. Explore how to package your changes for both the function and the database, such that you can keep everything synchronized.
1. Validate that your changes are successful, and that you’re getting similar results back (proving that the migration was successful).

## Implementing the journey

Step #1 is straightforward. JSS deployment may require to be destroyed and recreated. Works best on the projects with a standard billing accounts.

I used the following command to deploy the function

```bash
gcloud functions deploy week9-rag-test --source=. --runtime=python313 --trigger-http \
--set-env-vars DB_USER=retrieval-service,DB_NAME=assistantdemo,EMBEDDING_MODEL=textembedding-gecko@001,\
INSTANCE_CONNECTION_NAME=${CLOUD_SQL_INSTANCE_NAME},LOG_EXECUTION_ID=true \
--set-secrets=DB_PASS=${CLOUD_SQL_PASSWORD} \
--service-account=${COMPUTE_DEFAULT_SA} \
--entry-point=hello_http --project=${PROJECT_ID}
```

where the follow environment variables were initialized to

- `CLOUD_SQL_INSTANCE_NAME` is the name of Cloud SQL database instance
- `CLOUD_SQL_PASSWORD` the resource name in the Secret Manager (JSS creates only one resource)
- `COMPUTE_DEFAULT_SA` the service account email of the default compute service account
- `PROJECT_ID` the project ID of the Google Cloud project where JSS has been deployed

In addition I had to grant to the compute service account permissions to access secrets:

```bash
gcloud projects add-iam-policy-binding ${PROJECT_ID} --member=${COMPUTE_DEFAULT_SA} --role="roles/secretmanager.secretAccessor"
```

In Step #2 I used Cloud SQL Studio. The only table with embeddings is

```sql
SELECT embedding FROM amenities LIMIT 10
```

Add another column

```sql
alter table amenities add column embedding005 vector(768)
```

(the vector size is 768 because it is the max for the `text-embedding-005` model)

Then populate the column

```sql
UPDATE TABLE amenities SET embedding005 = embedding('text-embedding-005', concat(name, ' - ', description))::vector
```

Finally the test_function has to be changed by replacing the name of the `embedding` column in line 111 to be `embedding005`.

All of this is described in [documentation](https://cloud.google.com/sql/docs/postgres/generate-manage-vector-embeddings#generate-an-embedding).
