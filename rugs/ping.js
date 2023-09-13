var ping = new DuderRug("Ping", "Returns Pong");

ping.addCommand("ping", function(cmd) {

	cmd.replyToChannel("pong");
});