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
	cmd.replyToAuthor("rolled a " + sides + " sided :game_die: and got " + r + ".");
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
	cmd.replyToAuthor("I rub my magic :8ball: balls and the response is `" + this.eightBallResponses[r] + "`.");
});

funbox.addCommand("lebowski", function(cmd) {
	var content;
	var result;
	var json;
	if (cmd.args.length > 1) {
		content = HTTP.get(4, "http://lebowski.me/api/quotes/search?term=" + cmd.args[1]);
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
	var id = content.match('(?s)title="Permanent link to this quote."><b>.+?</b></a>');
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

funbox.auraResponses = [
	"Hello ora and chat This is the governor of Alabama, Al Bama. I regret to inform you that our scientists have gone extinct. One of the Alabama Beach Mice accidentally put Him in the microwave since he thought it was Jimmy Changa. Our bad. We have sent Ora a statue of the beloved scientist to put on his front lawn to commemorate his life. We know Ora loved the scientis as much as the Alabama Beach Mouse. We hope to see the scientist prominently displayed on his front lawn.",
	"To those who think lose = boo, what you of lesser intellect don't understand is Aurateur is such a titan of Mario Maker that he is literally guaranteed to win on every single good level. Since he only likes good levels, he wins only on levels he likes. When he boos, it's because he has deciphered every fiber that composes the level, meticulously figuring out the electron orientations that make it bad. Aura is a god of Mario, and when his boo hammer strikes, it brings justice!",
	"Hi Aurateur- It’s me, your only viewer. For years I have created the illusion that you’re streaming to a large audience. The truth? All these people in the chat are me. And now, to convince you, I will send this message from all my accounts",
	"Idiots. You fucking morons. What are you doing. Spawn here. Toad, stay here. No, what are you doing, you fucking fool. Ugh, you guys are all such oafs. Okay, let’s try this again. Toadette?! What are you doing? Why did you run off? You fucking nincompoop? Ugh, these clowns have no idea what they’re doing. Okay, pick me up Mario and throw me to that ledge. What was that?! You fucking dope. Don’t throw me at the wall. The ground there. Ugh, these imbeciles have no idea what they’re doing",
	"Look at this chat, it's ridiculous, one of the worst things i've ever seen, the chat does not respect anything or anyone, the only thing they can do is copy and paste the same thing over and over again, and not only that some dudes actually pay money to a bald man so they can make this nice girl named Joanna to say some dumb things that represents their stupidity as a human being. Please stop with this, no one likes it, not even the Streamer, or Joanna.",
	"hey aura. so. i got this new anime plot. basically there's this high school girl except she's got huge boobs. i mean some serious honkers. a real set of badonkers. packin some dobonhonkeros. massive dohoonkabhankoloos. big ol' tonhongerekoogers. what happens next?! transfer student shows up with even bigger bonkhonagahoogs. humongous hungolomghononoloughongous. so what do you think aura? pretty good right? hmm?",
	"Licky Ricky bald bald lucky licky Ricky, licky licky bald bald licky ricky bald. Licky Ricky bald bald lucky licky Ricky, licky licky bald bald licky ricky bald. Licky Ricky bald bald lucky licky Ricky, licky licky bald bald licky ricky bald. Licky Ricky bald bald lucky licky Ricky, licky licky bald bald licky ricky bald. Licky Ricky bald bald lucky licky Ricky, licky licky bald bald licky ricky bald. Licky Ricky bald bald lucky licky Ricky, licky licky bald bald licky ricky bald.",
	"So Aurateur, let's have a talk. I strongly feel that your channel would do better if you turned off TTS or at least filtered out the spam. Thing is, while I myself have never donated to you or any other streamer, and my only subs are twitch prime, I feel I know a thing or two about how to succeed as a streamer and I'm certain your channel would do better without all the spam, it's clear nobody wants that shit and is the only thing holding you back as a fulltime streamer. Thank you for listening.",
	"Toad what are you doing. Mario, die. Just die. All of you die. Now, now, spawn on me. Ok, here's what we're gonna do. Mario.. wait a second... Actually, toad, YOU pick up Mario. I'm going to jump off this edge, and you take Mario and jump off a little bit after, bounce off the top of my head, and then just throw him....... OMG, Toad, where were you? You were supposed to... Ok, were going to try this again... Now, listen close as I explain it to you one more time.",
	"The Alabama beach mouse (Peromyscus polionotus ammobates) is a federally endangered species which lives along the Alabama coast. The range of the Alabama beach mouse historically included much of the Fort Morgan Peninsula on the Alabama Gulf coast and extends from Ono Island to Fort Morgan. Coastal residential and commercial development and roadway construction have fragmented and destroyed habitat used by this species. Hurricanes, tropical storms.",
];

funbox.addCommand("aura", function(cmd) {
	var r = Math.getRandomInRange(0, this.auraResponses.length - 1);
	cmd.replyToChannel("`" + this.auraResponses[r] + "`.");
});
