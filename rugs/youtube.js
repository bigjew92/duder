var youtube = new DuderRug("YouTube", "Do stuff with YouTube.");
youtube.storage = youtube.loadStorage();

youtube.getAPIKey = function() {
	if (this.storage.settings === undefined) {
		return false;
	} else if (this.storage.settings.api_key === undefined) {
		return false;
	}
	return this.storage.settings.api_key;
};

youtube.setAPIKey = function(key) {
	if (this.storage.settings === undefined) {
		this.storage.settings = {};
	}
	this.storage.settings.api_key = key;
	this.saveStorage(this.storage);
};

youtube.getPlaylistID = function(guildID) {
	if (this.storage[guildID] === undefined) {
		return false;
	} else if (this.storage[guildID].playlist === undefined) {
		return false;
	}
	return this.storage[guildID].playlist;
};

youtube.setPlaylistID = function(guildID, id) {
	if (this.storage[guildID] === undefined) {
		this.storage[guildID] = {};
	}
	this.storage[guildID].playlist = id;
	this.saveStorage(this.storage);
};

youtube.getVideosCache = function(guildID) {
	if (this.storage[guildID] === undefined) {
		return false;
	} else if (this.storage[guildID].videos === undefined) {
		return false;
	} else if (this.storage[guildID].videos.cache === undefined) {
		return false;
	}
	return this.storage[guildID].videos.cache;
};

youtube.setVideosCache = function(guildID, cache) {
	if (this.storage[guildID] === undefined) {
		this.storage[guildID] = {};
	}
	if (this.storage[guildID].videos === undefined) {
		this.storage[guildID].videos = {};
	}
	this.storage[guildID].videos.cache = cache;
	this.saveStorage(this.storage);
};

youtube.addCommand("yt", function(cmd) {
	var action = cmd.args.length > 1 ? cmd.args[1].toLowerCase() : "playlist";

	if (action === "setkey") {
		if (cmd.args.length === 3) {
			this.setAPIKey(cmd.args[2]);
			cmd.replyToAuthor("YouTube API key has been saved.");
			return;
		} else {
			cmd.replyToAuthor("usage: `setkey YOUR_API_KEY`");
			return;
		}
	}

	var api_key = this.getAPIKey();
	if (api_key === false) {
		cmd.replyToAuthor("no YouTube API key provided.");
		return;
	}

	if (action === "setplaylist") {
		if (cmd.args.length === 3) {
			this.setPlaylistID(cmd.guildID, cmd.args[2]);
			cmd.replyToAuthor("playlist `" + cmd.args[2] + "` has been saved.");
			return;
		} else {
			cmd.replyToAuthor("usage: `setplaylist PLAYLIST_ID`.");
			return;
		}
	} else if (action === "getplaylist") {
		var playlist = this.getPlaylistID(cmd.guildID);
		if (playlist === false) {
			cmd.replyToAuthor("the playlist hasn't been set.");
			return;
		} else {
			cmd.replyToAuthor("the current playlist is `" + playlist + "`.");
			return;
		}
	}

	var videos = this.getVideosCache(cmd.guildID);
	if (videos === false || action === "refreshplaylist") {
		this.dprint("Updating video cache");
		videos = [];

		var playlistID = this.getPlaylistID(cmd.guildID);
		if (playlistID === false) {
			cmd.replyToAuthor("the playlist hasn't been set.");
			return;
		}
		var baseurl = "https://www.googleapis.com/youtube/v3/playlistItems?playlistId={0}&maxResults=50&part=contentDetails&key={1}".format(
			playlistID,
			api_key
		);
		var url = baseurl;
		for (var i = 0; i < 10; i++) {
			var content = HTTP.get(4, url);
			var json = JSON.parse(content);
			for (var k in json.items) {
				var video = json.items[k];
				if (video.contentDetails !== undefined) {
					videos.push(video.contentDetails.videoId);
				}
			}

			if (json.nextPageToken !== undefined) {
				url = baseurl + "&pageToken=" + json.nextPageToken;
			} else {
				break;
			}
		}
		this.setVideosCache(cmd.guildID, videos);
		videos = this.getVideosCache(cmd.guildID);
		if (action === "refreshplaylist") {
			cmd.replyToAuthor("the playlist has been refreshed.");
		}
	} else {
		this.dprint("Using video cache");
	}

	var r = Math.getRandomInRange(0, videos.length);

	//this.print("amount of videos " + videos.length);
	cmd.replyToChannel("https://youtu.be/" + videos[r]);
});
