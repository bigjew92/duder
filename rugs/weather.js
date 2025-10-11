const OPENWEATHER_API_KEY = process.env.OPENWEATHER_API_KEY; // <-- Uses the API key from environment variable

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

// Helper to get city and state from input
function parseCityState(citystate) {
	return encodeURIComponent(citystate.trim());
}

weather.addCommand("weather", function(cmd) {
	var citystate = "";
	var saveLocation = false;

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
	var locationParam = parseCityState(citystate);

	// Fetch current weather for the location (for city name validation)
	var urlCurrent =
		"https://api.openweathermap.org/data/2.5/weather?q=" +
		locationParam +
		"&appid=" + OPENWEATHER_API_KEY + "&units=imperial";

	var contentCurrent = HTTP.get(4, urlCurrent);
	var jsonCurrent = JSON.parse(contentCurrent);

	if (!jsonCurrent || jsonCurrent.cod != 200) {
		cmd.replyToAuthor("no weather results found for that location.");
		return;
	}

	var cityName = jsonCurrent.name;
	var country = jsonCurrent.sys && jsonCurrent.sys.country ? jsonCurrent.sys.country : "";

	// Fetch 3-day forecast for the location
	var urlForecast =
		"https://api.openweathermap.org/data/2.5/forecast?q=" +
		locationParam +
		"&appid=" + OPENWEATHER_API_KEY + "&units=imperial";

	var contentForecast = HTTP.get(4, urlForecast);
	var jsonForecast = JSON.parse(contentForecast);

	if (!jsonForecast || !jsonForecast.list || jsonForecast.list.length === 0) {
		cmd.replyToAuthor("no forecast results found for that location.");
		return;
	}

	// OpenWeatherMap gives 3-hourly forecasts; we'll extract the next 3 days at 12:00 pm
	var days = {};
	for (var i = 0; i < jsonForecast.list.length; i++) {
		var entry = jsonForecast.list[i];
		var date = entry.dt_txt.split(" ")[0];
		var time = entry.dt_txt.split(" ")[1];
		if (time === "12:00:00" && Object.keys(days).length < 3) {
			days[date] = entry;
		}
	}
	// Fallback: if not enough at 12:00, fill with next available
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
		var entry = days[day];
		var weatherMain = entry.weather[0].main;
		var weatherDesc = entry.weather[0].description;
		var icon = weather.weatherIcons[weatherMain] || ":question:";
		var tempMin = Math.round(entry.main.temp_min);
		var tempMax = Math.round(entry.main.temp_max);

		fields.push({
			name: icon + " " + day,
			value: "*" + weatherDesc.charAt(0).toUpperCase() + weatherDesc.slice(1) + "*\nLow: " + tempMin + "°F  High: " + tempMax + "°F"
		});
	}

	var title = cityName + (country ? ", " + country : "");

	var embed = {
		color: 3447003,
		title: "3 Day Forecast",
		description: title,
		fields: fields
	};

	if (saveLocation) {
		this.setUserLocation(cmd.author.id, citystate);
	}
	cmd.replyToChannelEmbed(JSON.stringify(embed));
});
