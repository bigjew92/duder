var starwars = new DuderRug("Star Wars", "Get information about Star Wars characters and locations.");

starwars.addCommand("starwars", function(cmd) {
    if (cmd.args.length < 3) {
        cmd.replyToAuthor("Usage: `starwars <character|location> <name>`");
        return;
    }

    var type = cmd.args[1].toLowerCase();
    var query = cmd.args.slice(2).join(" ");
    var url = "https://starwars-databank.vercel.app/api/v1/";

    Duder.startTyping(cmd.channelID);

    if (type === "character") {
        url += "characters/search/" + encodeURIComponent(query);

        var content = HTTP.get(10, url, {});
        if (content === false) {
            cmd.replyToAuthor("Something went wrong while searching for that character.");
            return;
        }

        var json = JSON.parse(content);
        if (json === undefined || json.characters.length === 0) {
            cmd.replyToAuthor("Could not find a character with that name.");
            return;
        }

        var character = json.characters[0];

        var embed = new EmbedMessage();
        embed.setTitle(character.name);
        embed.setDescription(character.description);
        embed.setThumbnail(character.image);
        embed.addField("Homeworld", character.homeworld);
        embed.addField("Species", character.species);
        embed.addField("Affiliations", character.affiliations.join(", "));
        embed.addField("Masters", character.masters.join(", "));
        embed.addField("Apprentices", character.apprentices.join(", "));

        cmd.replyToChannelEmbed(embed.compile());

    } else if (type === "location") {
        url += "locations/search/" + encodeURIComponent(query);

        var content = HTTP.get(10, url, {});
        if (content === false) {
            cmd.replyToAuthor("Something went wrong while searching for that location.");
            return;
        }

        var json = JSON.parse(content);
        if (json === undefined || json.locations.length === 0) {
            cmd.replyToAuthor("Could not find a location with that name.");
            return;
        }

        var location = json.locations[0];

        var embed = new EmbedMessage();
        embed.setTitle(location.name);
        embed.setDescription(location.description);
        embed.setThumbnail(location.image);
        embed.addField("Climate", location.climate);
        embed.addField("Terrain", location.terrain);
        embed.addField("Notable Inhabitants", location.notable_inhabitants.join(", "));

        cmd.replyToChannelEmbed(embed.compile());
    } else {
        cmd.replyToAuthor("Invalid search type. Use `character` or `location`.");
    }
});
