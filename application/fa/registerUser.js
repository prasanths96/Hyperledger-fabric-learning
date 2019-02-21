/*
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { FileSystemWallet, Gateway, X509WalletMixin } = require('fabric-network');
const fs = require('fs');
const path = require('path');
const yaml = require('js-yaml')

let ccp = yaml.safeLoad(fs.readFileSync('connection.yml', 'utf8'));

async function main() {
    try {
        var user = process.argv[2];
        var userpw = process.argv[3];
        var org = 'fa';
        var caToUse = `rca-${org}`;
        var adminToUse = `${caToUse}-admin`;
        // Create a new file system based wallet for managing identities.
        const walletPath = path.join(process.cwd(), 'wallet');
        const wallet = new FileSystemWallet(walletPath);
        console.log(`Wallet path: ${walletPath}`);

        // Check to see if we've already enrolled the user.
        const userExists = await wallet.exists(user);
        if (userExists) {
            console.log(`An identity for the user "${user}" already exists in the wallet`);
            return;
        }

        // Check to see if we've already enrolled the admin user.
        const adminExists = await wallet.exists(adminToUse);
        if (!adminExists) {
            console.log(`An identity for the admin user "${adminToUse}" does not exist in the wallet`);
            console.log('Run the enrollAdmin.js application before retrying');
            return;
        }

        // Create a new gateway for connecting to our peer node.
        const gateway = new Gateway();
        await gateway.connect(ccp, { wallet, identity: adminToUse, discovery: { enabled: false } });

        // Get the CA client object from the gateway for interacting with the CA.
        const ca = gateway.getClient().getCertificateAuthority();
        const adminIdentity = gateway.getCurrentIdentity();

        // Attribute
        //const attribute 
    
        console.log('No prob until this');
        // Register the user, enroll the user, and import the new identity into the wallet.
        const secret = await ca.register({ affiliation: org, maxEnrollments:0, enrollmentID: user, enrollmentSecret: userpw, role: 'client', attrs: [{name:'lan.role', value: org, ecert: true}]}, adminIdentity);
        console.log('registration step over');
        const enrollment = await ca.enroll({ enrollmentID: user, enrollmentSecret: secret });
        console.log('enrollment step over');
        const userIdentity = X509WalletMixin.createIdentity(`${org}MSP`, enrollment.certificate, enrollment.key.toBytes());
        wallet.import(user, userIdentity);
        console.log(`Successfully registered and enrolled admin user "${user}" and imported it into the wallet`);

    } catch (error) {
        console.error(`Failed to register user "${user}": ${error}`);
        process.exit(1);
    }
}

main();
