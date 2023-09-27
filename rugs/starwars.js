var pong = new DuderRug("Pong", "Returns Ping");

pong.addCommand("pong", function(cmd) {

	cmd.replyToChannel("ping");
});