var weather = new DuderRug("Weather", "Check the weather.");
weather.storage = weather.loadStorage();

weather.getUserLocation = function(userID) {
    if (this.storage['users'] == undefined) {
        return false;
    }
    for (var i = 0; i < this.storage['users'].length; i++) {
        var user = this.storage['users'][i];
        if (user['userID'] == userID) {
            return user['location'];
        }
    }
    return false;
}

weather.setUserLocation = function(userID,location) {
    if (this.storage['users'] == undefined) {
        this.storage['users'] = [];
    }
    var found = false;
    for (var i = 0; i < this.storage['users'].length; i++) {
        if (this.storage['users'][i]['userID'] == userID) {
            this.storage['users'][i]['location'] = location;
            found = true;
            break;
        }
    }
    if (!found) {
        this.storage['users'].push( {'userID': userID, 'location': location});
    }
    rug.saveStorage(this.storage);
}

weather.padRight = function(text, len) {
    var count = len - text.length;
    var p = "";
    for(var i = 0; i < count; i++) {
        p += " ";
    }
    return text + p;
}

weather.addCommand("weather", function() {
    var citystate = "";

    if (cmd.args.length < 2) {
        var location = rug.getUserLocation(cmd.author.id);
        if (location == false) {
            cmd.replyToAuthor("usage: `weather city, ST`");
            return;
        }
        citystate = location;
    } else {
        for(var i = 1; i < cmd.args.length; i++) {
            citystate += cmd.args[i] + " ";
        }
    }

    var yql=encodeURI("select * from weather.forecast where woeid in (select woeid from geo.places(1) where text=\"" + citystate + "\")");
    var url="https://query.yahooapis.com/v1/public/yql?q=" + yql + "&format=json&env=store%3A%2F%2Fdatatables.org%2Falltableswithkeys";

    var content = HTTP.get(4, url);
    json = JSON.parse(content);
    if (json.query.count == 0) {
        cmd.replyToAuthor("no weather results found for that location.");
        return;
    }

    var forecast = json.query.results.channel.item.forecast;

    var dates = "";
    var lows = "";
    var highs = "";
    var text = "";
    var count = 0;
    for(var day in forecast) {
        var padSize = 0;
        padSize = Math.max(padSize, forecast[day].date.length);
        padSize = Math.max(padSize, forecast[day].text.length);
        padSize += 4;

        dates += rug.padRight(forecast[day].date, padSize);
        lows += rug.padRight("Low: " + forecast[day].low, padSize);
        highs += rug.padRight("High: " + forecast[day].high, padSize);
        text += rug.padRight(forecast[day].text, padSize);
        if (++count > 2) {
            break;
        }
    }

    var title = json.query.results.channel.title;
    title = title.substring(17);
    var msg = ":white_sun_cloud: " + title + "\n";
    msg += "`" + dates + "|\n" + lows+ "|\n" + highs + "|\n" + text + "|" + "`";

    rug.setUserLocation(cmd.author.id, citystate);
    cmd.replyToChannel(msg);
});
