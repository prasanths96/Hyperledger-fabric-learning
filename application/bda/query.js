/*
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { FileSystemWallet, Gateway } = require('fabric-network');
const Client = require('fabric-client');
const fs = require('fs');
const path = require('path');
const yaml = require('js-yaml')

//const ccpPath = path.resolve(__dirname, '..', '..', 'basic-network', 'connection.json');
//const ccpJSON = fs.readFileSync(ccpPath, 'utf8');
//const ccp = JSON.parse(ccpJSON);
// let ccp = yaml.safeLoad(fs.readFileSync('connection.yml', 'utf8'));
var client = Client.loadFromConfig('./connection.yml');

const clientCert = fs.readFileSync(path.join(__dirname, '../../layout-approval-network/data/tls/peer0-bda-cli-client.crt'));
const clientKey = fs.readFileSync(path.join(__dirname, '../../layout-approval-network/data/tls/peer0-bda-cli-client.key'));

client.setTlsClientCertAndKey(Buffer.from(clientCert).toString(), Buffer.from(clientKey).toString());

const query = async function main(user, func, id) {
    try {
        //var user = process.argv[2];
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
        // await gateway.connect(ccp, { wallet, identity: 'user1', discovery: { enabled: false } });
        await gateway.connect(client, { wallet, identity: user, discovery: { enabled: false } });

        // Get the network (channel) our contract is deployed to.
        const network = await gateway.getNetwork('mychannel');

        // Get the contract from the network.
        const contract = network.getContract('mycc');

        // Evaluate the specified transaction.
        // queryCar transaction - requires 1 argument, ex: ('queryCar', 'CAR4')
        // queryAllCars transaction - requires no arguments, ex: ('queryAllCars')
        // const result = await contract.evaluateTransaction('getHistory', 'ALL_TRANSACTION_HISTORY');
        // const result = await contract.evaluateTransaction(process.argv[3], process.argv[4]);
        
        const result = await contract.evaluateTransaction(func, id);
        console.log(`Transaction has been evaluated, result is:\n\n ${result.toString()}`);

        // for server
        //return result.toString();
        // return {success: true, result: result.toString()}

    } catch (error) {
        console.error(`Failed to evaluate transaction: ${error}`);
        return {success: false, err: error}
        process.exit(1);
    }
}

query(process.argv[2], process.argv[3], process.argv[4]);

module.exports = query;
