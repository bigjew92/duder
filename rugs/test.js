var admin = new DuderRug("Admin", "Duder administration tools.");

admin.addCommand( "setuser", function() {
    if (cmd.mentions.length == 0) {
        cmd.replyToChannel("please mention a user");
        return;
    } else if (cmd.args.length < 3) {
        cmd.replyToChannel("please define permission");
        return;
    }

    cmd.replyToChannel("arg is " + cmd.args[2])
    /*
    resp = cmd.mentions[0].addPermission(cmd.channelID, 2);
    if (resp != null) {
        cmd.replyToChannel("Unable to add permission: " + resp);
    } else {
        cmd.replyToChannel("Success!")
    }
    */
});

admin.addCommand( "test", function() {
    cmd.replyToAuthor("args");
    var args = cmd.args;
    var arrayLength = args.length;
    for (var i = 0; i < arrayLength; i++) {
        cmd.replyToChannel( "args " + i + ": " + args[i]);
    }
});
