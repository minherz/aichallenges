# Challenge #2

## Description

The goal is to implement a sample chat application, similar to one in Challenge 1, that connects to Gemini model.
The application does:

* Manage a multi-turn chat with users
* Help users to ask questions about travel, and learn about places they are going to go
* Provides users ways to get help about their specific travel plans
* Support customizable system instructions

## Technical challenges

Use Gemini 1.5 flash model.
Instrument the application so you can log and monitor its performance effectively.
Examples of custom metrics can be:

* Counter of successfully responded messages
* Elapsed time of the messages that have been sent and received; Display it using a histogram

Upload the system instructions from the provided location.
Update instructions dynamically when changed.

## Technological choices

The challenge implementation follows the path of least resistance:

* Use a single CloudRun service to deploy both the front-end and the backend
* Leverage [Vertex audit logs][audit], [managed] and [log-based] metrics to simplify logging and monitoring of the application.
* Use mounting volumes feature of CloudRun to mount to Cloud Storage

The application is written in Go.
For middleware it uses [echo](https://pkg.go.dev/github.com/labstack/echo/v4) web server.
Logging is implemented using built-in logging capabilities of the web server and [log/slog](https://pkg.go.dev/log/slog) structured logging.

[audit]: https://cloud.google.com/vertex-ai/docs/general/audit-logging
[managed]: https://cloud.google.com/monitoring/api/metrics_gcp
[log-based]: https://cloud.google.com/logging/docs/logs-based-metrics

## Deploying AI model

Because the challenge uses Gemini there is no need for deployment of the AI model.
It is possible to use multiple versions of Gemini. See Gemini model versions in [documentation](https://cloud.google.com/vertex-ai/generative-ai/docs/learn/model-versions).

## Deploying to Cloud Run

The application is deployed as Cloud Run service using continuously deploy (CD) from a repository feature.
CD is configured to build the service using [Dockerfile](https://github.com/minherz/aichallenges/blob/main/challenge2/Dockerfile).
The service is configured to allow unauthenticated invocations.
The service container is configured to mount the GCS bucket. The expected object hierarchy has a single object with the path `/current/system_instructions.txt`.
The bucket has object versioning enabled to comply with the challenge's requirements.
In order to run correctly the service requires the following environment variables to be set for the service container:

| Variable name | Value description |
|---|---|
| GEMINI_MODEL_NAME | The name of the Gemini model version. If not provided uses `gemini-1.5-flash-001`. |
| REGION_NAME | (Optional) The name of the region where the model inference is invoked. If not provided it uses the same region as the Cloud Run service. |
| SYS_INSTRUCTION_PATH | The path to the volume in the service container that is configured to mount to GCS bucket with the system instructions. |
| DO_DEBUG | (Optional) set to "1" to enable debug level logging for the echo webserver and the application. |

## Cost considerations

I did not find documentation describing pricing of deploying an open model on Vertex AI.
Cloud Run service costs are expected to be within a [free tier](https://cloud.google.com/run/pricing)