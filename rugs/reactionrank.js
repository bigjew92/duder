var reactionrank = new DuderRug("Reaction Rank", "Tracks reactions");
reactionrank.storage = reactionrank.loadStorage();

reactionrank.modifyUserEmoji = function(guildID, userID, emojiID, delta) {
	if (this.storage.guilds === undefined) {
		this.storage.guilds = {};
	}
	if (this.storage.guilds[guildID] === undefined) {
		this.storage.guilds[guildID] = {};
	}
	if (this.storage.guilds[guildID][userID] === undefined) {
		this.storage.guilds[guildID][userID] = {};
	}
	if (this.storage.guilds[guildID][userID].rated === undefined) {
		this.storage.guilds[guildID][userID].rated = {};
	}
	if (this.storage.guilds[guildID][userID].rated[emojiID] === undefined) {
		this.storage.guilds[guildID][userID].rated[emojiID] = 0;
	}
	if (this.storage.guilds[guildID][userID].total === undefined) {
		this.storage.guilds[guildID][userID].total = 0;
	}

	this.storage.guilds[guildID][userID].rated[emojiID] = Math.max(
		this.storage.guilds[guildID][userID].rated[emojiID] + delta,
		0
	);
	this.storage.guilds[guildID][userID].total = Math.max(this.storage.guilds[guildID][userID].total + delta, 0);

	this.saveStorage(this.storage);
};

reactionrank.addUserXP = function(guildID, userID) {
	if (this.storage.guilds === undefined) {
		this.storage.guilds = {};
	}
	if (this.storage.guilds[guildID] === undefined) {
		this.storage.guilds[guildID] = {};
	}
	if (this.storage.guilds[guildID][userID] === undefined) {
		this.storage.guilds[guildID][userID] = {};
	}
	if (this.storage.guilds[guildID][userID].xp === undefined) {
		this.storage.guilds[guildID][userID].xp = 0;
	}
	if (this.storage.guilds[guildID][userID].lastXP === undefined) {
		this.storage.guilds[guildID][userID].lastXP = 0;
	}
	if (this.storage.guilds[guildID][userID].level === undefined) {
		this.storage.guilds[guildID][userID].level = 0;
	}
	if (this.storage.guilds[guildID][userID].nextLevel === undefined) {
		this.storage.guilds[guildID][userID].nextLevel = 10;
	}

	// xp every minute
	var age = Date.now() - this.storage.guilds[guildID][userID].lastXP;
	// convert to minutes
	age /= 1000 * 60;
	if (age < 1) {
		return;
	}

	this.storage.guilds[guildID][userID].xp = this.storage.guilds[guildID][userID].xp + 1;
	this.storage.guilds[guildID][userID].lastXP = Date.now();

	// check for level
	if (this.storage.guilds[guildID][userID].xp >= this.storage.guilds[guildID][userID].nextLevel) {
		this.storage.guilds[guildID][userID].level = this.storage.guilds[guildID][userID].level + 1;
		this.storage.guilds[guildID][userID].nextLevel = Math.ceil(
			this.storage.guilds[guildID][userID].nextLevel * 1.1
		);
		this.storage.guilds[guildID][userID].xp = 0;
	}

	this.saveStorage(this.storage);
};

reactionrank.getUserXP = function(guildID, userID) {
	if (this.storage.guilds === undefined) {
		return 0;
	}
	if (this.storage.guilds[guildID] === undefined) {
		return 0;
	}
	if (this.storage.guilds[guildID][userID] === undefined) {
		return 0;
	}
	if (this.storage.guilds[guildID][userID].xp === undefined) {
		return 0;
	}

	return this.storage.guilds[guildID][userID].xp;
};

reactionrank.getUserLevel = function(guildID, userID) {
	if (this.storage.guilds === undefined) {
		return 0;
	}
	if (this.storage.guilds[guildID] === undefined) {
		return 0;
	}
	if (this.storage.guilds[guildID][userID] === undefined) {
		return 0;
	}
	if (this.storage.guilds[guildID][userID].level === undefined) {
		return 0;
	}

	return this.storage.guilds[guildID][userID].level;
};

reactionrank.getUserNextLevel = function(guildID, userID) {
	if (this.storage.guilds === undefined) {
		return 0;
	}
	if (this.storage.guilds[guildID] === undefined) {
		return 0;
	}
	if (this.storage.guilds[guildID][userID] === undefined) {
		return 0;
	}
	if (this.storage.guilds[guildID][userID].nextLevel === undefined) {
		return 0;
	}

	return this.storage.guilds[guildID][userID].nextLevel;
};

reactionrank.onMessageReactionAdd(function(reaction) {
	//this.dprint_r(reaction);
	if (reaction.instigator.id === reaction.message.author.id) {
		return;
	}
	//this.dprint("adding");
	this.modifyUserEmoji(reaction.guildID, reaction.message.author.id, reaction.emoji.name, 1);

	// add xp
	var oldLevel = this.getUserLevel(reaction.guildID, reaction.instigator.id);
	this.addUserXP(reaction.guildID, reaction.instigator.id);
	var newLevel = this.getUserLevel(reaction.guildID, reaction.instigator.id);

	// check for new level
	if (newLevel > oldLevel) {
		reaction.replyToChannel(
			":fireworks: **{0} has reach level {1}!**".format(reaction.instigator.username, newLevel)
		);
	}
});

reactionrank.onMessageReactionRemove(function(reaction) {
	//this.dprint_r(reaction);
	if (reaction.instigator.id === reaction.message.author.id) {
		return;
	}
	//this.dprint("removing");
	this.modifyUserEmoji(reaction.guildID, reaction.message.author.id, reaction.emoji.name, -1);
});


reactionrank.addCommand("rank", function(cmd) {
	var userID = cmd.author.id;

	var xp = this.getUserXP(cmd.guildID, userID);
	var level = this.getUserLevel(cmd.guildID, userID);
	var nextLevel = this.getUserNextLevel(cmd.guildID, userID);

	var percent = Math.floor((xp / nextLevel) * 100);

	cmd.replyToAuthor("you are currently level {0} and {1}% to your next level".format(level, percent));
});
