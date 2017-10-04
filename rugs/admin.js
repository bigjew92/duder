var admin = new DuderRug("Admin", "Duder administration tools.");

admin.addCommand("setuser", function() {
    if (!cmd.author.isOwner) {
        cmd.replyToAuthor("you are not authorized.");
        return;
    } else if ((cmd.mentions.length != 1) || (cmd.args.length < 3)) {
        cmd.replyToAuthor("usage: `setuser @mention (+/-)permission`");
        return;
    }

    var modifier = cmd.args[2].substring(0, 1);

    if (modifier != "+" && modifier != "-") {
        cmd.replyToAuthor("the permission must start with `+` or `-` to add or remove.");
        return;
    }

    var perm = cmd.args[2].substring(1);

    resp = cmd.mentions[0].modifyPermission(cmd.channelID, perm, modifier == "+");
    if (resp != null) {
        cmd.replyToAuthor("unable to add permission: *" + resp + "*");
    } else {
        perms = cmd.mentions[0].getPermissions(cmd.channelID);
        if (perms.length == 0) {
            cmd.replyToAuthor(cmd.mentions[0].username + " doesn't have any permissions.");
        } else {
            cmd.replyToAuthor(cmd.mentions[0].username + " now has permission(s) " + DuderPermission.getNames(perms) + ".");
        }
    }
});

admin.addCommand("viewuser", function() {
    if (!cmd.author.isOwner) {
        cmd.replyToAuthor("you are not authorized.");
        return;
    } else if ((cmd.mentions.length != 1) || (cmd.args.length != 2)) {
        cmd.replyToAuthor("usage: `viewuser @mention`")
        return;
    }

    var perms = cmd.mentions[0].getPermissions(cmd.channelID);
    if (perms.length == 0) {
        cmd.replyToAuthor(cmd.mentions[0].username + " doesn't have any permissions.");
    } else {
        cmd.replyToAuthor(cmd.mentions[0].username + " has permission(s) " + DuderPermission.getNames(perms) + ".");
    }
});

admin.addCommand("viewself", function() {
    if (cmd.author.isOwner) {
        cmd.replyToAuthor("you are the owner.");
        return;
    }

    var perms = cmd.author.getPermissions(cmd.channelID);
    if (perms.length == 0) {
        cmd.replyToAuthor("you don't have any permissions.");
    } else {
        cmd.replyToAuthor("you have permission(s) " + DuderPermission.getNames(perms) + ".");
    }
});

admin.addCommand("test", function() {
});
