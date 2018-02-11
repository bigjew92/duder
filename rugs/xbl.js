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
	if (this.storage[cmd.guildID] == undefined) {
		return false;
	} else if (this.storage[cmd.guildID].users == undefined) {
		return false;
	}
	for (var i = 0; i < this.storage[cmd.guildID].users.length; i++) {
		var user = this.storage[cmd.guildID].users[i];
		if (user.userID == userID) {
			return user.XUID;
		}
	}
	return false;
};

xbl.setUserXUID = function(userID, XUID) {
	if (this.storage[cmd.guildID] == undefined) {
		this.storage[cmd.guildID] = {};
	}
	if (this.storage[cmd.guildID].users == undefined) {
		this.storage[cmd.guildID].users = [];
	}
	var found = false;
	for (var i = 0; i < this.storage[cmd.guildID].users.length; i++) {
		if (this.storage[cmd.guildID].users[i].userID == userID) {
			this.storage[cmd.guildID].users[i].XUID = XUID;
			found = true;
			break;
		}
	}
	if (!found) {
		this.storage[cmd.guildID].users.push({ userID: userID, XUID: XUID });
	}
	rug.saveStorage(this.storage);
};

xbl.getUserProfileCache = function(userID) {
	if (this.storage[cmd.guildID] == undefined) {
		return false;
	} else if (this.storage[cmd.guildID].users == undefined) {
		return false;
	}
	for (var i = 0; i < this.storage[cmd.guildID].users.length; i++) {
		var user = this.storage[cmd.guildID].users[i];
		if (user.userID == userID) {
			if (user.profileCache == undefined) {
				return false;
			}
			return user.XUID;
		}
	}
	return false;
};

xbl.setUserProfileCache = function(XUID, profileCache) {
	if (this.storage[cmd.guildID] == undefined) {
		this.storage[cmd.guildID] = {};
	}
	if (this.storage[cmd.guildID].users == undefined) {
		this.storage[cmd.guildID].users = [];
	}
	var found = false;
	for (var i = 0; i < this.storage[cmd.guildID].users.length; i++) {
		if (this.storage[cmd.guildID].users[i].XUID == XUID) {
			this.storage[cmd.guildID].users[i].profileCache = profileCache;
			this.storage[cmd.guildID].users[i].profileCacheUpdated = Date.now();
			found = true;
			break;
		}
	}
	if (!found) {
		this.storage[cmd.guildID].users.push({
			XUID: XUID,
			profileCache: profileCache,
			profileCacheUpdated: Date.now()
		});
	}
	rug.saveStorage(this.storage);
};

xbl.getUserProfileCache = function(XUID) {
	if (this.storage[cmd.guildID] == undefined) {
		return false;
	} else if (this.storage[cmd.guildID].users == undefined) {
		return false;
	}

	for (var i = 0; i < this.storage[cmd.guildID].users.length; i++) {
		var user = this.storage[cmd.guildID].users[i];
		if (user.XUID == XUID) {
			if (
				user.profileCache == undefined ||
				user.profileCacheUpdated == undefined
			) {
				return false;
			}
			var age = Date.now() - user.profileCacheUpdated;
			// convert to minutes
			age /= 1000 * 60 * 60;
			rug.dprint("profile cache is {0} hours old".format(age));
			return age > 1 ? false : user.profileCache;
		}
	}
	return false;
};

xbl.parseProfileSettings = function(profileSettings) {
	var profile = {};
	for (var i = 0; i < profileSettings.length; i++) {
		var setting = profileSettings[i];
		profile[setting.id] = setting.value;
	}
	return profile;
};

xbl.addCommand("xbl", function() {
	var action = cmd.args.length > 1 ? cmd.args[1].toLowerCase() : "me";
	var XUID = false;
	var url = "";
	var content = "";
	var json = null;
	var profile = null;

	if (action == "setkey") {
		if (cmd.args.length == 3) {
			rug.setAPIKey(cmd.args[2]);
			cmd.replyToAuthor("XboxAPI API key has been saved.");
			return;
		} else {
			cmd.replyToAuthor("usage: `setkey YOUR_API_KEY`");
			return;
		}
	}

	var api_key = rug.getAPIKey();
	if (api_key == false) {
		rug.wprint("bad api key!");
		rug.setAPIKey("API_KEY_HERE");
		return;
	}

	var headers = {};
	headers["X-Authorization"] = api_key;

	if (action == "me") {
		XUID = rug.getUserXUID(cmd.author.id);
		if (XUID == false) {
			cmd.replyToAuthor("you need to linked your gamertag");
			return;
		}
		action = "profile";
	}

	//curl --header "X-Authorization: 0kswckkwccs4ck8cc48kosok80w4woc8kow" https://xbl.io/api/v2/account

	if (action == "link") {
		if (cmd.args.length < 3) {
			cmd.replyToAuthor('usage: `xbl link "YourGamerTag"`');
			return;
		}
		Duder.startTyping(cmd.channelID);
		var gamertag = cmd.args[2];
		url = "https://xbl.io/api/v2/friends/search?gt={0}".format(gamertag.replaceAll(" ","%20"));
		rug.dprint(url);
		content = HTTP.get(10, url, headers);
		//rug.print(content);
		json = JSON.parse(content);
		if (json.profileUsers == undefined) {
			cmd.replyToAuthor(
				"unable to link gamertag **{0}**".format(gamertag)
			);
		}
		profile = json.profileUsers[0];
		for (var i = 0; i < profile.settings.length; i++) {
			var setting = profile.settings[i];
			if (setting.id == "Gamertag" && setting.value == gamertag) {
				rug.setUserXUID(cmd.author.id, profile.id);
				rug.setUserProfileCache(
					profile.id,
					rug.parseProfileSettings(profile.settings)
				);
				cmd.replyToAuthor("link successful :link:");
				return;
			}
		}
		cmd.replyToAuthor("unable to match gamertag **{0}**".format(gamertag));
		return;
	} else if (action == "profile") {
		if (XUID == false) {
			if (cmd.mentions.length != 1) {
				cmd.replyToAuthor('usage: `xbl profile "@mention"`');
				return;
			}
			XUID = rug.getUserXUID(cmd.mentions[0].id);
			if (XUID == false) {
				cmd.replyToAuthor(
					"{0} hasn't linked their gamertag".format(
						cmd.mentions[0].username
					)
				);
				return;
			}
		}
		Duder.startTyping(cmd.channelID);
		profile = rug.getUserProfileCache(XUID);
		if (json == false) {
			rug.dprint("requesting profile");
			url = "https://xbl.io/api/v2/account/{0}".format(XUID);
			content = HTTP.get(10, url, headers);
			json = JSON.parse(content);
			if (json.profileUsers == undefined) {
				cmd.replyToAuthor("unable to retrieve profile");
				return;
			}
			profile = parseProfileSettings(json.profileUsers[0].settings);
		} else {
			rug.dprint("using profile cache");
		}
		var embed = new EmbedMessage();
		embed.setTitle("{0}'s Profile".format(profile.Gamertag));
		embed.setColor(1080336);
		embed.setThumbnail(profile.GameDisplayPicRaw);
		embed.setDescription(profile.Bio);
		embed.setFooter(profile.GameDisplayPicRaw, profile.Location);
		embed.addField(":trophy: Gamerscore", profile.Gamerscore);
		embed.addField(
			":military_medal: Tenure",
			"{0} years".format(profile.TenureLevel)
		);
		cmd.replyToChannelEmbed(embed.compile());
	} else if (action == "list") {
		if (
			rug.storage[cmd.guildID] == undefined ||
			rug.storage[cmd.guildID].users == undefined
		) {
			cmd.replyToAuthor("no one has linked their accounts :confused:");
		} else {
			for (var i = 0; i < rug.storage[cmd.guildID].users.length; i++) {
				url = "https://xboxapi.com/v2/{0}/profile".format(XUID);
			}
		}
	}
});
