import requests

def get_weather(longtitude: str, latitude: str) -> dict:
    """Retrieves the current weather report for a specified location.

    Args:
        longitude (str): The longitude of the location
        latitude (str): The latitude of the location

    Returns:
        dict: A dictionary containing the weather information.
              Includes a 'status' key ('success' or 'error').
              If 'success', includes a 'report' key with weather details.
              If 'error', includes an 'error_message' key.
    """
# Get a weather report from https://open-meteo.com/ using Open Meteo API and the longtitude and latitude as input arguments
    base_url = "https://api.open-meteo.com/v1/forecast"
    params = {
        "latitude": latitude,
        "longitude": longtitude,
        "current_weather": True,
        "temperature_unit": "celsius",
        "windspeed_unit": "kmh",
        "precipitation_unit": "mm",
        "timezone": "auto",
    }

    try:
        response = requests.get(base_url, params=params)
        response.raise_for_status()  # Raise an exception for HTTP errors
        weather_data = response.json()

        if "current_weather" in weather_data:
            current_weather = weather_data["current_weather"]
            report = (
                f"Current weather: Temperature {current_weather['temperature']}°C, "
                f"Wind Speed {current_weather['windspeed']} km/h, "
                f"Weather Code {current_weather['weathercode']}."
            )
            return {"status": "success", "report": report}
        else:
            return {"status": "error", "error_message": "Could not retrieve current weather data."}
    except requests.exceptions.RequestException as e:
        return {"status": "error", "error_message": f"Error fetching weather data: {e}"}
