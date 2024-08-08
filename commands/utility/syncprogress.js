// const Discord = require("discord.js"); // discord.js node module.
// const CardanoJs = require("cardanocli-js");
// const fs = require('node:fs');

// const cOptions = new CardanoJs.CardanoCliJsOptions({
//     shelleyGenesisPath: "/home/cardano/cardano/config/shelley-genesis.json",
//     socketPath: "/home/cardano/cardano/db/socket"
// });
// const C = new CardanoJs.CardanoCliJs(cOptions);

// const tip = C.query.tip();

// exports.run = (client, message, args) => {
//     if (message.author.bot === true) { return; }
//     console.log("A new message was written");

//     if (message.content === "!rename") {
//         message.guild.channels.fetch().then(
//             channels => {
//                 channels.forEach(channel => {
//                     if (channel.id == args) { // TODO: this probably doesn't work
//                         channel.setName(`Sync Progress: ${tip}`);
//                     }
//                 })
//             }
//         ).catch(error => console.error(error))
//     }
// }

// Client.login(token);