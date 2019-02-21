// Server sent Events server part

var express = require('express');
var app = express();
//eventListener = require("./eventListener");
var route = express.Router();

var connections = [];
route.get('/data', function (req, res) {
    req.socket.setTimeout(50000000);
    res.writeHead(200, {
        'Content-Type':'text/event-stream',
        'Cache-Control':'no-cache',
        'Connection':'keep-alive',
        "Access-Control-Allow-Origin": "*"
        })
    res.write('/n');
    connections.push(res);
    
    
    req.on("close", function () {
        var rem=0;
        for (let i=0; i<connections.length; i++) {
            if (connections[i] ==res) {
                rem=i;
                break;
            }
        }
        connections.splice(rem, 1)    
    });   
    
});

/*
setInterval(function () {
        val ++;
        connections.forEach(function (res) {
            var d=new Date();
            res.write(`data: ${d.getSeconds()}\n\n`);
            console.log(val);
        });
    }, 1000); */

app.use('/api', route);

app.listen(8000, function () {
    console.log("Listening on port 8000");
});


////////-------------------------------



/*
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { FileSystemWallet, Gateway } = require('fabric-network');
const Client = require('fabric-client');
const fs = require('fs');
const path = require('path');

var client = Client.loadFromConfig('connection.yml');

const clientCert = fs.readFileSync(path.join(__dirname, '../../layout-approval-network/data/tls/peer0-bda-cli-client.crt'));
const clientKey = fs.readFileSync(path.join(__dirname, '../../layout-approval-network/data/tls/peer0-bda-cli-client.key'));

client.setTlsClientCertAndKey(Buffer.from(clientCert).toString(), Buffer.from(clientKey).toString());

// async 
//var store_path = path.join(__dirname, 'wallet');
var member_user = null;
let eventPayload="";
var thisvar;
let eventListener = async function main(user) {
    try {
        //var user = process.argv[2];

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
            return {login: false, err: "User does not exist"};
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
            throw new Error('Failed to get user1.... run registerUser.js');
            
        }
        var test=[];
        var notfun = function a (event, block_num, txnid, status) {
            console.log('Successfully got a chaincode event with transid:'+ txnid + ' with status:'+status);
            console.log('Successfully received the chaincode event on block number '+ block_num);
            //console.log(event);
            var eventPayload = event.payload.toString('utf8');
            test.push(eventPayload);
            console.log (eventPayload);  
            console.log(test); 
            console.log("connections:",connections);
            connections.forEach(function (res) {                
                res.write(`data: eventPayload\n\n`);
            });  
                     
        };

        var channel = client.getChannel();
        var eventHub = channel.getChannelEventHubsForOrg('bda')[0];
        console.log('Successfully created eventHub');
        eventHub.connect(true);
        console.log('Successfully connected to eventHub');
        console.log('Listening...');
        eventHub.registerChaincodeEvent('mycc','mainEvent',
             (event, block_num, txnid, status) => {
                 notfun(event, block_num, txnid, status) },
            (error)=>{
                console.log('Failed to receive the chaincode event ::'+error);
                
            });
        

    } catch (error) {
        console.error(`Failed : ${error}`);
        
        process.exit(1);
    }
}


async function main () {
var a = await eventListener("user1");
}

main();


module.exports = eventListener;
