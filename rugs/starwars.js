var starwars = new DuderRug("Star Wars", "Get information about the Star Wars universe.");

starwars.addCommand("starwars", function(cmd) {
    var validTypes = ["character", "location", "creature", "droid", "organization", "species", "vehicle"];

    if (cmd.args.length < 3) {
        cmd.replyToAuthor("Usage: `starwars <type> <name>`\nValid types are: `" + validTypes.join(", ") + "`");
        return;
    }

    var type = cmd.args[1].toLowerCase();
    var query = cmd.args.slice(2).join(" ");
    
    if (!validTypes.contains(type)) {
        cmd.replyToAuthor("Invalid search type. Use one of the following: `" + validTypes.join(", ") + "`");
        return;
    }

    var endpoint;
    if (type === "species") {
        endpoint = "species";
    } else {
        endpoint = type + "s";
    }
    
    var url = "https://starwars-databank-server.vercel.app/api/v1/" + endpoint + "/name/" + encodeURIComponent(query);

    Duder.startTyping(cmd.channelID);

    var headers = {
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
    };
    
    var content = HTTP.get(10, url, headers);

    // --- DEBUGGING LINE ---
    this.dprint("API Response for " + url + ":\n" + content);

    if (content === false) {
        cmd.replyToAuthor("Something went wrong while searching.");
        return;
    }

    if (content.trim().substring(0, 1) === "<") {
        cmd.replyToAuthor("Could not find a " + type + " with that name.");
        this.wprint("API returned HTML instead of JSON from URL: " + url);
        return;
    }

    var json;
    try {
        json = JSON.parse(content);
    } catch (e) {
        cmd.replyToAuthor("There was an error parsing the data from the Star Wars API.");
        this.wprint("Failed to parse JSON: " + e.message);
        return;
    }
    
    if (json === undefined || !Array.isArray(json) || json.length === 0) {
        cmd.replyToAuthor("Could not find a " + type + " with that name.");
        return;
    }

    var item = json[0];

    var embed = new EmbedMessage();
    embed.setTitle(item.name);
    embed.setDescription(item.description);
    embed.setImage(item.image);

    cmd.replyToChannelEmbed(embed.compile());
});
