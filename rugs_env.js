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

// Define DuderPermission class
function DuderPermission() {}
DuderPermission.permissions = __BIND__;
DuderPermission.getName = function(value) {
	value = value.toString();
	for (var name in DuderPermission.permissions) {
		if (DuderPermission.permissions[name] == value) {
			return name;
		}
	}

	return "invalid";
};
DuderPermission.getNames = function(values) {
	var names = "";
	for (var value in values) {
		if (names.length > 0) {
			names += ", ";
		}
		names += DuderPermission.getName(values[value]);
	}

	return names.length > 0 ? names : "none";
};

// Define DuderUser class
function DuderUser(channelID, id, username) {
	this.id = id;
	this.username = username;
	this.isOwner = __BIND__(channelID, id);
	this.isModerator = __BIND__(channelID, id);
}
DuderUser.prototype.getPermissions = function(channelID) {
	return __BIND__(channelID, this.id);
};
DuderUser.prototype.setPermissions = function(channelID, permission, add) {
	// ensure add is boolean
	add = add == true;
	return __BIND__(channelID, this.id, permission, add);
};
DuderUser.getUsernameByID = function(channelID, userID) {
	return __BIND__(channelID, userID);
};

// Define DuderCommand class
function DuderCommand() {
	this.mentions = [];
}
DuderCommand.prototype.replyToChannel = function(content) {
	__BIND__(this.channelID, content);
};
DuderCommand.prototype.replyToChannelEmbed = function(content) {
	__BIND__(this.channelID, content);
};
DuderCommand.prototype.replyToAuthor = function(content, mention) {
	// ensure 'mention' is bool and default is false
	mention = mention == true;
	__BIND__(
		this.channelID,
		this.author.id,
		this.author.username,
		content,
		mention
	);
};
DuderCommand.prototype.isMention = function(str) {
	return str.substring(0, 2) == "<@" && str.substring(str.length - 1) == ">";
};
DuderCommand.prototype.deleteMessage = function() {
	__BIND__(this.channelID, this.messageID);
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
		return typeof args[number] != "undefined" ? args[number] : match;
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
		if (this[i] == elem) return true;
	}
	return false;
};

// HTTP
function HTTP() {}
HTTP.get = function(timeout, url, as_string) {
	// ensure 'as_string' is bool and default is true
	as_string = as_string !== false;
	return __BIND__(timeout, url, as_string);
};
HTTP.post = function(timeout, url, data) {
	return __BIND__(timeout, url, data);
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
