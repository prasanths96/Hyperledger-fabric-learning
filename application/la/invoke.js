/*
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { FileSystemWallet, Gateway } = require('fabric-network');
const Client = require('fabric-client');
const fs = require('fs');
const path = require('path');

var client = Client.loadFromConfig('connection.yml');

const clientCert = fs.readFileSync(path.join(__dirname, '../../layout-approval-network/data/tls/peer0-la-cli-client.crt'));
const clientKey = fs.readFileSync(path.join(__dirname, '../../layout-approval-network/data/tls/peer0-la-cli-client.key'));

client.setTlsClientCertAndKey(Buffer.from(clientCert).toString(), Buffer.from(clientKey).toString());
async function main() {
    try {
        var user = process.argv[2];
        var func = process.argv[3];
        var id = process.argv[4];

        // Create a new file system based wallet for managing identities.
        const walletPath = path.join(process.cwd(), 'wallet');
        const wallet = new FileSystemWallet(walletPath);
        console.log(`Wallet path: ${walletPath}`);

        // Check to see if we've already enrolled the user.
        const userExists = await wallet.exists(user);
        if (!userExists) {
            console.log(`An identity for the user "${user}" does not exist in the wallet`);
            console.log('Run the registerUser.js application before retrying');
            return;
        }

        
        // Create a new gateway for connecting to our peer node.
        const gateway = new Gateway();
        //await gateway.connect(ccp, { wallet, identity: 'user1', discovery: { enabled: false } });
        await gateway.connect(client, { wallet, identity: user, discovery: { enabled: false } });

        // Get the network (channel) our contract is deployed to.
        const network = await gateway.getNetwork('mychannel');

        // Get the contract from the network.
        const contract = network.getContract('mycc');

        // Submit the specified transaction.
        // createCar transaction - requires 5 argument, ex: ('createCar', 'CAR12', 'Honda', 'Accord', 'Black', 'Tom')
        // changeCarOwner transaction - requires 2 args , ex: ('changeCarOwner', 'CAR10', 'Dave')
        await contract.submitTransaction(func, id);
        console.log('Transaction has been submitted');

        // Disconnect from the gateway.
        await gateway.disconnect(); 

    } catch (error) {
        console.error(`Failed to submit transaction: ${error}`);
        process.exit(1);
    }
}

main();
