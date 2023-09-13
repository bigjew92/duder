var admin = new DuderRug("Admin", "Duder administration tools.");

admin.addCommand("status", function(cmd) {
	if (!cmd.author.isOwner) {
		cmd.replyToAuthor("you are not authorized.");
		return;
	} else if (cmd.args.length < 2) {
		cmd.replyToAuthor('usage: `status "New Status"`');
		return;
	}
	var resp = Duder.setStatus(cmd.args[1]);
	if (resp !== true) {
		cmd.replyToAuthor("unable to update status: `" + resp + "`");
	}
});

admin.avatarContentTypes = ["image/png", "image/jpeg"];

admin.addCommand("avatar", function(cmd) {
	if (!cmd.author.isOwner) {
		cmd.replyToAuthor("you are not authorized.");
		return;
	} else if (cmd.args.length < 2) {
		cmd.replyToAuthor("usage: `avatar <action>`.");
		return;
	}
	var action = cmd.args[1];
	if (action === "url") {
		if (cmd.args.length < 3) {
			cmd.replyToAuthor("usage: `avatar url http://link.to/image.png`.");
			return;
		}
		Duder.startTyping(cmd.channelID);

		var url = HTTP.parseURL(cmd.args[2]);
		if (url === false) {
			cmd.replyToAuthor("invalid link.");
			return;
		}
		var bytes = HTTP.get(4, url, {}, false);
		if (bytes === false) {
			cmd.replyToAuthor("unable to download the link.");
			return;
		}

		var contentType = HTTP.detectContentType(bytes);
		if (contentType === false) {
			cmd.replyToAuthor("unable to detect content type.");
			return;
		}

		if (!this.avatarContentTypes.contains(contentType)) {
			cmd.replyToAuthor("invalid image type.");
			return;
		}

		var base64 = base64.encodeToString(bytes);
		if (base64 === false) {
			cmd.replyToAuthor("unable to encode image.");
			return;
		}
		var avatar = "data:{0};base64,{1}".format(contentType, base64);
		Duder.setAvatar(avatar);
		cmd.replyToChannel(":ok_hand:");
	} else if (action === "save") {
		if (cmd.args.length < 3) {
			cmd.replyToAuthor("usage: `avatar save filename.png`.");
			return;
		}
		Duder.startTyping(cmd.channelID);

		var resp = Duder.saveAvatar(cmd.args[2]);
		if (resp !== true) {
			cmd.replyToAuthor("unable to save avatar: " + resp);
			return;
		}
		cmd.replyToChannel(":ok_hand:");
	} else if (action === "list") {
		var avatars = Duder.getAvatars();
		if (avatars === false) {
			cmd.replyToAuthor("unable to list avatars.");
			return;
		} else if (avatars.length == 0) {
			cmd.replyToAuthor("no save avatars.");
			return;
		}
		var msg = "```";
		for (var i = 0; i < avatars.length; i++) {
			if (i > 0) {
				msg += "\n";
			}
			msg += avatars[i];
		}
		msg += "```";
		cmd.replyToChannel(msg);
	} else if (action === "use") {
		if (cmd.args.length < 3) {
			cmd.replyToAuthor("usage: `avatar use filename.png`.");
			return;
		}
		Duder.startTyping(cmd.channelID);

		var r = Duder.useAvatar(cmd.args[2]);
		if (r !== true) {
			cmd.replyToAuthor("unable to use avatar: " + r);
			return;
		}
		cmd.replyToChannel(":ok_hand:");
	}
});
