from google.adk.agents.callback_context import CallbackContext
from google.adk.models.llm_request import LlmRequest
from google.adk.models.llm_response import LlmResponse
from google.genai import types # For creating response content
from typing import Optional, Dict

import google.cloud.dlp
import os

PARENT=f'projects/{os.getenv("GOOGLE_CLOUD_PROJECT")}'
INFO_TYPES = [
   {"name": "CREDIT_CARD_NUMBER"},
   {"name": "EMAIL_ADDRESS"},
   {"name": "PERSON_NAME"},
   {"name": "US_SOCIAL_SECURITY_NUMBER"},
   {"name": "PHONE_NUMBER"},
   {"name": "STREET_ADDRESS"},
]

dlp_client = google.cloud.dlp_v2.DlpServiceClient()

def model_pii_guardrail(
    callback_context: CallbackContext, llm_request: LlmRequest
) -> Optional[LlmResponse]:
    agent_name = callback_context.agent_name
    print(f"--- Callback: model_guardrail running for agent: {agent_name} ---")
    if llm_request.contents:
        # Find the most recent message with role 'user'
        for content in reversed(llm_request.contents):
            if content.role == 'user' and content.parts:
                # Assuming text is in the first part for simplicity
                if content.parts[0].text:
                    content.parts[0].text = redact_string(content.parts[0].text, INFO_TYPES)
                    break # Found the last user message text


    return None

def redact_string(text: str, info_types:Dict) -> str:
    item = {'value': text}
    inspect_config = {
        'info_types': info_types,
#        'min_likelihood': google.cloud.dlp_v2.Likelihood.POSSIBLE,
    }
    deidentify_config = {
        'info_type_transformations': {
            'transformations': [
                {
                    'info_types': info_types,
                    'primitive_transformation': {
                        'replace_config': {
                            'new_value': {
                                'string_value': '[REDACTED]'
                            }
                        }
                    }
                }
            ]
        }
    }
    response = dlp_client.deidentify_content(request={
        "parent": PARENT,
        "inspect_config": inspect_config,
        "deidentify_config": deidentify_config,
        "item": item
    })
    return response.item.value
