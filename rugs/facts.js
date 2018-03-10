var facts = new DuderRug("Facts", "Random facts");

facts.addCommand("catfact", function(cmd) {
	var url = "https://catfact.ninja/fact";

	Duder.startTyping(cmd.channelID);

	var content = HTTP.get(10, url, {});
	if (content === false) {
		cmd.replyToAuthor("something went wrong");
		return;
	}
	var json = JSON.parse(content);
	if (json === undefined || json.fact === undefined) {
		cmd.replyToAuthor("unable to get request json");
		return;
	}

	cmd.replyToChannel("```{0}```".format(json.fact));
});
