var lastseen = new DuderRug("Last Seen", "Tracks people");
lastseen.storage = lastseen.loadStorage();

lastseen.updateUser = function(userID) {
	if (this.storage.users === undefined) {
		this.storage.users = {};
	}

	this.storage.users[userID] = {
		lastSeen: Date.now()
	};

	this.saveStorage(this.storage);
};

lastseen.getUser = function(userID) {
	if (this.storage.users === undefined) {
		return false;
	} else if (this.storage.users[userID] === undefined) {
		return false;
	}

	return this.storage.users[userID];
};

lastseen.onMessage(function(msg) {
	this.updateUser(msg.author.id);
});

lastseen.onPresenceUpdate(function(guildID, user, presence) {
	this.updateUser(user.id);
	//this.dprint("guild {0} username {1} presence {2}".format(guildID, user.username, presence));
});

lastseen.addCommand("lastseen", function(cmd) {
	var id = "";

	if (cmd.mentions.length !== 1) {
		if (cmd.args.length === 2) {
			id = DuderUser.getIDbyNickname(cmd.guildID, cmd.args[1]);
			if (id === false) {
				cmd.replyToAuthor("who?");
				return;
			}
		} else {
			cmd.replyToAuthor("usage: `lastseen @mention`.");
			return;
		}
	} else {
		id = cmd.mentions[0].id;
	}

	//this.dprint("looking for id {0}".format(id));

	var data = this.getUser(id);
	if (data === false) {
		cmd.replyToChannel("haven't seent 'em");
		return;
	}

	var count = 0;
	var unit = "";

	// time stored in milliseconds
	var age = Date.now() - data.lastSeen;
	// convert to minutes
	var minutes = age / (1000 * 60);
	if (minutes < 1440) {
		if (minutes < 60) {
			count = minutes;
			unit = "minute";
		} else {
			count = minutes / 60;
			unit = "hour";
		}
	} else {
		count = minutes / (60 * 24);
		unit = "day";
	}

	count = Math.floor(count);
	var plural = count !== 1 ? "s" : "";
	cmd.replyToChannel("Last seen {0} {1}{2} ago".format(count, unit, plural));
	/*
	var embed = new EmbedMessage();
	embed.setTitle("{0} title".format(who.username));
	embed.setColor(1080336);
	//embed.setThumbnail(profile.GameDisplayPicRaw);
	embed.setDescription("desc");
	embed.setFooter("last seen {0} {1} ago".format(count, unit));
	embed.addField("Message", "`{0}`".format(data.content));
	cmd.replyToChannelEmbed(embed.compile());
	*/
});
