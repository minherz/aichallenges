# Challenge 10

The journey:

1. Use ADK to create a simple application that uses agentic pipeline to processa a prompt.
1. Implement a method on the LLM call where, pre and post call, you can validate that you do not have a sensitive data leak (ex. Someone’s credit card information) that is shared with downstream processes or logs
1. Implement downstream logging that provides an analyst the necessary data to be able to design or improve the LLM prompts and LLM agents that are used. You’re building your own internal dataset that can be used for additional training, tuning, or engineering. This should include:
   * The prompt and response (Similar to what you did in Challenge 1)
   * Redact any sensitive material, you don’t want this to leak into your training or validation data!
   * Provide a stream endpoint (ex. A topic with the data) or a table that someone could consume the data you’ve cleaned from. This could be a Pub/Sub topic, a table in a database or even a set of objects in a GCS bucket with the chat transcripts.

## Implementing the journey

Step #1 follows the standard ADK for Python tutorial to define a single agent that simply asks questions.
The agent uses a model defined in `MODEL_ID` envvar that defaults to `gemini-2.5-flash`.

Step #2 features the guardrail supported in ADK agents that allow to verify input to the model (with `before_model_callback` parameter) or to the tool(s) (with `before_tool_callback` parameter).
The defined guardrail (`model_pii_guardrail`) calls DLP API to replace PII with random dummy data.
The code is PoC and is not production ready implementation.

Step #3 is not implemented explicitly. The model logs the prompts. The missing part is to enforce that logging is done using the structured logging format so the stored logged can be further analysed using Log Analytics. Alternatively, the logging can be disabled and the input can be stored in a bucket or DB or can be pushed to PubSub.

## How to run

1. Create `.venv` file
1. Copy the following code into the `.venv` file and set up your API key:

   ```text
   GOOGLE_GENAI_USE_VERTEXAI="False"
   GOOGLE_API_KEY="Your API Key goes here"
   MODEL_ID="gemini-2.5-flash"
   GOOGLE_CLOUD_PROJECT="Project ID for calling DLP API"
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

   Note that DLP API should be enabled on the project that you set with `GOOGLE_CLOUD_PROJECT` envvar.

1. Run the Web interface with `adk web` or run CLI interface with `adk run challenge10` from the repository level folder.
   If you run the Web interface, select `challenge10`.
   > [!NOTE]
   > The repository level folder is `aifolder/` and not `aifolder/challenge10`.

## Observations

* DLP does not redact certain information if it considers the values invalid:
  * SSN is not redacted if the number is not "real" (e.g. a number starting with "9")
  * Credit card numbers that do not pass basic validation i.e. 1st digit is IIN, first 6-8 digits are IIN or BIN, etc.
* For some reason DLP redacts an email as three consequitive but distinct blocks
* ADK writes to log a lot of information and all of it at INFO level. Implementing Step #3 as described above means to compose a query that does it or to log them explicitly
* Alternative implementations of Step #3 can be done in the same guardrail callback that redacts PII from the prompt and by using `after_model_callback` callback to catch the response

## Test input

Consider using the following input to see how PII guardrail works:

```text
My name is John Doe, I live at 123 Main Street, Anytown, USA. My credit card number is 3782-8224-6310-005 and my social security number is 123-88-7777. My email is john.doe@example.com and my phone number is 555-123-4567. I want to learn about the benefits of cloud computing for someone with my background. What are the privacy and security implications I should be aware of?
```
