/*
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const FabricCAServices = require('fabric-ca-client');
const { FileSystemWallet, X509WalletMixin } = require('fabric-network');
const fs = require('fs');
const path = require('path');
const yaml = require('js-yaml')

// const ccpPath = path.resolve(__dirname, 'bda-connection.json');
// const ccpJSON = fs.readFileSync(ccpPath, 'utf8');
// const ccp = JSON.parse(ccpJSON);
let ccp = yaml.safeLoad(fs.readFileSync('connection.yml', 'utf8'));

async function main() {
    try {
        var user = process.argv[2];
        var userpw = process.argv[3];
        var org = 'la';
        var caToUse = 'rca-la';
        console.log(user);
        console.log(userpw);
        // Create a new CA client for interacting with the CA.
        const caURL = ccp.certificateAuthorities[caToUse].url;
        const ca = new FabricCAServices(caURL);

        // Create a new file system based wallet for managing identities.
        const walletPath = path.join(process.cwd(), 'wallet');
        const wallet = new FileSystemWallet(walletPath);
        console.log(`Wallet path: ${walletPath}`);

        // Check to see if we've already enrolled the admin user.
        const adminExists = await wallet.exists(user);
        if (adminExists) {
            console.log(`An identity for the admin user "${user}" already exists in the wallet`);
            return;
        }

        // Enroll the admin user, and import the new identity into the wallet.
        const enrollment = await ca.enroll({ enrollmentID: user, enrollmentSecret: userpw });
        const identity = X509WalletMixin.createIdentity('laMSP', enrollment.certificate, enrollment.key.toBytes());
        wallet.import(user, identity);
        console.log(`Successfully enrolled user "${user}" and imported it into the wallet`);

    } catch (error) {
        console.error(`Failed to enroll  user "${user}": ${error}`);
        process.exit(1);
    }
}

main();
