var test = new Rug("Test Rug", "The best rug");

test.dostuff = function() {
    this.whatever = 321;
}

test.someshit = function() {
    print("hello 2! " + rug.name);
}

test.whatever = 123;
test.addCommand( "testcmd1", function() {
        print("hello 1!");
    }
);
test.addCommand( "testcmd2", test.someshit );
