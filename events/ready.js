const chalk = require("chalk")
const { Events } = require('discord.js');

module.exports = {
	name: Events.ClientReady,
	once: true,
	execute(client) {
		console.info(`Ready! Logged in as ${client.user.tag}`);
        console.log(chalk.magenta(`Bot Made by PREEB \nCommand prefix is ${process.env.PREFIX}`));
	},
};