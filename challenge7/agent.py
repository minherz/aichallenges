from google.adk.agents import Agent
from google.adk.tools import google_search
from google.adk.tools.agent_tool import AgentTool
from .tools import get_weather
import os

MODEL_ID=os.getenv('MODEL_ID', 'gemini-2.5-flash')

location_agent = Agent(
    name='location_agent_v1',
    model=MODEL_ID,
    description='Returns geolocation of the place.',
    instruction="You are an experienced cartographer. "
                "When the user asks a question about a specific place, "
                "use 'google_search' to find the geocoordinates of the place. "
                "convert the geocoordinates to the longitude and latitude."
                "Make sure that the longitude and latitude are in the decimal degrees format."
                "If they are in the different format, convert the longitude and latitude to the decimal degrees format."
                "Return the longtitude and latitude with the two digits after decimal point. "
                "If you cannot find it, return a clear error message.",
    tools=[google_search],
)

root_agent = Agent(
    name='weather_agent_v1',
    model=MODEL_ID,
    description='Returns weather information for requested location.',
    instruction="You are a helpful weather assistant. "
                "When the user asks for the weather in a specific city, "
                "use the 'location_agent' tool to find the geocoordinates of the city in the user's request. "
                "convert the geocoordinates to the longitude and latitude."
                "make sure that the longitude and latitude are in the decimal degrees format."
                "if they are in the different format, convert the longitude and latitude to the decimal degrees format."
                "pass the longtitude and latitude to the 'get_weather' tool to find the information about the weather. "
                "Use WMO weather interpretation codes to interprete weather_code returned in the response."
                "If one of the tool returns an error or cannot provide results, inform the user politely. "
                "If the tool is successful, present the weather report clearly."
                "Handle only weather. Politely respond if the user asks for something else.",
    tools=[AgentTool(agent=location_agent), get_weather],
 )
