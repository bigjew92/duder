var memes = new DuderRug("Memes", "Make memes");

memes.templates = {
	"twbg": "563423",
	"nomb": "16464531"
};

memes.addCommand("meme", function(cmd) {
	if (cmd.args.length < 4) {
		cmd.replyToAuthor('usage: `meme template "top text" "bottom text"`.');
		return;
	}
	var template = cmd.args[1];

	if (this.templates[template] === undefined) {
		cmd.replyToAuthor("invalid template");
		return;
	}

	var top = cmd.args[2];
	var bottom = cmd.args[3];

	var url = "https://api.imgflip.com/caption_image";

	var values = {};
	values.username = "foszor";
	values.password = "@imfl0PPer";
	values.template_id = this.templates[template];
	values.text0 = top;
	values.text1 = bottom;

	var content = HTTP.post(10, url, values);
	this.dprint(content);
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
