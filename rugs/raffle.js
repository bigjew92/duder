var raffle = new DuderRug("Raffle", "Make raffles.");
raffle.storage = raffle.loadStorage();

raffle.getActive = function() {
    if (this.storage[cmd.channelID] == undefined) {
        return false;
    } else if (this.storage[cmd.channelID]['active'] == undefined) {
        return false;
    }
    return this.storage[cmd.channelID]['active'];
}

raffle.setActive = function(active) {
    if (this.storage[cmd.channelID] == undefined) {
        this.storage[cmd.channelID] = {};
    }
    this.storage[cmd.channelID]['active'] = active;
    rug.saveStorage(this.storage)
}

raffle.getDescription = function() {
    if (this.storage[cmd.channelID] == undefined) {
        return "No description";
    } else if (this.storage[cmd.channelID]['description'] == undefined) {
        return "No description";
    }
    return this.storage[cmd.channelID]['description'];
}

raffle.setDescription = function(description) {
    if (this.storage[cmd.channelID] == undefined) {
        this.storage[cmd.channelID] = {};
    }
    this.storage[cmd.channelID]['description'] = description;
    rug.saveStorage(this.storage)
}

raffle.getCreator = function() {
    if (this.storage[cmd.channelID] == undefined) {
        return "Unknown";
    } else if (this.storage[cmd.channelID]['creator'] == undefined) {
        return "Unknown";
    }
    return this.storage[cmd.channelID]['creator'];
}

raffle.setCreator = function(creatorID) {
    if (this.storage[cmd.channelID] == undefined) {
        this.storage[cmd.channelID] = {};
    }
    this.storage[cmd.channelID]['creator'] = creatorID;
    rug.saveStorage(this.storage)
}

raffle.addParticipant = function(participantID) {
    if (this.storage[cmd.channelID] == undefined) {
        this.storage[cmd.channelID] = {};
    }
    if (this.storage[cmd.channelID]['participants'] == undefined) {
        this.storage[cmd.channelID]['participants'] = [];
    }
    this.storage[cmd.channelID]['participants'].push(participantID);
    rug.saveStorage(this.storage)
}

raffle.hasParticipant = function(participantID) {
    if (this.storage[cmd.channelID] == undefined) {
        return false;
    }
    if (this.storage[cmd.channelID]['participants'] == undefined) {
        return false;
    }
    for (var i = 0; i < this.storage[cmd.channelID]['participants'].length; i++) {
        if (this.storage[cmd.channelID]['participants'][i] == participantID) {
            return true;
        }
    }
    return false;
}

raffle.participantCount = function() {
    if (this.storage[cmd.channelID] == undefined) {
        return 0;
    }
    if (this.storage[cmd.channelID]['participants'] == undefined) {
        return 0;
    }
    return this.storage[cmd.channelID]['participants'].length;
}

raffle.clearParticipants = function() {
    if (this.storage[cmd.channelID] == undefined) {
        this.storage[cmd.channelID] = {};
    }
    this.storage[cmd.channelID]['participants'] = [];
    rug.saveStorage(this.storage)
}

raffle.addCommand("raffle", function() {
    var action = (cmd.args.length > 1) ? cmd.args[1].toLowerCase() : "status";

    if (action == "status") {
        if (rug.getActive() == true) {
            var joined = (rug.hasParticipant(cmd.author.id)) ? "already" : "not";
            cmd.replyToAuthor("you have " + joined + " joined the raffle *\"" + rug.getDescription() + "\"* which currently has " + rug.participantCount() + " participant(s).");
            return;
        } else {
            cmd.replyToAuthor("there aren't any active raffles.");
            return;
        }
    } else if (action == "start" || action == "new") {
        if (rug.getActive() == true) {
            cmd.replyToAuthor("the raffle *\"" + rug.getDescription() + "\"* hasn't finished yet.");
            return;
        } else if (cmd.args.length < 3) {
            cmd.replyToAuthor("you need to provide a description `raffle start \"Awesome raffle!\"`.");
            return;
        }
        rug.setActive(true);
        rug.setCreator(cmd.author.id);
        rug.setDescription(cmd.args[2]);
        rug.clearParticipants();
        cmd.replyToChannel(cmd.author.username + " has started raffle *\"" + rug.getDescription() + "\"*.");
    } else if (action == "join" || action == "j" || action == "enter") {
        if (rug.getActive() == false) {
            cmd.replyToAuthor("there aren't any raffles to join.");
            return;
        } else if (rug.hasParticipant(cmd.author.id) == true) {
            cmd.replyToAuthor("you've already joined this raffle.");
            return;
        }
        rug.addParticipant(cmd.author.id);
        cmd.replyToAuthor("you joined the raffle :tickets:");
    } else if (action == "finish" || action == "end") {
        if (rug.getActive() == false) {
            cmd.replyToAuthor("there aren't any raffles to finish.");
            return;
        } else if (cmd.author.id != rug.getCreator()) {
            cmd.replyToAuthor("you cannot end this raffle.");
            return;
        }

        var msg = "The raffle *\"" + rug.getDescription() + "\"* has ended! ";

        if (rug.participantCount() == 0) {
            msg += "but there weren't any participants so no one wins :face_palm:"
        } else {
            var r = Math.getRandomInRange(0, rug.storage[cmd.channelID]['participants'].length - 1);
            var winnerID = rug.storage[cmd.channelID]['participants'][r];
            msg += "<@" + winnerID + "> won :rotating_light::fireworks::mega:"
        }

        cmd.replyToChannel(msg);
        rug.setActive(false);
    }
});
