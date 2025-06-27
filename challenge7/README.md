# Challenge 7

## Description

This challenge implements the following:

* Builds an application that shows a current weather based on NL questions.
* Uses Open Weather API at [https://open-meteo.com/] to get weather information based on geo-coordinates.
* Implements the application using Gemini and an agentic framework (specifically [ADK](https://google.github.io/adk-docs/)).

## How to run

1. Create `.venv` file
1. Copy the following code into the `.venv` file and set up your API key:

   ```text
   GOOGLE_GENAI_USE_VERTEXAI="False"
   GOOGLE_API_KEY="Your API Key goes here"
   MODEL_ID="gemini-2.5-flash"
   ```

   If the model that you use is hosted in other provider than Google, use the following table to match API key environment variable to the host:

   | Model provider | EnvVar name for API Key |
   |---|---|
   | Open AI | `OPENAI_API_KEY` |
   | Anthropic | `ANTHROPIC_API_KEY` |
   | Google | `GOOGLE_API_KEY` |

   If you want to use Google Vertex use the following envvars instead:

   ```text
   GOOGLE_GENAI_USE_VERTEXAI="True"
   GOOGLE_APPLICATION_CREDENTIALS="Path to JSON file or JSON string with credentials"
   GOOGLE_CLOUD_PROJECT="project ID to use Vertex"
   MODEL_ID="gemini-2.5-flash"
   ```

1. Run the Web interface with `adk web` or run CLI interface with `adk run challenge7` from the repository level folder.
   **NOTE:** The repository level folder is `aifolder/` and not `aifolder/challenge7`.
