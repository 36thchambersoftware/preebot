const { Client, Collection, Events, GatewayIntentBits, Partials } = require("discord.js");

const fs = require("node:fs");
const path = require('node:path')
require('dotenv').config()
const chalk = require('chalk');

const client = new Client({
    intents: [
        GatewayIntentBits.GuildMessages,
        GatewayIntentBits.GuildMembers,
        GatewayIntentBits.DirectMessages,
        GatewayIntentBits.MessageContent,
        GatewayIntentBits.Guilds,
    ],
    partials: [
        Partials.Message,
        Partials.Channel,
        Partials.GuildMember,
        Partials.User,
        Partials.GuildScheduledEvent,
    ],
});

client.commands = new Collection();

const foldersPath = path.join(__dirname, 'commands');
console.log(foldersPath)
const commandFolders = fs.readdirSync(foldersPath);
console.log(commandFolders)

for (const folder of commandFolders) {
	const commandsPath = path.join(foldersPath, folder);
	const commandFiles = fs.readdirSync(commandsPath).filter(file => file.endsWith('.js'));
	for (const file of commandFiles) {
		const filePath = path.join(commandsPath, file);
		const command = require(filePath);
        // TODO: update category to reflect a non-hardcoded path. 
        command.category = 'utility';
		// Set a new item in the Collection with the key as the command name and the value as the exported module
		if ('data' in command && 'execute' in command) {
			client.commands.set(command.data.name, command);
		} else {
			console.log(`[WARNING] The command at ${filePath} is missing a required "data" or "execute" property.`);
		}
	}
}

const eventsPath = path.join(__dirname, 'events');
const eventFiles = fs.readdirSync(eventsPath).filter(file => file.endsWith('.js'));

for (const file of eventFiles) {
	const filePath = path.join(eventsPath, file);
	const event = require(filePath);
	if (event.once) {
		client.once(event.name, (...args) => event.execute(...args));
	} else {
		client.on(event.name, (...args) => event.execute(...args));
	}
}

// fs.readdir("./events/", (err, files) => {
//     if (err) return console.error(err);
//         files.forEach(file => {
//             const event = require(`./events/${file}`);
//             let eventName = file.split(".")[0];
//             client.on(eventName, event.bind(null, client));
//             console.log(eventName, file)
//         });
//   });
  
//   client.commands = new Enmap();
  
//   fs.readdir("./commands/", (err, files) => {
//     if (err) return console.error(err);
//     files.forEach(file => {
//         if (!file.endsWith(".js")) return;
//         let props = require(`./commands/${file}`);
//         let commandName = file.split(".")[0];
//         console.log(chalk.green(`[+] ${commandName}`));
//         client.commands.set(commandName, props);
//         console.log(commandName, file)
//     });
// });


client.login(process.env.TOKEN)