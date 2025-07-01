from google.adk.agents import Agent
from google.adk.tools import google_search
from .guardrails import model_pii_guardrail
import os

MODEL_ID=os.getenv('MODEL_ID', 'gemini-2.5-flash')

root_agent = Agent(
    name='challenge10_v1',
    model=MODEL_ID,
    description='Answers questions.',
    instruction="You are a helpful assistant. "
                "When the user asks anything, "
                "Find out the answer and present it to the user clearly. "
                "Use 'google_search' to find necessary information and verify your response.",
    tools=[google_search],
    before_model_callback=model_pii_guardrail,
 )
