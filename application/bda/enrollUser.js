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

const enrollUser = async function enrollUser(user, userpw) {
    try {
        //var user = process.argv[2];
        //var userpw = process.argv[3];
        var org = 'bda';
        var caToUse = 'rca-bda';
        
        // Create a new CA client for interacting with the CA.
        const caURL = ccp.certificateAuthorities[caToUse].url;
        const ca = new FabricCAServices(caURL);

        // Create a new file system based wallet for managing identities.
        const walletPath = path.join(process.cwd(), 'wallet');
        const wallet = new FileSystemWallet(walletPath);
        console.log(`Wallet path: ${walletPath}`);

        // Check to see if we've already enrolled the user.
        const userExists = await wallet.exists(user);
        
        // Enroll the user, and import the new identity into the wallet.
        const enrollment = await ca.enroll({ enrollmentID: user, enrollmentSecret: userpw });
        const identity = X509WalletMixin.createIdentity('bdaMSP', enrollment.certificate, enrollment.key.toBytes());
        
        if (!userExists) {
            wallet.import(user, identity);
            return {enrolled: true};
        }
        console.log(`Successfully enrolled user "${user}" and imported it into the wallet`);
        return {enrolled: true};

    } catch (error) {
        console.error(`Failed to enroll  user "${user}": ${error}`);
        return {enrolled: false, err: error};
        process.exit(1);
    }
}

// enrollUser();}

module.exports = enrollUser;
