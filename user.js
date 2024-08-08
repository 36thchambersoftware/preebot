const fs = require("fs");
require('dotenv').config();

class User {
    id
    stakeKey
    poolId
    stakeAmount

    constructor(author) {
        this.id = author.id;
        const path = process.env.USER_DATA_STORAGE_PATH;
        const filename = `${path}/${this.id}.json`;
        try {
            fs.readFile(this.id, (err, data) => {
                if (err) {
                    console.info(`creating profile for ${author.id}`);
                    const contents = JSON.stringify(this);
                    fs.writeFile(filename, contents, err => console.error(`could not create profile ${author.id}: ${err}`));
                    console.info(`profile successfully created for ${author.id}`);
                } else {
                    console.info(`profile exists for ${author.id}`);
                }
            });
        } catch (err) {
            console.error(`could not read file ${filename}: ${err}`);
        }
    }
}

module.exports = User;