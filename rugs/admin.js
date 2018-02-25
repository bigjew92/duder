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

admin.addCommand("viewuser", function(cmd) {
	if (!cmd.author.isOwner) {
		cmd.replyToAuthor("you are not authorized.");
		return;
	} else if (cmd.mentions.length !== 1 || cmd.args.length !== 2) {
		cmd.replyToAuthor("usage: `viewuser @mention`");
		return;
	}

	var perms = cmd.mentions[0].getPermissions(cmd.guildID);
	if (perms.length === 0) {
		cmd.replyToAuthor(
			cmd.mentions[0].username + " doesn't have any permissions."
		);
	} else {
		cmd.replyToAuthor(
			cmd.mentions[0].username +
				" has permission(s) " +
				DuderPermission.getNames(perms) +
				"."
		);
	}
});

admin.addCommand("setuser", function(cmd) {
	if (!cmd.author.isOwner) {
		cmd.replyToAuthor("you are not authorized.");
		return;
	} else if (cmd.mentions.length !== 1 || cmd.args.length < 3) {
		cmd.replyToAuthor("usage: `setuser @mention (+/-)permission`");
		return;
	}

	var modifier = cmd.args[2].substring(0, 1);

	if (modifier !== "+" && modifier !== "-") {
		cmd.replyToAuthor(
			"the permission must start with `+` or `-` to add or remove."
		);
		return;
	}

	var perm = cmd.args[2].substring(1);

	var resp = cmd.mentions[0].setPermissions(
		cmd.guildID,
		perm,
		modifier === "+"
	);
	if (resp !== null) {
		cmd.replyToAuthor("unable to add permission: *" + resp + "*");
	} else {
		var perms = cmd.mentions[0].getPermissions(cmd.guildID);
		if (perms.length === 0) {
			cmd.replyToAuthor(
				cmd.mentions[0].username + " doesn't have any permissions."
			);
		} else {
			cmd.replyToAuthor(
				cmd.mentions[0].username +
					" now has permission(s) " +
					DuderPermission.getNames(perms) +
					"."
			);
		}
	}
});

admin.addCommand("viewself", function(cmd) {
	if (cmd.author.isOwner) {
		cmd.replyToAuthor("you are the owner.");
		return;
	}

	var perms = cmd.author.getPermissions(cmd.guildID);
	if (perms.length === 0) {
		cmd.replyToAuthor("you don't have any permissions.");
	} else {
		cmd.replyToAuthor(
			"you have permission(s) " + DuderPermission.getNames(perms) + "."
		);
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

		var base64 = Base64.encodeToString(bytes);
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
