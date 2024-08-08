const discord = require ("discord.js");

exports.run = (client, message, args) =>{
    const help = new discord.EmbedBuilder()
    .setColor('#b434eb')
    .setTitle('PREEBOT')
    .setURL("https://github.com/36thchambersoftware/preebot")
    .addFields({name: "Profile Setup", value: "Your profile has been saved!"})
    .setFooter({text: "PREEB", iconURL: "https://avatars.githubusercontent.com/u/1509069?v=4"})
    message.channel.send({embeds: [help] });
};
