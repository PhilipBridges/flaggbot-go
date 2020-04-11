# Flaggbot-2.0

A discord bot that uses a MongoDB for basic storage. Uses text from your discord's text chat to generate markov chains (updates with every message).

1. Install packages
2. Edit config.json to your liking
3. Add your mongoDB address to the main.go file
4. In the AutoClean function (found in handlers/messageHandler), add your text channel ID in place of the default.
5. Run the bot with the "-t _yourtoken_" flag, or add the token to the main.go file/config.json file.

Audio functions are thanks to https://github.com/romainisnel/hertz

Commands are: 
- !stop (stop all audio)
- !create (profile)
- !bet <amount>
- !clean (delete last 50 bot related messages)
- !fbux
- !broke (irreversibly resets flaggbux to 100)
- !meme
- !memecount
- !u (sound)
- !gear (sound)
- !youtube <link>
- !mark (generates markov model from your discord text chat)
- !gen (generates a markov chain)
