var admin = new DuderRug("Admin", "Duder administration tools.");

admin.addCommand( "setuser", function() {
    if (!cmd.author.isOwner) {
        cmd.replyToChannel("Nope.");
        return;
    } else if ((cmd.mentions.length != 1) || (cmd.args.length < 3)) {
        cmd.replyToChannel("Usage: `setuser @mention (+/-)permission`")
        return;
    }

    var modifier = cmd.args[2].substring(0,1);

    if (modifier != "+" && modifier != "-") {
        cmd.replyToChannel("The permission must start with + or - to add or remove");
        return;
    }

    var perm = cmd.args[2].substring(1);

    resp = cmd.mentions[0].modifyPermission(cmd.channelID, perm, modifier == "+");
    if (resp != null) {
        cmd.replyToChannel("Unable to add permission: " + resp);
    } else {
        perms = cmd.mentions[0].getPermissions(cmd.channelID);
        if (perms.length == 0) {
            cmd.replyToChannel(cmd.mentions[0].username + " doesn't have any permissions");
        } else {
            cmd.replyToChannel(cmd.mentions[0].username + " now has permission(s) " + DuderPermission.getNames(perms));
        }
    }
});

admin.addCommand( "viewuser", function() {
    if (!cmd.author.isOwner) {
        cmd.replyToChannel("Nope.");
        return;
    } else if ((cmd.mentions.length != 1) || (cmd.args.length != 2)) {
        cmd.replyToChannel("Usage: `viewuser @mention`")
        return;
    }

    perms = cmd.mentions[0].getPermissions(cmd.channelID);
    if (perms.length == 0) {
        cmd.replyToChannel(cmd.mentions[0].username + " doesn't have any permissions");
    } else {
        cmd.replyToChannel(cmd.mentions[0].username + " has permission(s) " + DuderPermission.getNames(perms));
    }
});

admin.addCommand( "test", function() {
    if (!cmd.author.isOwner) {
        cmd.replyToChannel("Nope.");
        return;
    }

    cmd.replyToChannel(cmd.channelID);
});
