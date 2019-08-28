package bot

// Instance : The current instance of the bot and its session
type Instance struct {
	clientSession *DiscordGoSession
	db            *DB
}
