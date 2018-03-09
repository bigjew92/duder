// Define Duder class
function Duder() {}
Duder.setStatus = function(status) {
	return __BIND__(status);
};
Duder.setAvatar = function(avatar) {
	return __BIND__(avatar);
};
Duder.saveAvatar = function(filename) {
	return __BIND__(filename);
};
Duder.getAvatars = function() {
	return __BIND__();
};
Duder.useAvatar = function(filename) {
	return __BIND__(filename);
};
Duder.startTyping = function(channelID) {
	return __BIND__(channelID);
};

// Define DuderUser class
function DuderUser(guildID, id, username) {
	this.id = id;
	this.username = username;
	this.isOwner = __BIND__(id);
	this.isManager = __BIND__(guildID, id);
	this.isModerator = __BIND__(guildID, id);
}
DuderUser.prototype.setNickname = function(guildID, nickname) {
	return __BIND__(guildID, this.id, nickname);
};
DuderUser.getUsernameByID = function(guildID, userID) {
	return __BIND__(guildID, userID);
};

// Define DuderCommand class
function DuderCommand(guildID, channelID, messageID, author, mentions, args) {
	this.guildID = guildID;
	this.channelID = channelID;
	this.messageID = messageID;
	this.author = author;
	this.mentions = mentions;
	this.args = args;
}
DuderCommand.prototype.replyToChannel = function(content) {
	__BIND__(this.channelID, content);
};
DuderCommand.prototype.replyToChannelEmbed = function(content) {
	__BIND__(this.channelID, content);
};
DuderCommand.prototype.replyToAuthor = function(content, mention) {
	// ensure 'mention' is bool and default is false
	mention = mention === true;
	__BIND__(
		this.channelID,
		this.author.id,
		this.author.username,
		content,
		mention
	);
};
DuderCommand.prototype.isMention = function(str) {
	return str.substring(0, 2) === "<@" && str.substring(str.length - 1) === ">";
};
DuderCommand.prototype.deleteMessage = function() {
	__BIND__(this.channelID, this.messageID);
};
DuderCommand.prototype.sendFile = function(channelID, name, data) {
	__BIND__(channelID, name, data);
};

// Define DuderRug class
function DuderRug(name, description) {
	__BIND__(this, name, description);
}
DuderRug.prototype.addCommand = function(trigger, exec) {
	__BIND__(this, trigger, exec);
};
DuderRug.prototype.loadStorage = function() {
	var data = __BIND__(this);
	return JSON.parse(data);
};
DuderRug.prototype.saveStorage = function(json) {
	var data = JSON.stringify(json, null, "\t");
	return __BIND__(this, data);
};
DuderRug.prototype.dprint = function(msg) {
	__BIND__(this, msg);
};
DuderRug.prototype.wprint = function(msg) {
	__BIND__(this, msg);
};

// Math
Math.getRandomInRange = function(min, max) {
	min = Math.ceil(min);
	max = Math.floor(max);
	return Math.floor(Math.random() * (max - min + 1)) + min;
};
Math.clamp = function(val, min, max) {
	return Math.max(min, Math.min(val, max));
};

// String
String.prototype.decodeHTML = function() {
	return __BIND__(this);
};
String.prototype.replaceAll = function(search, replacement) {
	var target = this;
	return target.replace(new RegExp(search, "g"), replacement);
};
String.prototype.toBinary = function() {
	var result = [];
	for (var i = 0; i < this.length; i++) {
		result.push(this.charCodeAt(i));
	}
	return result;
};
String.prototype.format = function() {
	var args = arguments;
	return this.replace(/{(\d+)}/g, function(match, number) {
		return typeof args[number] !== "undefined" ? args[number] : match;
	});
};
String.prototype.matchAll = function(regexp) {
	var matches = [];
	this.replace(regexp, function() {
		var arr = [].slice.call(arguments, 0);
		var extras = arr.splice(-2);
		arr.index = extras[0];
		arr.input = extras[1];
		matches.push(arr);
	});
	return matches.length ? matches : null;
};

// Array
Array.prototype.contains = function(elem) {
	for (var i in this) {
		if (this[i] === elem) return true;
	}
	return false;
};

// HTTP
function HTTP() {}
HTTP.get = function(timeout, url, headers, as_string) {
	// ensure 'as_string' is bool and default is true
	as_string = as_string !== false;

	// ensure 'headers' is an array
	if (headers === null || headers === undefined) {
		headers = {};
	} else if (!(headers instanceof Object)) {
		headers = {};
	}

	return __BIND__(timeout, url, headers, as_string);
};
HTTP.post = function(timeout, url, values) {
	return __BIND__(timeout, url, values);
};
HTTP.detectContentType = function(content) {
	return __BIND__(content);
};
HTTP.parseURL = function(str) {
	return __BIND__(str);
};

// Base64
function Base64() {}
Base64.encodeToString = function(bytes) {
	return __BIND__(bytes);
};

// Misc
function isNumeric(n) {
	return !isNaN(parseFloat(n)) && isFinite(n);
}

// EmbedMessage
function EmbedMessage() {
	this.data = {
		title: null,
		description: null,
		url: null,
		color: null,
		timestamp: null,
		footer: null,
		thumbnail: null,
		image: null,
		author: null,
		fields: null
	};
}
EmbedMessage.prototype.setTitle = function(title) {
	this.data.title = title;
};
EmbedMessage.prototype.setDescription = function(description) {
	this.data.description = description;
};
EmbedMessage.prototype.setURL = function(url) {
	this.data.url = url;
};
EmbedMessage.prototype.setColor = function(color) {
	this.data.color = color;
};
EmbedMessage.prototype.setTimestamp = function() {
	var dt = new Date();
	this.data.timestamp = dt.toISOString();
};
EmbedMessage.prototype.setFooter = function(icon, text) {
	this.data.footer = '{\n\t\t"icon_url": "{0}",\n\t\t"text": "{1}"\n\t}'.format(
		icon,
		text
	);
};
EmbedMessage.prototype.setThumbnail = function(thumbnail) {
	this.data.thumbnail = '{\n\t\t"url": "{0}"\n\t}'.format(thumbnail);
};
EmbedMessage.prototype.setImage = function(image) {
	this.data.image = '{\n\t\t"url": "{0}"\n\t}'.format(image);
};
EmbedMessage.prototype.setAuthor = function(name, url, icon) {
	this.data.author = '{\n\t\t"name": "{0}",\n\t\t"url": "{1}",\n\t\t"icon_url": "{2}"\n\t}'.format(
		name,
		url,
		icon
	);
};
EmbedMessage.prototype.addField = function(name, value) {
	if (this.data.fields === null) {
		this.data.fields = [];
	}
	var field = [name, value];
	this.data.fields.push(field);
};
EmbedMessage.prototype.compile = function() {
	var content = "";
	var COMMA = function(c) {
		return c.length > 0 ? ",\n" : "";
	};
	if (this.data.title !== null) {
		content += COMMA(content) + '\t"title": "{0}"'.format(this.data.title);
	}
	if (this.data.description !== null) {
		content +=
			COMMA(content) +
			'\t"description": "{0}"'.format(this.data.description);
	}
	if (this.data.url !== null) {
		content += COMMA(content) + '\t"url": "{0}"'.format(this.data.url);
	}
	if (this.data.color !== null) {
		content += COMMA(content) + '\t"color": {0}'.format(this.data.color);
	}
	if (this.data.timestamp !== null) {
		content +=
			COMMA(content) + '\t"timestamp": "{0}"'.format(this.data.timestamp);
	}
	if (this.data.footer !== null) {
		content += COMMA(content) + '\t"footer": {0}'.format(this.data.footer);
	}
	if (this.data.thumbnail !== null) {
		content +=
			COMMA(content) + '\t"thumbnail": {0}'.format(this.data.thumbnail);
	}
	if (this.data.image !== null) {
		content += COMMA(content) + '\t"image": {0}'.format(this.data.image);
	}
	if (this.data.author !== null) {
		content += COMMA(content) + '\t"author": {0}'.format(this.data.author);
	}
	if (this.data.fields !== null) {
		var fields = '\t"fields": [\n';
		for (var i = 0; i < this.data.fields.length; i++) {
			var field = this.data.fields[i];
			fields += '\t\t{\n\t\t\t"name": "{0}",\n\t\t\t"value": "{1}"\n\t\t}{2}\n'.format(
				field[0],
				field[1],
				i < this.data.fields.length - 1 ? "," : ""
			);
		}
		fields += "\t]";
		content += COMMA(content) + fields;
	}

	content = "{\n" + content + "\n}";

	return content;
};
