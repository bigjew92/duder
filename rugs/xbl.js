var xbl = new DuderRug("Xbox Live", "Xbox Live stuff");
xbl.storage = xbl.loadStorage();

xbl.getAPIKey = function() {
	if (this.storage.settings === undefined) {
		return false;
	} else if (this.storage.settings.api_key === undefined) {
		return false;
	}
	return this.storage.settings.api_key;
};

xbl.setAPIKey = function(key) {
	if (this.storage.settings === undefined) {
		this.storage.settings = {};
	}
	this.storage.settings.api_key = key;
	this.saveStorage(this.storage);
};

xbl.getUserXUID = function(guildID, userID) {
	if (this.storage[guildID] === undefined) {
		return false;
	} else if (this.storage[guildID].users === undefined) {
		return false;
	}
	for (var i = 0; i < this.storage[guildID].users.length; i++) {
		var user = this.storage[guildID].users[i];
		if (user.userID === userID) {
			return user.XUID;
		}
	}
	return false;
};

xbl.setUserXUID = function(guildID, userID, XUID) {
	if (this.storage[guildID] === undefined) {
		this.storage[guildID] = {};
	}
	if (this.storage[guildID].users === undefined) {
		this.storage[guildID].users = [];
	}
	var found = false;
	for (var i = 0; i < this.storage[guildID].users.length; i++) {
		if (this.storage[guildID].users[i].userID === userID) {
			this.storage[guildID].users[i].XUID = XUID;
			found = true;
			break;
		}
	}
	if (!found) {
		this.storage[guildID].users.push({ userID: userID, XUID: XUID });
	}
	this.saveStorage(this.storage);
};

xbl.getUserProfileCache = function(guildID, userID) {
	if (this.storage[guildID] === undefined) {
		return false;
	} else if (this.storage[guildID].users === undefined) {
		return false;
	}
	for (var i = 0; i < this.storage[guildID].users.length; i++) {
		var user = this.storage[guildID].users[i];
		if (user.userID === userID) {
			if (user.profileCache === undefined) {
				return false;
			}
			return user.XUID;
		}
	}
	return false;
};

xbl.setUserProfileCache = function(guildID, XUID, profileCache) {
	if (this.storage[guildID] === undefined) {
		this.storage[guildID] = {};
	}
	if (this.storage[guildID].users === undefined) {
		this.storage[guildID].users = [];
	}
	var found = false;
	for (var i = 0; i < this.storage[guildID].users.length; i++) {
		if (this.storage[guildID].users[i].XUID === XUID) {
			this.storage[guildID].users[i].profileCache = profileCache;
			this.storage[guildID].users[i].profileCacheUpdated = Date.now();
			found = true;
			break;
		}
	}
	if (!found) {
		this.storage[guildID].users.push({
			XUID: XUID,
			profileCache: profileCache,
			profileCacheUpdated: Date.now()
		});
	}
	this.saveStorage(this.storage);
};

xbl.getUserProfileCache = function(guildID, XUID) {
	if (this.storage[guildID] === undefined) {
		return false;
	} else if (this.storage[guildID].users === undefined) {
		return false;
	}

	for (var i = 0; i < this.storage[guildID].users.length; i++) {
		var user = this.storage[guildID].users[i];
		if (user.XUID === XUID) {
			if (
				user.profileCache === undefined ||
				user.profileCacheUpdated === undefined
			) {
				return false;
			}
			var age = Date.now() - user.profileCacheUpdated;
			// convert to minutes
			age /= 1000 * 60 * 60;
			this.dprint("profile cache is {0} hours old".format(age));
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

xbl.addCommand("xbl", function(cmd) {
	var action = cmd.args.length > 1 ? cmd.args[1].toLowerCase() : "me";
	var XUID = false;
	var content = "";
	var json = null;
	var profile = null;

	if (action === "setkey") {
		if (cmd.args.length === 3) {
			this.setAPIKey(cmd.args[2]);
			cmd.replyToAuthor("XboxAPI API key has been saved.");
			return;
		} else {
			cmd.replyToAuthor("usage: `setkey YOUR_API_KEY`");
			return;
		}
	}

	var api_key = this.getAPIKey();
	if (api_key === false) {
		this.wprint("bad api key!");
		this.setAPIKey("API_KEY_HERE");
		return;
	}

	var headers = {};
	headers["X-Authorization"] = api_key;

	if (action === "me") {
		XUID = this.getUserXUID(cmd.guildID, cmd.author.id);
		if (XUID === false) {
			cmd.replyToAuthor("you need to linked your gamertag");
			return;
		}
		action = "profile";
	}

	if (action === "link") {
		if (cmd.args.length < 3) {
			cmd.replyToAuthor('usage: `xbl link "YourGamerTag"`');
			return;
		}
		Duder.startTyping(cmd.channelID);
		var gamertag = cmd.args[2];
		var url = "https://xbl.io/api/v2/friends/search?gt={0}".format(
			gamertag.replaceAll(" ", "%20")
		);
		this.dprint(url);
		var content = HTTP.get(10, url, headers);
		//this.dprint(content);
		json = JSON.parse(content);
		if (json.profileUsers === undefined) {
			cmd.replyToAuthor(
				"unable to link gamertag **{0}**".format(gamertag)
			);
		}
		profile = json.profileUsers[0];
		for (var i = 0; i < profile.settings.length; i++) {
			var setting = profile.settings[i];
			if (setting.id === "Gamertag" && setting.value === gamertag) {
				this.setUserXUID(cmd.guildID, cmd.author.id, profile.id);
				this.setUserProfileCache(
					cmd.guildID,
					profile.id,
					this.parseProfileSettings(profile.settings)
				);
				cmd.replyToAuthor("link successful :link:");
				return;
			}
		}
		cmd.replyToAuthor("unable to match gamertag **{0}**".format(gamertag));
		return;
	} else if (action === "profile") {
		if (XUID === false) {
			if (cmd.mentions.length !== 1) {
				cmd.replyToAuthor('usage: `xbl profile "@mention"`');
				return;
			}
			XUID = this.getUserXUID(cmd.guildID, cmd.mentions[0].id);
			if (XUID === false) {
				cmd.replyToAuthor(
					"{0} hasn't linked their gamertag".format(
						cmd.mentions[0].username
					)
				);
				return;
			}
		}

		Duder.startTyping(cmd.channelID);
		profile = this.getUserProfileCache(cmd.guildID, XUID);
		if (profile === false) {
			this.dprint("requesting profile");
			var url = "https://xbl.io/api/v2/account/{0}".format(XUID);
			var content = HTTP.get(10, url, headers);
			if (content === undefined || content.length === 0) {
				cmd.replyToAuthor("unable to retrieve profile");
				return;
			}
			this.dprint(content);
			json = JSON.parse(content);
			if (json.profileUsers === undefined) {
				cmd.replyToAuthor("unable to parse profile");
				return;
			}
			profile = this.parseProfileSettings(json.profileUsers[0].settings);
			this.setUserProfileCache(cmd.guildID, XUID, profile);
		} else {
			this.dprint("using profile cache");
		}
		var embed = new EmbedMessage();
		embed.setTitle("{0}'s Profile".format(profile.Gamertag));
		embed.setColor(1080336);
		embed.setThumbnail(profile.GameDisplayPicRaw);
		embed.setDescription(profile.Bio);
		embed.setFooter(profile.GameDisplayPicRaw, profile.Location);
		embed.addField(":trophy: Gamerscore", profile.Gamerscore);
		cmd.replyToChannelEmbed(embed.compile());
	} else if (action === "list") {
		if (
			this.storage[cmd.guildID] === undefined ||
			this.storage[cmd.guildID].users === undefined
		) {
			cmd.replyToAuthor("no one has linked their accounts :confused:");
		} else {
			var users = [];
			for (var i = 0; i < this.storage[cmd.guildID].users.length; i++) {
				var user = this.storage[cmd.guildID].users[i];
				users.push(user.XUID);
			}
			if (users.length > 0) {
				var url = "https://xbl.io/api/v1/[{0}]/presence".format(
					users.join(",")
				);
				//url = "https://xbl.io/api/v2/presence";
				//url = encodeURI(url);
				this.dprint(url);
			}
			var content = HTTP.get(10, url, headers);
			if (content === undefined || content.length === 0) {
				cmd.replyToAuthor("unable to retrieve presence");
				return;
			}
			this.dprint(content);
			//json = JSON.parse(content);
		}
	}
});
