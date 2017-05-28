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

var eightBallResponses = new Array(
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
    r = Math.getRandomInRange(0, eightBallResponses.length - 1);
    cmd.replyToAuthor("I rub my magic :8ball: balls and the response is `" + eightBallResponses[r] + "`");
});

funbox.addCommand("fact", function() {
    /*
    var text = Web.get("http://catfacts-api.appspot.com/api/facts");
    var obj = JSON.parse(text);
    print(obj.facts[0]);
    */
    /*
    var text = Web.get("http://numbersapi.com/random");
    if (text) {
        cmd.replyToAuthor(text);
    }
    */
});
