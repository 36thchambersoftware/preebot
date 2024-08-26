const { SlashCommandBuilder } = require('discord.js');

module.exports = {
    cooldown: 10,
    data: new SlashCommandBuilder().setName('ping').setDescription('Replies with pong'),
    async execute(interaction) {
        await interaction.editReply('Pong!');
    }
}