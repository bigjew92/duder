const GEOCODE_URL = "https://api.openweathermap.org/geo/1.0/direct";
const FORECAST_URL = "https://api.openweathermap.org/data/2.5/forecast";

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
        var query = citystate;
        // Append ",us" to be more specific for US locations as per the working example
        if (query.split(',').length === 2) {
            query += ',us';
        }

        var url =
            GEOCODE_URL +
            "?q=" + query +
            "&limit=1&appid=" +
            apiKey;
        var content = HTTP.get(4, url);
        var data = JSON.parse(content);

        if (!Array.isArray(data) || data.length === 0) {
            this.dprint("Geocoding API returned no results for location: " + citystate);
            return null;
        }
        return {
            lat: data[0].lat,
            lon: data[0].lon,
            display_name:
                (data[0].name || "") +
                (data[0].state ? ", " + data[0].state : "") +
                (", " + data[0].country || "")
        };
    } catch (err) {
        this.wprint("Error during geocoding request for location '" + citystate + "': " + err);
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
                citystate += cmd.args
