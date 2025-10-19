var starwars = new DuderRug("Star Wars", "Get information about Star Wars characters and locations.");

starwars.addCommand("starwars", function(cmd) {
    if (cmd.args.length < 3) {
        cmd.replyToAuthor("Usage: `starwars <character|location> <name>`");
        return;
    }

    var type = cmd.args[1].toLowerCase();
    var query = cmd.args.slice(2).join(" ");
    var url = "https://starwars-databank-server.vercel.app/api/v1/";

    Duder.startTyping(cmd.channelID);

    if (type === "character") {
        url += "characters/name/" + encodeURIComponent(query);
    } else if (type === "location") {
        // Updated to use the more direct /name/ endpoint for locations
        url += "locations/name/" + encodeURIComponent(query);
    } else {
        cmd.replyToAuthor("Invalid search type. Use `character` or `location`.");
        return;
    }

    var content = HTTP.get(10, url, {});
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

    if (type === "character") {
        if (json === undefined || json.name === undefined) {
            cmd.replyToAuthor("Could not find a character with that name.");
            return;
        }

        var character = json;

        var embed = new EmbedMessage();
        embed.setTitle(character.name);
        embed.setDescription(character.description);
        embed.setThumbnail(character.image);
        embed.addField("Homeworld", character.homeworld || "Unknown");
        embed.addField("Species", character.species || "Unknown");
        embed.addField("Affiliations", character.affiliations ? character.affiliations.join(", ") : "None");
        embed.addField("Masters", character.masters ? character.masters.join(", ") : "None");
        embed.addField("Apprentices", character.apprentices ? character.apprentices.join(", ") : "None");

        cmd.replyToChannelEmbed(embed.compile());

    } else if (type === "location") {
        // Adjusted to handle the single location object
        if (json === undefined || json.name === undefined) {
            cmd.replyToAuthor("Could not find a location with that name.");
            return;
        }

        var location = json;

        var embed = new EmbedMessage();
        embed.setTitle(location.name);
        embed.setDescription(location.description);
        embed.setThumbnail(location.image);
        embed.addField("Climate", location.climate || "Unknown");
        embed.addField("Terrain", location.terrain || "Unknown");
        embed.addField("Notable Inhabitants", location.notable_inhabitants ? location.notable_inhabitants.join(", ") : "None");

        cmd.replyToChannelEmbed(embed.compile());
    }
});
