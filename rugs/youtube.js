var youtube = new DuderRug("YouTube", "Do stuff with YouTube.");
youtube.storage = youtube.loadStorage();

youtube.getAPIKey = function() {
	if (this.storage.settings == undefined) {
		return false;
	} else if (this.storage.settings.api_key == undefined) {
		return false;
	}
	return this.storage.settings.api_key;
};

youtube.setAPIKey = function(key) {
	if (this.storage.settings == undefined) {
		this.storage.settings = {};
	}
	this.storage.settings.api_key = key;
	rug.saveStorage(this.storage);
};

youtube.getPlaylistID = function() {
	if (this.storage[cmd.channelID] == undefined) {
		return false;
	} else if (this.storage[cmd.channelID].playlist == undefined) {
		return false;
	}
	return this.storage[cmd.channelID].playlist;
};

youtube.setPlaylistID = function(id) {
	if (this.storage[cmd.channelID] == undefined) {
		this.storage[cmd.channelID] = {};
	}
	this.storage[cmd.channelID].playlist = id;
	rug.saveStorage(this.storage);
};

youtube.getVideosCache = function() {
	if (this.storage[cmd.channelID] == undefined) {
		return false;
	} else if (this.storage[cmd.channelID].videos == undefined) {
		return false;
	} else if (this.storage[cmd.channelID].videos.cache == undefined) {
		return false;
	}
	return this.storage[cmd.channelID].videos.cache;
};

youtube.setVideosCache = function(cache) {
	if (this.storage[cmd.channelID] == undefined) {
		this.storage[cmd.channelID] = {};
	} else if (this.storage[cmd.channelID].videos == undefined) {
		this.storage[cmd.channelID].videos = {};
	}
	this.storage[cmd.channelID].videos.cache = cache;
	rug.saveStorage(this.storage);
};

youtube.addCommand("yt", function() {
	var action = cmd.args.length > 1 ? cmd.args[1].toLowerCase() : "playlist";

	if (action == "setkey") {
		if (cmd.args.length == 3) {
			rug.setAPIKey(cmd.args[2]);
			cmd.replyToAuthor("YouTube API key has been saved.");
			return;
		} else {
			cmd.replyToAuthor("usage: `setkey YOUR_API_KEY`");
			return;
		}
	}

	var api_key = rug.getAPIKey();
	if (api_key == false) {
		cmd.replyToAuthor("no API key provided.");
		return;
	}

	if (action == "setplaylist") {
		if (cmd.args.length == 3) {
			rug.setPlaylistID(cmd.args[2]);
			cmd.replyToAuthor("playlist `" + cmd.args[2] + "` has been saved.");
			return;
		} else {
			cmd.replyToAuthor("usage: `setplaylist PLAYLIST_ID`.");
			return;
		}
	} else if (action == "getplaylist") {
		var playlist = rug.getPlaylistID();
		if (playlist == false) {
			cmd.replyToAuthor("the playlist hasn't been set.");
			return;
		} else {
			cmd.replyToAuthor("the current playlist is `" + playlist + "`.");
			return;
		}
	}

	var videos = rug.getVideosCache();
	if (videos == false || action == "refreshplaylist") {
		rug.dprint("Updating video cache");
		videos = [];

		var playlistID = rug.getPlaylistID();

		var baseurl = "https://www.googleapis.com/youtube/v3/playlistItems?playlistId={0}&maxResults=50&part=contentDetails&key={1}".format(
			playlistID,
			api_key
		);
		var url = baseurl;

		for (var i = 0; i < 10; i++) {
			var content = HTTP.get(4, url);
			//rug.print(content);
			json = JSON.parse(content);
			for (var k in json.items) {
				var video = json.items[k];
				videos.push(video.contentDetails.videoId);
			}

			if (json.nextPageToken != undefined) {
				url = baseurl + "&pageToken=" + json.nextPageToken;
			} else {
				break;
			}
		}
		rug.setVideosCache(videos);
		videos = rug.getVideosCache();
		if (action == "refreshplaylist") {
			cmd.replyToAuthor("the playlist has been refreshed.");
		}
	} else {
		rug.dprint("Using video cache");
	}

	var r = Math.getRandomInRange(0, videos.length);

	//rug.print("amount of videos " + videos.length);
	cmd.replyToChannel("https://youtu.be/" + videos[r]);
});
