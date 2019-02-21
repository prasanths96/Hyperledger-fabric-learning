/*
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const invoke = require('./autoInvoke');

const { FileSystemWallet, Gateway } = require('fabric-network');
const Client = require('fabric-client');
const fs = require('fs');
const path = require('path');
// const yaml = require('js-yaml')

// let ccp = yaml.safeLoad(fs.readFileSync('connection.yml', 'utf8'));
var client = Client.loadFromConfig('connection.yml');

const clientCert = fs.readFileSync(path.join(__dirname, '../../layout-approval-network/data/tls/peer0-fa-cli-client.crt'));
const clientKey = fs.readFileSync(path.join(__dirname, '../../layout-approval-network/data/tls/peer0-fa-cli-client.key'));

client.setTlsClientCertAndKey(Buffer.from(clientCert).toString(), Buffer.from(clientKey).toString());

// async 
//var store_path = path.join(__dirname, 'wallet');
var member_user = null;

async function main() {
    try {
        var user = process.argv[2];

        // Create a new file system based wallet for managing identities.
        const walletPath = path.join(process.cwd(), 'wallet');
        const wallet = new FileSystemWallet(walletPath);
        console.log(`Wallet path: ${walletPath}`);
        
        var store_path = path.join(__dirname, `wallet/${user}`);
        // Check to see if we've already enrolled the user.
        const userExists = await wallet.exists(user);
        if (!userExists) {
            console.log(`An identity for the user "${user}" does not exist in the wallet`);
            console.log('Run the registerUser.js application before retrying');
            return;
        }

        //... client configuration
        var state_store = await Client.newDefaultKeyValueStore({ path: store_path});
        client.setStateStore(state_store);
        var crypto_suite = Client.newCryptoSuite();
        // use the same location for the state store (where the users' certificate are kept)
        // and the crypto store (where the users' keys are kept)
        var crypto_store = Client.newCryptoKeyStore({path: store_path});
        crypto_suite.setCryptoKeyStore(crypto_store);
        client.setCryptoSuite(crypto_suite);

        var user_from_store = await client.getUserContext(user, true);

        if (user_from_store && user_from_store.isEnrolled()) {
            console.log(`Successfully loaded ${user} from persistence`);
            member_user = user_from_store;
        } else {
            throw new Error(`Failed to get ${user}.... run registerUser.js`);
        }

        var channel = client.getChannel();
        var eventHub = channel.getChannelEventHubsForOrg('fa')[0];
        console.log('Successfully loaded eventHub');
        eventHub.connect(true);
        console.log('Successfully connected to eventHub');
        console.log('Listening...');
        eventHub.registerChaincodeEvent('mycc','mainEvent',
            (event, block_num, txnid, status)=>{
                console.log('Successfully got a chaincode event with transid:'+ txnid + ' with status:'+status);
                console.log('Successfully received the chaincode event on block number '+ block_num);
                //console.log(event);
                var eventPayload = event.payload.toString('utf8');
                console.log(eventPayload);

                var doneid = '';
                // Invoke 
                if((eventPayload.indexOf("NOC Requested for layout: ") > -1) &&
                    doneid != txnid){ 
                    // Strip the Layout ID
                    // var id = eventPayload.split(": ").pop();                                        
                    var startIndex = eventPayload.indexOf(":");
                    startIndex += 2;
                    var endIndex = eventPayload.length;
                    //endIndex = endIndex - 1;
                    var id = eventPayload.substring(startIndex, endIndex);
                    console.log('ID stripped:' + id);
                    //
                    //
                    // invoke
                    console.log('Invoking chaincode to Approve or Reject');
                    invoke.approve(id);
                    // tx
                    doneid = txnid;
                    console.log(txnid);
                }
            },
            (error)=>{
                console.log('Failed to receive the chaincode event ::'+error);
            });
        

    } catch (error) {
        console.error(`Failed : ${error}`);
        process.exit(1);
    }
}

main();
