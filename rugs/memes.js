var memes = new DuderRug("Memes", "Make memes");
memes.storage = memes.loadStorage();

memes.getAPIUsername = function() {
	if (this.storage.settings === undefined) {
		return false;
	} else if (this.storage.settings.username === undefined) {
		return false;
	}
	return this.storage.settings.username;
};

memes.setAPIUsername = function(username) {
	if (this.storage.settings === undefined) {
		this.storage.settings = {};
	}
	this.storage.settings.username = username;
	this.saveStorage(this.storage);
};

memes.getAPIPassword = function() {
	if (this.storage.settings === undefined) {
		return false;
	} else if (this.storage.settings.password === undefined) {
		return false;
	}
	return this.storage.settings.password;
};

memes.setAPIPassword = function(password) {
	if (this.storage.settings === undefined) {
		this.storage.settings = {};
	}
	this.storage.settings.password = password;
	this.saveStorage(this.storage);
};

memes.getTemplateID = function(templateName) {
	if (this.storage.templates === undefined) {
		return false;
	} else if (this.storage.templates[templateName] === undefined) {
		return false;
	}

	return this.storage.templates[templateName];
};

memes.addTemplate = function(templateID, templateName) {
	if (this.storage.templates === undefined) {
		this.storage.templates = {};
	}
	this.storage.templates[templateName] = templateID;
	this.saveStorage(this.storage);
};

memes.addCommand("meme", function(cmd) {
	var action = cmd.args.length > 1 ? cmd.args[1].toLowerCase() : undefined;

	if (action === "setusername") {
		if (cmd.args.length === 3) {
			this.setAPIUsername(cmd.args[2]);
			cmd.replyToAuthor("Imgflip API username has been saved.");
			return;
		} else {
			cmd.replyToAuthor("usage: `setusername YOUR_API_USERNAME`");
			return;
		}
	} else if (action === "setpassword") {
		if (cmd.args.length === 3) {
			this.setAPIPassword(cmd.args[2]);
			cmd.replyToAuthor("Imgflip API password has been saved.");
			return;
		} else {
			cmd.replyToAuthor("usage: `setpassword YOUR_API_PASSWORD`");
			return;
		}
	} else if (action === "addtemplate") {
		if (!cmd.author.isModerator) {
			cmd.replyToAuthor("you don't have permissions to do that");
			return;
		}
		if (cmd.args.length === 4) {
			var id = cmd.args[2];
			var name = cmd.args[3];
			if (!isNumeric(id)) {
				cmd.replyToAuthor("templateID should be numeric");
				return;
			}
			this.addTemplate(id, name);
			cmd.replyToAuthor("templated added");
			return;
		} else {
			cmd.replyToAuthor("usage: `addtemplate templateID templateName`");
			return;
		}
	}

	if (this.getAPIUsername() === false || this.getAPIPassword() === false) {
		cmd.replyToAuthor("Imgflip API username and password are required.");
		return;
	}

	if (cmd.args.length < 4) {
		cmd.replyToAuthor('usage: `meme template "top text" "bottom text"`.');
		return;
	}
	var templateName = cmd.args[1];
	var templateID = this.getTemplateID(templateName);

	if (templateID === false) {
		cmd.replyToAuthor("invalid template");
		return;
	}

	var top = cmd.args[2];
	var bottom = cmd.args[3];

	var url = "https://api.imgflip.com/caption_image";

	var values = {};
	values.username = this.getAPIUsername();
	values.password = this.getAPIPassword();
	values.template_id = templateID;
	values.text0 = top;
	values.text1 = bottom;

	Duder.startTyping(cmd.channelID);

	var content = HTTP.post(10, url, values);
	//this.dprint(content);
	if (content === false) {
		cmd.replyToAuthor("something went wrong");
		return;
	}
	var json = JSON.parse(content);
	if (json.success === undefined) {
		cmd.replyToAuthor("unable to get request json");
		return;
	} else if (json.success === false) {
		cmd.replyToAuthor("error {0}".format(json.error_message));
		return;
	}

	var img = json.data.url;
	cmd.replyToChannel(img);
});
