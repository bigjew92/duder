var funbox = new DuderRug("Funbox", "Some silly things to play with.");

funbox.addCommand("dice", function(cmd) {
	sides = 6;
	if (cmd.args.length > 1) {
		s = parseInt(cmd.args[1]);
		if (!isNaN(s)) {
			sides = Math.clamp(sides, 2, 99);
		}
	}
	r = Math.getRandomInRange(2, sides);
	cmd.replyToAuthor(
		"rolled a " + sides + " sided :game_die: and got " + r + "."
	);
});

funbox.eightBallResponses = [
	"It is certain",
	"It is decidedly so",
	"Without a doubt",
	"Yes, definitely",
	"You may rely on it",
	"As I see it, yes",
	"Most likely",
	"Outlook good",
	"Yes",
	"Signs point to yes",
	"Reply hazy try again",
	"Ask again later",
	"Better not tell you now",
	"Cannot predict now",
	"Concentrate and ask again",
	"Don't count on it",
	"My reply is no",
	"My sources say no",
	"Outlook not so good",
	"Very doubtful"
];

funbox.addCommand("8ball", function(cmd) {
	var r = Math.getRandomInRange(0, this.eightBallResponses.length - 1);
	cmd.replyToAuthor(
		"I rub my magic :8ball: balls and the response is `" +
			this.eightBallResponses[r] +
			"`."
	);
});

funbox.addCommand("lebowski", function(cmd) {
	var content;
	var result;
	var json;
	if (cmd.args.length > 1) {
		content = HTTP.get(
			4,
			"http://lebowski.me/api/quotes/search?term=" + cmd.args[1]
		);
		json = JSON.parse(content);
		if (json.results.length === 0) {
			cmd.replyToChannel("¯\\_(ツ)_/¯");
			return;
		}
		result = json.results[0];
	} else {
		content = HTTP.get(4, "http://lebowski.me/api/quotes/random");
		json = JSON.parse(content);
		result = json.quote;
	}

	var quote = "```";
	for (var k in result.lines) {
		if (isNumeric(k)) {
			var line = result.lines[k];
			quote += line.character.name + ": " + line.text + "\n";
		}
	}
	quote += "```";
	cmd.replyToChannel(quote);
});

funbox.addCommand("bash", function(cmd) {
	var content = HTTP.get(4, "http://bash.org/?random1");
	var id = content.match(
		'(?s)title="Permanent link to this quote."><b>.+?</b></a>'
	);
	id = id[0].substring(42);
	id = id.substring(0, id.length - 8);
	var link = "<http://bash.org/?quote=" + id + ">";

	var quote = content.match('(?s)<p class="qt">.+?</p>');
	quote = quote[0].substring(14);
	quote = quote.substring(0, quote.length - 4);
	quote = unescape(quote);
	quote = quote.decodeHTML();
	quote = quote.replace(new RegExp("<br />", "g"), "");

	cmd.replyToChannel(link + "\n```" + quote + "```");
});

funbox.bigText = {
	" ": " ",
	0: ":zero:",
	1: ":one:",
	2: ":two:",
	3: ":three:",
	4: ":four:",
	5: ":five:",
	6: ":six:",
	7: ":seven:",
	8: ":eight:",
	9: ":nine:",
	"!": ":exclamation:",
	"?": ":question:",
	a: ":regional_indicator_a:",
	b: ":regional_indicator_b:",
	c: ":regional_indicator_c:",
	d: ":regional_indicator_d:",
	e: ":regional_indicator_e:",
	f: ":regional_indicator_f:",
	g: ":regional_indicator_g:",
	h: ":regional_indicator_h:",
	i: ":regional_indicator_i:",
	j: ":regional_indicator_j:",
	k: ":regional_indicator_k:",
	l: ":regional_indicator_l:",
	m: ":regional_indicator_m:",
	n: ":regional_indicator_n:",
	o: ":regional_indicator_o:",
	p: ":regional_indicator_p:",
	q: ":regional_indicator_q:",
	r: ":regional_indicator_r:",
	s: ":regional_indicator_s:",
	t: ":regional_indicator_t:",
	u: ":regional_indicator_u:",
	v: ":regional_indicator_v:",
	w: ":regional_indicator_w:",
	x: ":regional_indicator_x:",
	y: ":regional_indicator_y:",
	z: ":regional_indicator_z:"
};

// big
funbox.addCommand("big", function(cmd) {
	if (cmd.args.length < 2) {
		return;
	}

	cmd.args.splice(0, 1);
	print(cmd.args);
	var original = cmd.args.join("");
	var bigged = "";
	for (var i = 0; i < original.length; i++) {
		var char = original[i].toLowerCase();
		if (this.bigText[char] !== undefined) {
			bigged += this.bigText[char];
		}
	}

	cmd.replyToChannel(bigged);
	cmd.deleteMessage();
});

funbox.smallText = {
	" ": " ",
	"0": "⁰",
	"1": "¹",
	"2": "²",
	"3": "³",
	"4": "⁴",
	"5": "⁵",
	"6": "⁶",
	"7": "⁷",
	"8": "⁸",
	"9": "⁹",
	a: "ᵃ",
	b: "ᵇ",
	c: "ᶜ",
	d: "ᵈ",
	e: "ᵉ",
	f: "ᶠ",
	g: "ᵍ",
	h: "ʰ",
	i: "ᶦ",
	j: "ʲ",
	k: "ᵏ",
	l: "ˡ",
	m: "ᵐ",
	n: "ⁿ",
	o: "ᵒ",
	p: "ᵖ",
	q: "ᑫ",
	r: "ʳ",
	s: "ˢ",
	t: "ᵗ",
	u: "ᵘ",
	v: "ᵛ",
	w: "ʷ",
	x: "ˣ",
	y: "ʸ",
	z: "ᶻ",
	A: "ᴬ",
	B: "ᴮ",
	C: "ᶜ",
	D: "ᴰ",
	E: "ᴱ",
	F: "ᶠ",
	G: "ᴳ",
	H: "ᴴ",
	I: "ᴵ",
	J: "ᴶ",
	K: "ᴷ",
	L: "ᴸ",
	M: "ᴹ",
	N: "ᴺ",
	O: "ᴼ",
	P: "ᴾ",
	Q: "Q",
	R: "ᴿ",
	S: "ˢ",
	T: "ᵀ",
	U: "ᵁ",
	V: "ⱽ",
	W: "ᵂ",
	X: "ˣ",
	Y: "ʸ",
	Z: "ᶻ",
	"+": "⁺",
	"-": "⁻",
	"=": "⁼",
	"(": "⁽",
	")": "⁾"
};

// small
funbox.addCommand("smol", function(cmd) {
	if (cmd.args.length < 2) {
		return;
	}

	cmd.args.splice(0, 1);
	var original = cmd.args.join("");
	var smalled = "";
	for (var i = 0; i < original.length; i++) {
		var char = original[i].toLowerCase();
		if (this.smallText[char] !== undefined) {
			smalled += this.smallText[char];
		}
	}

	cmd.replyToChannel(smalled);
	cmd.deleteMessage();
});
