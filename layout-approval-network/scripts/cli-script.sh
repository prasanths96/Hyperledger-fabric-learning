function main(){
 # common
 export ORDERER_PORT_ARGS="-o orderer0:7050 --tls --cafile /data/ca-certs/rca-orderer.pem --clientauth"
# createChannel
initPeerVars bda 0 bdaMSP /data/ca-certs/rca-bda.pem 
switchToAdminIdentity /data/orgmspdirs/bda/admin/
export CHANNEL_NAME=mychannel
export CHANNEL_TX_FILE=/data/channel.tx
peer channel create -c $CHANNEL_NAME -f $CHANNEL_TX_FILE $ORDERER_CONN_ARGS



# Join channel
	# bda
	initPeerVars bda 0 bdaMSP /data/ca-certs/rca-bda.pem 
	switchToAdminIdentity /data/orgmspdirs/bda/admin/
	export CHANNEL_NAME=mychannel
	peer channel join -b $CHANNEL_NAME.block
	
	# fa
	initPeerVars fa 0 faMSP /data/ca-certs/rca-fa.pem 
	switchToAdminIdentity /data/orgmspdirs/fa/admin/
	export CHANNEL_NAME=mychannel
	peer channel join -b $CHANNEL_NAME.block

	# la
	initPeerVars la 0 laMSP /data/ca-certs/rca-la.pem 
	switchToAdminIdentity /data/orgmspdirs/la/admin/
	export CHANNEL_NAME=mychannel
	peer channel join -b $CHANNEL_NAME.block


# Update anchors
	# bda
	initPeerVars bda 0 bdaMSP /data/ca-certs/rca-bda.pem 
	switchToAdminIdentity /data/orgmspdirs/bda/admin/
	export CHANNEL_NAME=mychannel
	export ANCHOR_TX_FILE=/data/orgmspdirs/bda/anchors.tx
	peer channel update -c $CHANNEL_NAME -f $ANCHOR_TX_FILE $ORDERER_CONN_ARGS

	# fa
	initPeerVars fa 0 faMSP /data/ca-certs/rca-fa.pem 
	switchToAdminIdentity /data/orgmspdirs/fa/admin/
	export CHANNEL_NAME=mychannel
	export ANCHOR_TX_FILE=/data/orgmspdirs/fa/anchors.tx
	peer channel update -c $CHANNEL_NAME -f $ANCHOR_TX_FILE $ORDERER_CONN_ARGS

	# la
	initPeerVars la 0 laMSP /data/ca-certs/rca-la.pem 
	switchToAdminIdentity /data/orgmspdirs/la/admin/
	export CHANNEL_NAME=mychannel
	export ANCHOR_TX_FILE=/data/orgmspdirs/la/anchors.tx
	peer channel update -c $CHANNEL_NAME -f $ANCHOR_TX_FILE $ORDERER_CONN_ARGS

# Install chaincode
	# bda
	initPeerVars bda 0 bdaMSP /data/ca-certs/rca-bda.pem 
	switchToAdminIdentity /data/orgmspdirs/bda/admin/
	peer chaincode install -n mycc -v 1.0 -p github.com/hyperledger/chaincode/lacc

	# fa
	initPeerVars fa 0 faMSP /data/ca-certs/rca-fa.pem 
	switchToAdminIdentity /data/orgmspdirs/fa/admin/
	peer chaincode install -n mycc -v 1.0 -p github.com/hyperledger/chaincode/lacc

	# la
	initPeerVars la 0 laMSP /data/ca-certs/rca-la.pem 
	switchToAdminIdentity /data/orgmspdirs/la/admin/
	peer chaincode install -n mycc -v 1.0 -p github.com/hyperledger/chaincode/lacc
	

# Instantiate 	
	# Copying scripts to current directory in container for easy access 
	cp /scripts/initPeerVars ./
	cp /scripts/switchToAdminIdentity ./
	cp /scripts/switchToUserIdentity ./
	cp /scripts/registerUser ./
	
	# bda
	initPeerVars bda 0 bdaMSP /data/ca-certs/rca-bda.pem 
	switchToAdminIdentity /data/orgmspdirs/bda/admin/
	POLICY="OR ('bdaMSP.member','faMSP.member','laMSP.member')"
	CHANNEL_NAME=mychannel
	peer chaincode instantiate -C $CHANNEL_NAME -n mycc -v 1.0 -c '{"Args":[]}' -P "$POLICY" $ORDERER_CONN_ARGS


}


# Switch to the current org's user identity.  Enroll if not previously enrolled.
function switchToUserIdentity {

   	ORG=$1
   	CA_CHAINFILE=$2
	USER_NAME=$3
	USER_PASS=$4
	CA_HOST=$5
	ORG_ADMIN_HOME=$6

	export FABRIC_CA_CLIENT_HOME=/etc/hyperledger/fabric/orgs/$ORG/user
	mkdir -p /etc/hyperledger/fabric/orgs/$ORG/user
	export CORE_PEER_MSPCONFIGPATH=$FABRIC_CA_CLIENT_HOME/msp

	export FABRIC_CA_CLIENT_TLS_CERTFILES=$CA_CHAINFILE
	fabric-ca-client enroll -d -u https://$USER_NAME:$USER_PASS@$CA_HOST:7054

	# Set up admincerts directory if required	     
	ACDIR=$CORE_PEER_MSPCONFIGPATH/admincerts
	mkdir -p $ACDIR
	cp $ORG_ADMIN_HOME/msp/signcerts/* $ACDIR
    
}


function switchToAdminIdentity {     
      ORG_ADMIN_HOME=$1     
   export CORE_PEER_MSPCONFIGPATH=$ORG_ADMIN_HOME/msp
}



function initPeerVars {
   ORG=$1
   NUM=$2
   ORG_MSP_ID=$3
   CA_CHAINFILE=$4

   DATA=data
   PEER_HOST=peer${NUM}-${ORG}
   PEER_NAME=peer${NUM}-${ORG}
   PEER_PASS=adminpw
   PEER_NAME_PASS=${PEER_NAME}:${PEER_PASS}
   MYHOME=/opt/gopath/src/github.com/hyperledger/fabric/peer
   TLSDIR=$MYHOME/tls

   export FABRIC_CA_CLIENT=$MYHOME
   export CORE_PEER_ID=$PEER_HOST
   export CORE_PEER_ADDRESS=$PEER_HOST:7051
   export CORE_PEER_LOCALMSPID=$ORG_MSP_ID
   export CORE_VM_ENDPOINT=unix:///host/var/run/docker.sock
   # the following setting starts chaincode containers on the same
   # bridge network as the peers
   # https://docs.docker.com/compose/networking/
   #export CORE_VM_DOCKER_HOSTCONFIG_NETWORKMODE=${COMPOSE_PROJECT_NAME}_${NETWORK}
   #export CORE_VM_DOCKER_HOSTCONFIG_NETWORKMODE=net_${NETWORK}
   # export FABRIC_LOGGING_SPEC=ERROR
   export FABRIC_LOGGING_SPEC=DEBUG
   export CORE_PEER_TLS_ENABLED=true
   export CORE_PEER_TLS_CLIENTAUTHREQUIRED=true
   export CORE_PEER_TLS_ROOTCERT_FILE=$CA_CHAINFILE
   export CORE_PEER_TLS_CLIENTCERT_FILE=/$DATA/tls/$PEER_NAME-cli-client.crt
   export CORE_PEER_TLS_CLIENTKEY_FILE=/$DATA/tls/$PEER_NAME-cli-client.key
   export CORE_PEER_PROFILE_ENABLED=true
   # gossip variables
   export CORE_PEER_GOSSIP_USELEADERELECTION=true
   export CORE_PEER_GOSSIP_ORGLEADER=false
   export CORE_PEER_GOSSIP_EXTERNALENDPOINT=$PEER_HOST:7051

      # Point the non-anchor peers to the anchor peer, which is always the 1st peer
      export CORE_PEER_GOSSIP_BOOTSTRAP=peer0-${ORG}:7051

   export ORDERER_CONN_ARGS="$ORDERER_PORT_ARGS --keyfile $CORE_PEER_TLS_CLIENTKEY_FILE --certfile $CORE_PEER_TLS_CLIENTCERT_FILE"
}

main
