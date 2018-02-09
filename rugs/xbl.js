var xbl = new DuderRug("Xbox Live", "Xbox Live stuff");
xbl.storage = xbl.loadStorage();

xbl.getAPIKey = function() {
    if (this.storage.settings == undefined) {
        return false;
    } else if (this.storage.settings.api_key == undefined) {
        return false;
    }
	return this.storage.settings.api_key;
};

xbl.setAPIKey = function(key) {
    if (this.storage.settings == undefined) {
        this.storage.settings = {};
    }
    this.storage.settings.api_key = key;
	rug.saveStorage(this.storage);
};

xbl.getUserXUID = function(userID) {
	if (this.storage.users == undefined) {
		return false;
	}
	for (var i = 0; i < this.storage.users.length; i++) {
		var user = this.storage.users[i];
		if (user.userID == userID) {
			return user.XUID;
		}
	}
	return false;
};

xbl.setUserXUID = function(userID, ID) {
	if (this.storage.users == undefined) {
		this.storage.users = [];
	}
	var found = false;
	for (var i = 0; i < this.storage.users.length; i++) {
		if (this.storage.users[i].userID == userID) {
			this.storage.users[i].XUID = ID;
			found = true;
			break;
		}
	}
	if (!found) {
		this.storage.users.push({ userID: userID, XUID: ID });
	}
	rug.saveStorage(this.storage);
};

xbl.addCommand("xbl", function() {
	var api_key = rug.getAPIKey();
	if (api_key == false) {
		dprint("bad api key!");
		rug.setAPIKey("XBOXAPI_KEY_HERE");
	}

	var headers = {};
	headers['X-AUTH'] = api_key;

	var action = (cmd.args.length > 1) ? cmd.args[1].toLowerCase() : "me";
	var profileId = false;

	if (action == "me") {
		profileId = rug.getUserXUID(cmd.author.id);
		if (profileId == false) {
			cmd.replyToAuthor("you need to linked your gamertag");
			return;
		}
		action = "profile";
	}

	if (action == "link") {
		if (cmd.args.length < 3) {
			cmd.replyToAuthor("usage: `xbl link \"YourGamerTag\"`");
			return;
		}
		Duder.startTyping(cmd.channelID);
		var gt = cmd.args[2];
		var url = "https://xboxapi.com/v2/xuid/{0}".format(cmd.args[2]);
		dprint(url);
		var content = HTTP.get(10, url, headers);
		dprint(content);
		if (content.indexOf("error") != -1) {
			cmd.replyToAuthor("gamertag **{0}** wasn't found :confused:".format(gt));
			return;
		}
		rug.setUserXUID(cmd.author.id, content);
		cmd.replyToAuthor("link successful :link:");
		return;
	} else if (action == "profile") {
		/*
		var embed = new EmbedMessage();
		embed.setTitle("Hello, World!");
		embed.setDescription("This is a description, ok?");
		embed.setURL("http://google.com");
		embed.setColor(1080336);
		embed.setTimestamp();
		embed.setFooter("https://i.imgur.com/R6SUWFz.jpg","whos this?");
		embed.setThumbnail("https://i.imgur.com/R6SUWFz.jpg");
		embed.setImage("https://i.imgur.com/R6SUWFz.jpg");
		embed.setAuthor("foszor","http://foszor.com","https://i.imgur.com/R6SUWFz.jpg");
		embed.addField("Gamerscore", "over 9000");
		embed.addField("Account Tier", "GOLD");
		var c = embed.compile();
		dprint(c);
		cmd.replyToChannelEmbed(c);
		*/

		if (profileId == false) {
			if(cmd.mentions.length != 1) {
				cmd.replyToAuthor("usage: `xbl profile \"@mention\"`");
				return;
			}
			profileId = rug.getUserXUID(cmd.mentions[0].id);
			if (profileId == false) {
				cmd.replyToAuthor("{0} hasn't linked their gamertag".format(cmd.mentions[0].username));
				return;
			}
		}
		Duder.startTyping(cmd.channelID);
		var url = "https://xboxapi.com/v2/{0}/profile".format(profileId);
		dprint(url);
		var content = HTTP.get(10, url, headers);
		dprint(content);
		var json = JSON.parse(content);
		dprint(json);
		var embed = new EmbedMessage();
		embed.setTitle("{0}'s Profile".format(json.Gamertag));
		embed.setColor(1080336);
		embed.setThumbnail(json.GameDisplayPicRaw.decodeHTML());
		embed.addField(":trophy: Gamerscore", json.Gamerscore);
		embed.addField(":star: Tier", json.AccountTier);
		embed.addField(":yellow_heart: Reputation", json.XboxOneRep);
		embed.addField(":large_orange_diamond: Tenure", "{0} years".format(json.TenureLevel));
		cmd.replyToChannelEmbed(embed.compile());	
	}
	/*
	// 2533274816830336
	var url = "https://xboxapi.com/v2/2535465925875131/screenshots";
	var headers = {};
	headers['X-AUTH'] = "94fa49ed6d92aa6227e124a9f75d47c0eef4b0c9";
	var content = HTTP.get(4, url, headers);
	dprint(content);
	var json = JSON.parse(content);

	var xuid = json.xuid;
	url = "https://xboxapi.com/v2/{0}/profile".format(xuid);
	dprint(url);
	*/
});
