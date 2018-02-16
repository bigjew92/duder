var weather = new DuderRug("Weather", "Check the weather.");
weather.storage = weather.loadStorage();

weather.getUserLocation = function(userID) {
	if (this.storage.users == undefined) {
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
	"Sunny": ":sunny:",
	"Partly Cloudy": ":white_sun_cloud:",
	"Scattered Showers": ":white_sun_rain_cloud:",
	"Showers": ":cloud_rain:",
	"Rain": ":cloud_rain:",
	"Mostly Cloudy": ":cloud:",
	"Mostly Sunny": ":white_sun_cloud:",
	"Breezy": ":wind_blowing_face:",
	"Scattered Thunderstorms": ":thunder_cloud_rain:",
	"Thunderstorms": ":thunder_cloud_rain:",
	"Snow": ":cloud_snow:"
};

weather.addCommand("weather", function(cmd) {
	var citystate = "";
	
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
	}

	var yql = encodeURI(
		'select * from weather.forecast where woeid in (select woeid from geo.places(1) where text="' +
			citystate +
			'")'
	);
	var url =
		"https://query.yahooapis.com/v1/public/yql?q=" +
		yql +
		"&format=json&env=store%3A%2F%2Fdatatables.org%2Falltableswithkeys";

	var content = HTTP.get(4, url);
	var json = JSON.parse(content);
	if (json.query.count === 0) {
		cmd.replyToAuthor("no weather results found for that location.");
		return;
	}
	var forecast = json.query.results.channel.item.forecast;
	var title = json.query.results.channel.title.substring(17);

	var j = '{' +
			'"color": 3447003,' +
			'"title": "3 Day Forecast",' + 
			'"description": "{0}",'.format(title) + 
			'"fields":' +
			'[';

	var count = 0;
	for (var day in forecast) {
		var date = forecast[day].date.substring(0,forecast[day].date.length - 5);
		var icon = forecast[day].text;
		if (this.weatherIcons[icon] !== undefined) {
			icon = this.weatherIcons[icon];
		} else {
			icon = ":question:";
		}
		j += '{' +
			'"name": "{0} {1}",'.format(icon, date) +
			'"value": "*{0}*\\nLow: {1} High: {2}"'.format(forecast[day].text, forecast[day].low, forecast[day].high) +
		'}';
		if (++count > 2) {
			break;
		} else {
			j += ',';
		}
	}
	j += ']';

	j += '}';

	this.setUserLocation(cmd.author.id, citystate);
	cmd.replyToChannelEmbed(j);	
});
