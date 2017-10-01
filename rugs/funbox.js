var funbox = new DuderRug("Funbox", "Some silly things to play with.");

funbox.addCommand("dice", function() {
    sides = 6;
    if (cmd.args.length > 1) {
        s = parseInt(cmd.args[1]);
        if (!isNaN(s)) {
            sides = Math.clamp(sides, 2, 99);
        }
    }
    r = Math.getRandomInRange(2, sides);
    cmd.replyToAuthor("rolled a " + sides + " sided dice and got " + r);
});

funbox.eightBallResponses = new Array(
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
);

funbox.addCommand("8ball", function() {
    r = Math.getRandomInRange(0, rug.eightBallResponses.length - 1);
    cmd.replyToAuthor("I rub my magic :8ball: balls and the response is `" + rug.eightBallResponses[r] + "`");
});

funbox.lebowskiQuoteCallback = function( text ) {
    json = JSON.decode(text);
    quote = "```";
    for(var k in json['quote']['lines']) {
        line = json['quote']['lines'][k];
        quote += line['character']['name'] + ": " + line['text'] + "\n";
    }
    quote += "```";
    cmd.replyToChannel(quote);
}

funbox.bashQuoteCallback = function( text ) {
    print("bash!");
}

funbox.quoteSources = {
    "http://lebowski.me/api/quotes/random": funbox.lebowskiQuoteCallback,
    "http://bash.org/?random1": funbox.bashQuoteCallback
};

funbox.addCommand("quote", function() {
    var keys = Object.keys(rug.quoteSources);
    r = Math.getRandomInRange(0, keys.length - 1);
    var key = keys[r];
    key = "http://lebowski.me/api/quotes/random";
    var text = HTTP.get(4, key);
    rug.quoteSources[key]( text );
});
