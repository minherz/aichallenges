# Challenge #1

## Description

The goal is to implement a sample chat application that connects to an endpoint of the deployed Gemma 2 model.
The application supposed to:

* Manage a multi-turn chat with users
* Help users to ask questions about travel, and learn about places they are going to go
* Provides users ways to get help about their specific travel plans

## Technical challenges

Deploy an AI model from Google Cloud [Model Garden](https://cloud.google.com/model-garden).
Instrument the application so you can log and monitor its performance effectively.
Examples of custom metrics can be:

* Counter of successfully responded messages
* Elapsed time of the messages that have been sent and received; Display it using a histogram

## Technological choices

The challenge implementation follows the path of least resistance:

* Deploy model to Vertex (and not to GKE)
* Use GPU (and not TPU) to reduce costs
* Use a single CloudRun service to deploy both the front-end and the backend
* Leverage [Vertex audit logs][audit], [managed] and [log-based] metrics to simplify logging and monitoring of the application.

The application is written in Go.
For middleware it uses [echo](https://pkg.go.dev/github.com/labstack/echo/v4) web server.
Logging is implemented using built-in logging capabilities of the web server and [log/slog](https://pkg.go.dev/log/slog) structured logging.

[audit]: https://cloud.google.com/vertex-ai/docs/general/audit-logging
[managed]: https://cloud.google.com/monitoring/api/metrics_gcp
[log-based]: https://cloud.google.com/logging/docs/logs-based-metrics

## Deploying AI model

The challenge uses open model `gemma-2-9b-it` from [HuggingFace](https://huggingface.co/google/gemma-2-9b-it).
The HuggingFace version is selected because it is configured to deploy on GPU.
It is in opposite to the gemma-2 model from Google's [Model Garden](https://console.cloud.google.com/vertex-ai/publishers/google/model-garden/gemma2) which is deployed to TPU.
Deploying a model from HuggingFaces requires [access token](https://huggingface.co/docs/hub/en/security-tokens).

## Deploying to Cloud Run

The application is deployed as Cloud Run service using continuously deploy (CD) from a repository feature.
CD is configured to build the service using [Dockerfile](https://github.com/minherz/aichallenges/blob/main/challenge1/Dockerfile).
The service is configured to allow unauthenticated invocations.
In order to run correctly the service requires the following environment variables to be set for the service container:

| Variable name | Value description |
|---|---|
| ENDPOINT_ID | The endpoint identificator for the deployed model. |
| REGION_NAME | The name of the region where the model is deployed. |
| DO_DEBUG | (Optional) set to "1" to enable debug level logging for the echo webserver and the application. |

## Cost considerations

I did not find documentation describing pricing of deploying an open model on Vertex AI.
Cloud Run service costs are expected to be within a [free tier](https://cloud.google.com/run/pricing)
