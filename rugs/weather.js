//https://query.yahooapis.com/v1/public/yql?q=select%20item.condition%20from%20weather.forecast%20where%20woeid%20%3D%202487889&format=json&env=store%3A%2F%2Fdatatables.org%2Falltableswithkeys
//select item.condition from weather.forecast where woeid in (select woeid from geo.places(1) where text="fontana, ca")
var weather = new DuderRug("Weather", "Check the weather.");



weather.padRight = function(text, len) {
    var count = len - text.length;
    var p = "";
    for(var i = 0; i < count; i++) {
        p += " ";
    }
    return text + p;
}

weather.addCommand("weather", function() {
    if (cmd.args.length < 2) {
        cmd.replyToAuthor("usage: `weather city, ST`");
        return;
    }

    var citystate = "";
    for(var i = 1; i < cmd.args.length; i++) {
        citystate += cmd.args[i];
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
        dates += rug.padRight(forecast[day].date, 20);
        lows += rug.padRight("Low: " + forecast[day].low, 20);
        highs += rug.padRight("High: " + forecast[day].high, 20);
        text += rug.padRight(forecast[day].text, 20);
        if (++count > 2) {
            break;
        }
    }

    var title = json.query.results.channel.title;
    title = title.substring(17);
    var msg = ":white_sun_cloud: " + title + "\n";
    msg += "`" + dates + "|\n" + lows+ "|\n" + highs + "|\n" + text + "|" + "`";

    cmd.replyToChannel(msg);
});
