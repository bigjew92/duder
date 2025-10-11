var GEOCODE_URL = "https://api.openweathermap.org/geo/1.0/direct";
var WEATHER_URL = "https://api.openweathermap.org/data/2.5/weather";
// It seems you are getting a One Call API response, which is better.
// We will use the forecast URL but the parsing logic will handle the new format.
var FORECAST_URL = "https://api.openweathermap.org/data/2.5/forecast"; // This might redirect or your key defaults to a different API version.

var weather = new DuderRug("Weather", "Check the weather.");
weather.storage = weather.loadStorage();

weather.getAPIKey = function() {
	if (this.storage.settings === undefined) {
		return false;
	} else if (this.storage.settings.api_key === undefined) {
		return false;
	}
	return this.storage.settings.api_key;
};

weather.setAPIKey = function(key) {
	if (this.storage.settings === undefined) {
		this.storage.settings = {};
	}
	this.storage.settings.api_key = key;
	this.saveStorage(this.storage);
};


weather.getUserLocation = function(userID) {
    if (this.storage.users === undefined) {
        return false;
    }
    for (var i = 0; i < this.storage.users.length; i++) {
        var user = this.storage.users[i];
        if (user.userID === userID) {
            return user.location;
        }
    }
    return false;
};

weather.setUserLocation = function(userID, location) {
    if (this.storage.users === undefined) {
        this.storage.users = [];
    }
    var found = false;
    for (var i = 0; i < this.storage.users.length; i++) {
        if (this.storage.users[i].userID === userID) {
            this.storage.users[i].location = location;
            found = true;
            break;
        }
    }
    if (!found) {
        this.storage.users.push({ userID: userID, location: location });
    }
    this.saveStorage(this.storage);
};

weather.padRight = function(text, len) {
    var count = len - text.length;
    var p = "";
    for (var i = 0; i < count; i++) {
        p += " ";
    }
    return text + p;
};

weather.weatherIcons = {
    Clear: ":sunny:",
    Clouds: ":cloud:",
    Rain: ":cloud_rain:",
    Drizzle: ":white_sun_rain_cloud:",
    Thunderstorm: ":thunder_cloud_rain:",
    Snow: ":cloud_snow:",
    Mist: ":fog:",
    Haze: ":fog:",
    Fog: ":fog:",
    Smoke: ":fog:",
    Dust: ":fog:",
    Sand: ":fog:",
    Ash: ":fog:",
    Squall: ":wind_blowing_face:",
    Tornado: ":tornado:"
};

// Get coordinates using OpenWeatherMap Geocoding API
weather.getCoordinates = function(citystate, apiKey) {
    try {
        var url =
            GEOCODE_URL +
            "?q=" +
            citystate +
            "&limit=1&appid=" +
            apiKey;
        var content = HTTP.get(4, url);
        var data = JSON.parse(content);

        if (!Array.isArray(data) || data.length === 0) {
            console.error("Geocoding API returned no results for location:", citystate);
            return null;
        }
        return {
            lat: data[0].lat,
            lon: data[0].lon,
            display_name:
                (data[0].name || "") +
                (data[0].state ? ", " + data[0].state : "") +
                (data[0].country ? ", " + data[0].country : "")
        };
    } catch (err) {
        console.error("Error during geocoding request for location:", citystate, err);
        return null;
    }
};

weather.addCommand("weather", function(cmd) {
    if (cmd.args.length > 1 && cmd.args[1].toLowerCase() === "setkey") {
        if (cmd.args.length === 3) {
            this.setAPIKey(cmd.args[2]);
            cmd.replyToAuthor("OpenWeather API key has been saved.");
            return;
        } else {
            cmd.replyToAuthor("usage: `weather setkey YOUR_API_KEY`");
            return;
        }
    }

    var OPENWEATHER_API_KEY = this.getAPIKey();
    if (OPENWEATHER_API_KEY === false) {
        cmd.replyToAuthor("The OpenWeather API key has not been set. Use `weather setkey YOUR_API_KEY`.");
        return;
    }

    var citystate = "";
    var saveLocation = false;

    try {
        if (cmd.args.length < 2) {
            var location = this.getUserLocation(cmd.author.id);
            if (location === false) {
                cmd.replyToAuthor("usage: `weather city, ST`");
                return;
            }
            citystate = location;
        } else {
            for (var i = 1; i < cmd.args.length; i++) {
                citystate += cmd.args[i] + " ";
            }
            saveLocation = true;
        }

        citystate = citystate.trim();
        var coords = this.getCoordinates(citystate, OPENWEATHER_API_KEY);
        if (!coords) {
            cmd.replyToAuthor("No results found for that location.");
            return;
        }

        var lat = coords.lat;
        var lon = coords.lon;
        var locationLabel = coords.display_name;
        
        // --- Note: We're now using a different API endpoint for the forecast that returns richer data. ---
        // The free tier of OpenWeather often defaults to the "OneCall" API.
        var urlForecast = "https://api.openweathermap.org/data/2.5/onecall" +
            "?lat=" + lat +
            "&lon=" + lon +
            "&exclude=current,minutely,hourly,alerts" + // We only need the daily forecast
            "&appid=" + OPENWEATHER_API_KEY +
            "&units=imperial";

        var contentForecast, jsonForecast;
        try {
            contentForecast = HTTP.get(4, urlForecast);
            jsonForecast = JSON.parse(contentForecast);
        } catch (err) {
            console.error("Error fetching forecast for", locationLabel, err);
            cmd.replyToAuthor("Failed to retrieve weather forecast for that location.");
            return;
        }

        // --- START: Updated forecast parsing logic ---
        if (!jsonForecast || !jsonForecast.daily || jsonForecast.daily.length === 0) {
            console.error("Forecast API returned no valid daily result for", locationLabel, jsonForecast);
            cmd.replyToAuthor("No forecast results found for that location.");
            return;
        }

        var fields = [];
        // Loop through the first 3 days of the 'daily' array
        for (var i = 0; i < 3 && i < jsonForecast.daily.length; i++) {
            try {
                var entry = jsonForecast.daily[i];
                // Convert timestamp to a readable date
                var date = new Date(entry.dt * 1000).toISOString().split('T')[0];

                var weatherMain = entry.weather[0].main;
                var weatherDesc = entry.weather[0].description;
                var icon = weather.weatherIcons[weatherMain] || ":question:";
                // Temperatures are now in entry.temp.min and entry.temp.max
                var tempMin = Math.round(entry.temp.min);
                var tempMax = Math.round(entry.temp.max);

                fields.push({
                    name: icon + " " + date,
                    value:
                        "*" +
                        weatherDesc.charAt(0).toUpperCase() +
                        weatherDesc.slice(1) +
                        "*\\nLow: " + tempMin + "°F  High: " + tempMax + "°F"
                });
            } catch (err) {
                console.error("Error parsing forecast entry for day index:", i, err);
            }
        }
        // --- END: Updated forecast parsing logic ---

        var embed = {
            color: 3447003,
            title: "3 Day Forecast",
            description: locationLabel,
            fields: fields
        };

        if (saveLocation) {
            this.setUserLocation(cmd.author.id, citystate);
        }
        cmd.replyToChannelEmbed(JSON.stringify(embed));
    } catch (err) {
        console.error("Unexpected error in weather command handler:", err);
        cmd.replyToAuthor("Something went wrong while fetching the weather.");
    }
});
