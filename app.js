const Discord = require("discord.js");
const Enmap = require("enmap");
const fs = require("fs");
const config = require('./cfg/config.json')
const chalk = require('chalk');

const client = new Discord.Client({
    intents: [
        Discord.GatewayIntentBits.GuildMessages,
        Discord.GatewayIntentBits.GuildMembers,
        Discord.GatewayIntentBits.DirectMessages,
        Discord.GatewayIntentBits.MessageContent,
        Discord.GatewayIntentBits.Guilds,
    ],
    partials: [
        Discord.Partials.Message,
        Discord.Partials.Channel,
        Discord.Partials.GuildMember,
        Discord.Partials.User,
        Discord.Partials.GuildScheduledEvent,
    ],
});

client.commands = new Enmap();

client.config = config;


fs.readdir("./events/", (err, files) => {
    if (err) return console.error(err);
        files.forEach(file => {
            const event = require(`./events/${file}`);
            let eventName = file.split(".")[0];
            client.on(eventName, event.bind(null, client));
            console.log(eventName, file)
        });
  });
  
  client.commands = new Enmap();
  
  fs.readdir("./commands/", (err, files) => {
    if (err) return console.error(err);
    files.forEach(file => {
        if (!file.endsWith(".js")) return;
        let props = require(`./commands/${file}`);
        let commandName = file.split(".")[0];
        console.log(chalk.green(`[+] ${commandName}`));
        client.commands.set(commandName, props);
        console.log(commandName, file)
    });
});

client.on("ready", () => {
    client.user.setActivity('Set Activity', { type: 'WATCHING' });
});

client.login(config.token)