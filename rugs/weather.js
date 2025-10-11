var OPENWEATHER_API_KEY = process.env.OPENWEATHER_API_KEY; // API key must be provided in your environment
var GEOCODE_URL = "https://api.openweathermap.org/geo/1.0/direct";
var WEATHER_URL = "https://api.openweathermap.org/data/2.5/weather";
var FORECAST_URL = "https://api.openweathermap.org/data/2.5/forecast";

var weather = new DuderRug("Weather", "Check the weather.");
weather.storage = weather.loadStorage();

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
		p += " ";
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
weather.getCoordinates = function(citystate) {
	try {
		var url =
			GEOCODE_URL +
			"?q=" +
			encodeURIComponent(citystate) +
			"&limit=1&appid=" +
			OPENWEATHER_API_KEY;
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
		var coords = this.getCoordinates(citystate);
		if (!coords) {
			cmd.replyToAuthor("No results found for that location.");
			return;
		}

		var lat = coords.lat;
		var lon = coords.lon;
		var locationLabel = coords.display_name;

		// Fetch current weather for validation
		var urlCurrent =
			WEATHER_URL +
			"?lat=" +
			lat +
			"&lon=" +
			lon +
			"&appid=" +
			OPENWEATHER_API_KEY +
			"&units=imperial";
		var contentCurrent, jsonCurrent;
		try {
			contentCurrent = HTTP.get(4, urlCurrent);
			jsonCurrent = JSON.parse(contentCurrent);
		} catch (err) {
			console.error("Error fetching current weather for", locationLabel, err);
			cmd.replyToAuthor("Failed to retrieve current weather for that location.");
			return;
		}

		if (!jsonCurrent || jsonCurrent.cod != 200) {
			console.error("Weather API returned no valid result for", locationLabel, jsonCurrent);
			cmd.replyToAuthor("No weather results found for that location.");
			return;
		}

		// Fetch 3-day forecast for the location
		var urlForecast =
			FORECAST_URL +
			"?lat=" +
			lat +
			"&lon=" +
			lon +
			"&appid=" +
			OPENWEATHER_API_KEY +
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

		if (!jsonForecast || !jsonForecast.list || jsonForecast.list.length === 0) {
			console.error("Forecast API returned no valid result for", locationLabel, jsonForecast);
			cmd.replyToAuthor("No forecast results found for that location.");
			return;
		}

		// Extract the next 3 days at 12:00 pm
		var days = {};
		for (var i = 0; i < jsonForecast.list.length; i++) {
			var entry = jsonForecast.list[i];
			var date = entry.dt_txt.split(" ")[0];
			var time = entry.dt_txt.split(" ")[1];
			if (time === "12:00:00" && Object.keys(days).length < 3) {
				days[date] = entry;
			}
		}
		// Fallback: fill with next available
		var idx = 0;
		while (Object.keys(days).length < 3 && idx < jsonForecast.list.length) {
			var entry = jsonForecast.list[idx];
			var date = entry.dt_txt.split(" ")[0];
			if (!days[date]) {
				days[date] = entry;
			}
			idx++;
		}

		var fields = [];
		for (var day in days) {
			try {
				var entry = days[day];
				var weatherMain = entry.weather[0].main;
				var weatherDesc = entry.weather[0].description;
				var icon = weather.weatherIcons[weatherMain] || ":question:";
				var tempMin = Math.round(entry.main.temp_min);
				var tempMax = Math.round(entry.main.temp_max);

				fields.push({
					name: icon + " " + day,
					value:
						"*" +
						weatherDesc.charAt(0).toUpperCase() +
						weatherDesc.slice(1) +
						"*\\nLow: " +
						tempMin +
						"°F  High: " +
						tempMax +
						"°F"
				});
			} catch (err) {
				console.error("Error parsing forecast entry for day:", day, err);
			}
		}

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
