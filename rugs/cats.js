var cats = new DuderRug("Cats", "Random cat stuff");

cats.addCommand("catfact", function(cmd) {
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

cats.imageTypes = ["jpg", "png"];

cats.addCommand("catpic", function(cmd) {
	var r = Math.getRandomInRange(0, this.imageTypes.length - 1);
	var type = this.imageTypes[r];

	var url = "http://thecatapi.com/api/images/get?format=xml&type={0}".format(type);

	this.showPic(cmd, url);
});

cats.addCommand("catgif", function(cmd) {
	var url = "http://thecatapi.com/api/images/get?format=xml&type=gif";

	this.showPic(cmd, url);
});

cats.showPic = function(cmd, url) {
	Duder.startTyping(cmd.channelID);

	var content = HTTP.get(10, url, {});
	if (content === false) {
		cmd.replyToAuthor("something went wrong");
		return;
	}
	//this.dprint(content);
	var json = XML.toJSON(content);
	if (
		json === undefined ||
		json.response === undefined ||
		json.response.data === undefined ||
		json.response.data.images === undefined ||
		json.response.data.images.image === undefined ||
		json.response.data.images.image.url === undefined
	) {
		cmd.replyToAuthor("could get pic");
		return;
	}
	cmd.replyToChannel(json.response.data.images.image.url);
};
