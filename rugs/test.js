var admin = new Rug("Admin", "Duder administration tools.");

admin.addCommand( "setuser", function() {
    print("hello 1!");
});

admin.addCommand("shutdown", function() {
    shutdown();    
});
