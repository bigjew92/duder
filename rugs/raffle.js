var raffle = new DuderRug("Raffle", "Make raffles.");
raffle.storage = raffle.loadStorage();

raffle.getActive = function(guildID) {
	if (this.storage[guildID] === undefined) {
		return false;
	} else if (this.storage[guildID].active === undefined) {
		return false;
	}
	return this.storage[guildID].active;
};

raffle.setActive = function(guildID, active) {
	if (this.storage[guildID] === undefined) {
		this.storage[guildID] = {};
	}
	this.storage[guildID].active = active;
	this.saveStorage(this.storage);
};

raffle.getDescription = function(guildID) {
	if (this.storage[guildID] === undefined) {
		return "No description";
	} else if (this.storage[guildID].description === undefined) {
		return "No description";
	}
	return this.storage[guildID].description;
};

raffle.setDescription = function(guildID, description) {
	if (this.storage[guildID] === undefined) {
		this.storage[guildID] = {};
	}
	this.storage[guildID].description = description;
	this.saveStorage(this.storage);
};

raffle.getCreator = function(guildID) {
	if (this.storage[guildID] === undefined) {
		return "Unknown";
	} else if (this.storage[guildID].creator === undefined) {
		return "Unknown";
	}
	return this.storage[guildID].creator;
};

raffle.setCreator = function(guildID, creatorID) {
	if (this.storage[guildID] === undefined) {
		this.storage[guildID] = {};
	}
	this.storage[guildID].creator = creatorID;
	this.saveStorage(this.storage);
};

raffle.addParticipant = function(guildID, participantID) {
	if (this.storage[guildID] === undefined) {
		this.storage[guildID] = {};
	}
	if (this.storage[guildID].participants === undefined) {
		this.storage[guildID].participants = [];
	}
	this.storage[guildID].participants.push(participantID);
	this.saveStorage(this.storage);
};

raffle.hasParticipant = function(guildID, participantID) {
	if (this.storage[guildID] === undefined) {
		return false;
	}
	if (this.storage[guildID].participants === undefined) {
		return false;
	}
	for (var i = 0; i < this.storage[guildID].participants.length; i++) {
		if (this.storage[guildID].participants[i] === participantID) {
			return true;
		}
	}
	return false;
};

raffle.participantCount = function(guildID) {
	if (this.storage[guildID] === undefined) {
		return 0;
	}
	if (this.storage[guildID].participants === undefined) {
		return 0;
	}
	return this.storage[guildID].participants.length;
};

raffle.clearParticipants = function(guildID) {
	if (this.storage[guildID] === undefined) {
		this.storage[guildID] = {};
	}
	this.storage[guildID].participants = [];
	this.saveStorage(this.storage);
};

raffle.addCommand("raffle", function(cmd) {
	var action = cmd.args.length > 1 ? cmd.args[1].toLowerCase() : "status";

	if (action === "status") {
		if (this.getActive(cmd.guildID) === true) {
			var joined = this.hasParticipant(cmd.guildID, cmd.author.id)
				? "already"
				: "not";
			cmd.replyToAuthor(
				"you have " +
					joined +
					' joined the raffle *"' +
					this.getDescription(cmd.guildID) +
					'"* which currently has ' +
					this.participantCount(cmd.guildID) +
					" participant(s)."
			);
			return;
		} else {
			cmd.replyToAuthor("there aren't any active raffles.");
			return;
		}
	} else if (action === "start" || action === "new") {
		if (this.getActive(cmd.guildID) === true) {
			cmd.replyToAuthor(
				'the raffle *"' +
					this.getDescription(cmd.guildID) +
					"\"* hasn't finished yet."
			);
			return;
		} else if (cmd.args.length < 3) {
			cmd.replyToAuthor(
				'you need to provide a description `raffle start "Awesome raffle!"`.'
			);
			return;
		}
		this.setActive(cmd.guildID, true);
		this.setCreator(cmd.guildID, cmd.author.id);
		this.setDescription(cmd.guildID, cmd.args[2]);
		this.clearParticipants(cmd.guildID);
		cmd.replyToChannel(
			cmd.author.username +
				' has started raffle *"' +
				this.getDescription(cmd.guildID) +
				'"*.'
		);
	} else if (action === "join" || action === "j" || action === "enter") {
		if (this.getActive(cmd.guildID) === false) {
			cmd.replyToAuthor("there aren't any raffles to join.");
			return;
		} else if (this.hasParticipant(cmd.author.id) === true) {
			cmd.replyToAuthor("you've already joined this raffle.");
			return;
		}
		this.addParticipant(cmd.guildID, cmd.author.id);
		cmd.replyToAuthor("you joined the raffle :tickets:");
	} else if (action === "finish" || action === "end") {
		if (this.getActive(cmd.guildID) === false) {
			cmd.replyToAuthor("there aren't any raffles to finish.");
			return;
		} else if (cmd.author.id !== this.getCreator(cmd.guildID)) {
			cmd.replyToAuthor("you cannot end this raffle.");
			return;
		}

		var msg =
			'The raffle *"' +
			this.getDescription(cmd.guildID) +
			'"* has ended! ';

		if (this.participantCount(cmd.guildID) === 0) {
			msg +=
				"but there weren't any participants so no one wins :face_palm:";
		} else {
			var r = Math.getRandomInRange(
				0,
				this.storage[cmd.guildID].participants.length - 1
			);
			var winnerID = this.storage[cmd.guildID].participants[r];
			msg += "<@" + winnerID + "> won :rotating_light::fireworks::mega:";
		}

		cmd.replyToChannel(msg);
		this.setActive(cmd.guildID, false);
	}
});
