const chalk = require("chalk")
const config = require("../cfg/config.json")
module.exports = (client) => {
    console.log(chalk.magenta(`Bot Made by PREEB \nCommand prefix is ${config.prefix}`));
}
