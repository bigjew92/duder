# Duder - Discord Bot

Duder is a Discord bot written in pure Go with support for slash commands.

## Requirements

- Go 1.20 or later
- A Discord bot token

## Quick Start

1. **Build the bot:**
   ```bash
   go build -o duder .
   ```

2. **Configure the bot:**
   Create a `config.json` file:
   ```json
   {
     "botToken": "YOUR_DISCORD_BOT_TOKEN",
     "ownerID": "YOUR_DISCORD_USER_ID",
     "commandPrefix": "!d",
     "status": "The Dude abides",
     "avatarsPath": "avatars",
     "permissionsFile": "permissions.json",
     "rugsPath": "rugs"
   }
   ```

3. **Run the bot:**
   ```bash
   ./duder
   ```

## Docker

### Using Docker directly
```bash
docker build -t duder .
docker run -v $(pwd)/config.json:/app/config.json duder
```

### Using Docker Compose (Recommended)

1. **Copy the example environment file:**
   ```bash
   cp .env.example .env
   ```

2. **Edit `.env` with your tokens and API keys:**
   ```bash
   nano .env
   ```
   
   Required variables:
   | Variable | Description |
   |----------|-------------|
   | `BOT_TOKEN` | Your Discord bot token |
   | `OWNER_ID` | Your Discord user ID |

   Optional API keys (for specific commands):
   | Variable | Command | Get Key At |
   |----------|---------|------------|
   | `WEATHER_API_KEY` | `/weather` | [openweathermap.org](https://openweathermap.org/api) |
   | `IMGFLIP_USERNAME` | `/meme` | [imgflip.com](https://imgflip.com/signup) |
   | `IMGFLIP_PASSWORD` | `/meme` | [imgflip.com](https://imgflip.com/signup) |
   | `OPENXBL_API_KEY` | `/xbl` | [xbl.io](https://xbl.io/) |
   | `YOUTUBE_API_KEY` | `/youtube` | [Google Cloud Console](https://console.cloud.google.com/apis/credentials) |

3. **Start the bot:**
   ```bash
   docker-compose up -d
   ```

4. **View logs:**
   ```bash
   docker-compose logs -f
   ```

5. **Stop the bot:**
   ```bash
   docker-compose down
   ```

## Discord Bot Setup (First-Time Setup)

If you've never created a Discord bot before, follow these steps:

### Step 1: Create a Discord Application

1. Go to the [Discord Developer Portal](https://discord.com/developers/applications)
2. Click the **"New Application"** button (top right)
3. Give your application a name (e.g., "Duder") and click **Create**

### Step 2: Create the Bot User

1. In your application, click **"Bot"** in the left sidebar
2. Click **"Add Bot"** and confirm
3. Under **"Token"**, click **"Reset Token"** and then **"Copy"**
4. Save this token securely - this is your `botToken` for `config.json`

> **Never share your bot token!** If leaked, anyone can control your bot.

### Step 3: Enable Required Intents

On the same **Bot** page, scroll down to **"Privileged Gateway Intents"** and enable:

| Intent | Required For |
|--------|-------------|
| **Presence Intent** | `/lastseen` command (tracking online status) |
| **Server Members Intent** | Accessing member information |
| **Message Content Intent** | Reading message content for event handlers |

Click **"Save Changes"** at the bottom.

### Step 4: Configure OAuth2 Scopes & Permissions

1. Click **"OAuth2"** in the left sidebar
2. Click **"URL Generator"**

**Select these SCOPES** (checkboxes):

| Scope | Purpose |
|-------|---------|
| `bot` | Allows the bot to join servers |
| `applications.commands` | **Required** for slash commands to appear |

**Select these BOT PERMISSIONS** (checkboxes that appear after selecting `bot`):

| Permission | Purpose |
|------------|---------|
| Send Messages | Reply to commands |
| Send Messages in Threads | Reply in threads |
| Embed Links | Send rich embed messages |
| Attach Files | Send images (cat pics, memes) |
| Read Message History | Access message context |
| Add Reactions | React to messages |
| Use External Emojis | Use emojis from other servers |
| Change Nickname | Allow nickname changes |

For full functionality, you can also select:
| Permission | Purpose |
|------------|---------|
| Manage Messages | Delete messages (optional) |
| Manage Roles | Manage user roles (optional) |

### Step 5: Invite the Bot to Your Server

1. After selecting scopes and permissions, scroll down to see the **Generated URL**
2. Click **"Copy"**
3. Paste the URL in your browser
4. Select the server you want to add the bot to
5. Click **Authorize**

### Step 6: Get Your User ID (for ownerID)

1. In Discord, go to **User Settings** → **Advanced** → Enable **Developer Mode**
2. Right-click your username anywhere in Discord
3. Click **"Copy User ID"**
4. This is your `ownerID` for `config.json`

### Quick Links

| Page | URL |
|------|-----|
| Discord Developer Portal | https://discord.com/developers/applications |
| Your Applications | https://discord.com/developers/applications |
| OAuth2 URL Generator | Select your app → OAuth2 → URL Generator |
| Bot Settings | Select your app → Bot |

---

## Available Commands

| Command | Description |
|---------|-------------|
| `/ping` | Check if bot is alive |
| `/status` | Set bot status (owner only) |
| `/avatar` | Manage bot avatar (owner only) |
| `/bananaball` | Banana Ball player/team stats lookup |
| `/catfact` | Random cat fact |
| `/catpic` | Random cat picture |
| `/catgif` | Random cat GIF |
| `/dice` | Roll a dice |
| `/8ball` | Ask the magic 8-ball |
| `/lebowski` | Random Big Lebowski quote |
| `/dadjoke` | Random dad joke |
| `/fact` | Get a useless fact (random or today) |
| `/big` | Convert text to big emoji letters |
| `/smol` | Convert text to tiny letters |
| `/aura` | Random Aura Copypasta |
| `/lastseen` | Check when a user was last active |
| `/weather` | Get weather forecast (use `location:` option) |
| `/meme` | Create a meme |
| `/mlb` | MLB player/team stats lookup |
| `/nhl` | NHL player/team stats lookup |
| `/starwars` | Star Wars databank lookup |
| `/xbl` | Xbox Live profile lookup |
| `/youtube` | Random video from playlist |

---

## Adding a New Command

Creating a new slash command is simple. Here's a step-by-step guide:

### 1. Create a new file in `rugs/`

Create a file like `rugs/hello.go`:

```go
package rugs

import (
    "github.com/bwmarrin/discordgo"
)

func init() {
    Register(&HelloCommand{})
}

type HelloCommand struct{}

func (c *HelloCommand) Name() string {
    return "hello"
}

func (c *HelloCommand) Description() string {
    return "Say hello!"
}

func (c *HelloCommand) Options() []*discordgo.ApplicationCommandOption {
    return []*discordgo.ApplicationCommandOption{
        {
            Type:        discordgo.ApplicationCommandOptionString,
            Name:        "name",
            Description: "Name to greet",
            Required:    false,
        },
    }
}

func (c *HelloCommand) Execute(ctx CommandContext) error {
    name := ctx.GetString("name")
    if name == "" {
        name = ctx.User().Username
    }
    return ctx.Reply("Hello, " + name + "! 👋")
}
```

### 2. Rebuild and run

```bash
go build -o duder . && ./duder
```

That's it! The command will automatically register with Discord.

### Command Context Helpers

The `CommandContext` interface provides these helper methods:

| Method | Description |
|--------|-------------|
| `ctx.Reply(text)` | Send a text response |
| `ctx.ReplyEmbed(embed)` | Send an embed response |
| `ctx.ReplyEphemeral(text)` | Send a private response |
| `ctx.DeferReply()` | Acknowledge interaction (for slow commands) |
| `ctx.FollowUp(text)` | Send follow-up after DeferReply |
| `ctx.FollowUpEmbed(embed)` | Send embed follow-up |
| `ctx.GetString(name)` | Get a string option |
| `ctx.GetInt(name)` | Get an integer option |
| `ctx.GetUser(name)` | Get a user option |
| `ctx.IsOwner()` | Check if user is bot owner |

### Using Storage

For persistent data, use the `Storage` helper:

```go
type MyCommand struct {
    storage *Storage
}

func (c *MyCommand) getStorage() *Storage {
    if c.storage == nil {
        c.storage = NewStorage("mycommand") // Creates rugs/mycommand.json
        c.storage.Load()
    }
    return c.storage
}

func (c *MyCommand) Execute(ctx CommandContext) error {
    storage := c.getStorage()
    
    // Read
    value := storage.GetString("key")
    
    // Write
    storage.Set("key", "value")
    storage.Save()
    
    return ctx.Reply("Done!")
}
```

### Event Handlers

To handle events like messages or reactions, implement the optional interfaces:

```go
// Handle all messages (like lastseen tracking)
func (c *MyCommand) OnMessage(ctx *MessageContext) {
    // Called for every message
}

// Handle reaction adds (like XP system)
func (c *MyCommand) OnReactionAdd(ctx *ReactionContext) {
    // Called when a reaction is added
}

func (c *MyCommand) OnReactionRemove(ctx *ReactionContext) {
    // Called when a reaction is removed
}
```

---

## License

GNU General Public License v3.0
