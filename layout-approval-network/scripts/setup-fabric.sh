function main(){
# Register orderer identity:
	# Enroll CA Admin first
	export FABRIC_CA_CLIENT_HOME=$HOME/ca-admins/rca-orderer
   	export FABRIC_CA_CLIENT_TLS_CERTFILES=/data/ca-certs/rca-orderer.pem
   	fabric-ca-client enroll -d -u https://rca-orderer-admin:adminpw@rca-orderer:7054
ORDERER_NAME=orderer0
ORDERER_PASS=adminpw
fabric-ca-client register -d --id.name $ORDERER_NAME --id.secret $ORDERER_PASS --id.type orderer

# The admin identity has the "admin" attribute which is added to ECert by default
fabric-ca-client register -d --id.name orderer-admin --id.secret adminpw --id.attrs "admin=true:ecert"



# Register Peer identities:
	# Enroll CA Admin first
	export FABRIC_CA_CLIENT_HOME=$HOME/ca-admins/rca-bda
   	export FABRIC_CA_CLIENT_TLS_CERTFILES=/data/ca-certs/rca-bda.pem
   	fabric-ca-client enroll -d -u https://rca-bda-admin:adminpw@rca-bda:7054
	PEER_NAME=peer0-bda
	PEER_PASS=adminpw
	fabric-ca-client register -d --id.name $PEER_NAME --id.secret $PEER_PASS --id.type peer

	# The admin identity has the "admin" attribute which is added to ECert by default
	ADMIN_NAME=bda-admin
	ADMIN_PASS=adminpw
	USER_NAME=bda-user0
	USER_PASS=userpw
	fabric-ca-client register -d --id.name $ADMIN_NAME --id.secret $ADMIN_PASS --id.attrs "hf.Registrar.Roles=client,hf.Registrar.Attributes=*,hf.Revoker=true,hf.GenCRL=true,admin=true:ecert,abac.init=true:ecert"
	fabric-ca-client register -d --id.name $USER_NAME --id.secret $USER_PASS




	# Enroll CA Admin first
	export FABRIC_CA_CLIENT_HOME=$HOME/ca-admins/rca-fa
   	export FABRIC_CA_CLIENT_TLS_CERTFILES=/data/ca-certs/rca-fa.pem
   	fabric-ca-client enroll -d -u https://rca-fa-admin:adminpw@rca-fa:7054
PEER_NAME=peer0-fa
PEER_PASS=adminpw
fabric-ca-client register -d --id.name $PEER_NAME --id.secret $PEER_PASS --id.type peer

	# The admin identity has the "admin" attribute which is added to ECert by default
	ADMIN_NAME=fa-admin
	ADMIN_PASS=adminpw
	USER_NAME=fa-user0
	USER_PASS=userpw
	fabric-ca-client register -d --id.name $ADMIN_NAME --id.secret $ADMIN_PASS --id.attrs "hf.Registrar.Roles=client,hf.Registrar.Attributes=*,hf.Revoker=true,hf.GenCRL=true,admin=true:ecert,abac.init=true:ecert"
	fabric-ca-client register -d --id.name $USER_NAME --id.secret $USER_PASS




	# Enroll CA Admin first
	export FABRIC_CA_CLIENT_HOME=$HOME/ca-admins/rca-la
   	export FABRIC_CA_CLIENT_TLS_CERTFILES=/data/ca-certs/rca-la.pem
   	fabric-ca-client enroll -d -u https://rca-la-admin:adminpw@rca-la:7054
PEER_NAME=peer0-la
PEER_PASS=adminpw
fabric-ca-client register -d --id.name $PEER_NAME --id.secret $PEER_PASS --id.type peer


	# The admin identity has the "admin" attribute which is added to ECert by default
	ADMIN_NAME=la-admin
	ADMIN_PASS=adminpw
	USER_NAME=la-user0
	USER_PASS=userpw
	fabric-ca-client register -d --id.name $ADMIN_NAME --id.secret $ADMIN_PASS --id.attrs "hf.Registrar.Roles=client,hf.Registrar.Attributes=*,hf.Revoker=true,hf.GenCRL=true,admin=true:ecert,abac.init=true:ecert"
	fabric-ca-client register -d --id.name $USER_NAME --id.secret $USER_PASS




# Generate msp folder structure in shared dir
	# Step
	export FABRIC_CA_CLIENT_HOME=$HOME/ca-admins/rca-orderer
   	export FABRIC_CA_CLIENT_TLS_CERTFILES=/data/ca-certs/rca-orderer.pem
	ORG_MSP_DIR=/data/orgmspdirs/orderer/msp
	fabric-ca-client getcacert -d -u https://rca-orderer:7054 -M $ORG_MSP_DIR
	mkdir $ORG_MSP_DIR/tlscacerts
	cp $ORG_MSP_DIR/cacerts/* $ORG_MSP_DIR/tlscacerts/

	# Populate admincerts dir
	ORG_ADMIN_HOME=/data/orgmspdirs/orderer/admin
	export FABRIC_CA_CLIENT_HOME=$ORG_ADMIN_HOME
      	export FABRIC_CA_CLIENT_TLS_CERTFILES=/data/ca-certs/rca-orderer.pem
	ADMIN_NAME=orderer-admin
	ADMIN_PASS=adminpw
	CA_HOST=rca-orderer
	
	# Copy admincerts to msp admincert folder
      	fabric-ca-client enroll -d -u https://$ADMIN_NAME:$ADMIN_PASS@$CA_HOST:7054
	mkdir -p $ORG_MSP_DIR/admincerts
	cp $ORG_ADMIN_HOME/msp/signcerts/* $ORG_MSP_DIR/admincerts/cert.pem
	# Copy it to admin's admincert folder
	mkdir $ORG_ADMIN_HOME/msp/admincerts
	cp $ORG_ADMIN_HOME/msp/signcerts/* $ORG_ADMIN_HOME/msp/admincerts/



	# Step
	export FABRIC_CA_CLIENT_HOME=$HOME/ca-admins/rca-bda
   	export FABRIC_CA_CLIENT_TLS_CERTFILES=/data/ca-certs/rca-bda.pem
	ORG_MSP_DIR=/data/orgmspdirs/bda/msp
	fabric-ca-client getcacert -d -u https://rca-bda:7054 -M $ORG_MSP_DIR
	mkdir $ORG_MSP_DIR/tlscacerts
	cp $ORG_MSP_DIR/cacerts/* $ORG_MSP_DIR/tlscacerts/

	# Populate admincerts dir
	ORG_ADMIN_HOME=/data/orgmspdirs/bda/admin
	export FABRIC_CA_CLIENT_HOME=$ORG_ADMIN_HOME
      	export FABRIC_CA_CLIENT_TLS_CERTFILES=/data/ca-certs/rca-bda.pem
	ADMIN_NAME=bda-admin
	ADMIN_PASS=adminpw
	CA_HOST=rca-bda
	
	# Copy admincerts to msp admincert folder
      	fabric-ca-client enroll -d -u https://$ADMIN_NAME:$ADMIN_PASS@$CA_HOST:7054
	mkdir -p $ORG_MSP_DIR/admincerts
	cp $ORG_ADMIN_HOME/msp/signcerts/* $ORG_MSP_DIR/admincerts/cert.pem
	# Copy it to admin's admincert folder
	mkdir $ORG_ADMIN_HOME/msp/admincerts
	cp $ORG_ADMIN_HOME/msp/signcerts/* $ORG_ADMIN_HOME/msp/admincerts/


	# Step
	export FABRIC_CA_CLIENT_HOME=$HOME/ca-admins/rca-fa
   	export FABRIC_CA_CLIENT_TLS_CERTFILES=/data/ca-certs/rca-fa.pem
	ORG_MSP_DIR=/data/orgmspdirs/fa/msp
	fabric-ca-client getcacert -d -u https://rca-fa:7054 -M $ORG_MSP_DIR
	mkdir $ORG_MSP_DIR/tlscacerts
	cp $ORG_MSP_DIR/cacerts/* $ORG_MSP_DIR/tlscacerts/

	# Populate admincerts dir
	ORG_ADMIN_HOME=/data/orgmspdirs/fa/admin
	export FABRIC_CA_CLIENT_HOME=$ORG_ADMIN_HOME
      	export FABRIC_CA_CLIENT_TLS_CERTFILES=/data/ca-certs/rca-fa.pem
	ADMIN_NAME=fa-admin
	ADMIN_PASS=adminpw
	CA_HOST=rca-fa
	
	# Copy admincerts to msp admincert folder
      	fabric-ca-client enroll -d -u https://$ADMIN_NAME:$ADMIN_PASS@$CA_HOST:7054
	mkdir -p $ORG_MSP_DIR/admincerts
	cp $ORG_ADMIN_HOME/msp/signcerts/* $ORG_MSP_DIR/admincerts/cert.pem
	# Copy it to admin's admincert folder
	mkdir $ORG_ADMIN_HOME/msp/admincerts
	cp $ORG_ADMIN_HOME/msp/signcerts/* $ORG_ADMIN_HOME/msp/admincerts/


	# Step
	export FABRIC_CA_CLIENT_HOME=$HOME/ca-admins/la-bda
   	export FABRIC_CA_CLIENT_TLS_CERTFILES=/data/ca-certs/rca-la.pem
	ORG_MSP_DIR=/data/orgmspdirs/la/msp
	fabric-ca-client getcacert -d -u https://rca-la:7054 -M $ORG_MSP_DIR
	mkdir $ORG_MSP_DIR/tlscacerts
	cp $ORG_MSP_DIR/cacerts/* $ORG_MSP_DIR/tlscacerts/

	# Populate admincerts dir
	ORG_ADMIN_HOME=/data/orgmspdirs/la/admin
	export FABRIC_CA_CLIENT_HOME=$ORG_ADMIN_HOME
      	export FABRIC_CA_CLIENT_TLS_CERTFILES=/data/ca-certs/rca-la.pem
	ADMIN_NAME=la-admin
	ADMIN_PASS=adminpw
	CA_HOST=rca-la
	
	# Copy admincerts to msp admincert folder
      	fabric-ca-client enroll -d -u https://$ADMIN_NAME:$ADMIN_PASS@$CA_HOST:7054
	mkdir -p $ORG_MSP_DIR/admincerts
	cp $ORG_ADMIN_HOME/msp/signcerts/* $ORG_MSP_DIR/admincerts/cert.pem
	# Copy it to admin's admincert folder
	mkdir $ORG_ADMIN_HOME/msp/admincerts
	cp $ORG_ADMIN_HOME/msp/signcerts/* $ORG_ADMIN_HOME/msp/admincerts/

# Make configtx.yml


# Generate Channel artifacts 
	
	cp /data/configtx.yaml $FABRIC_CFG_PATH/
	GENESIS_BLOCK_FILE=/data/genesis.block
	CHANNEL_TX_FILE=/data/channel.tx
	CHANNEL_NAME=mychannel
	generateChannelArtifacts

# Updating anchor peers

	# bda org
	ANCHOR_TX_FILE=/data/orgmspdirs/bda/anchors.tx
	ORG=bda
	generateAnchorPeers

	# fa org
	ANCHOR_TX_FILE=/data/orgmspdirs/fa/anchors.tx
	ORG=fa
	generateAnchorPeers

	# la org
	ANCHOR_TX_FILE=/data/orgmspdirs/la/anchors.tx
	ORG=la
	generateAnchorPeers
}

function generateChannelArtifacts() {
  which configtxgen
  if [ "$?" -ne 0 ]; then
    fatal "configtxgen tool not found. exiting"
  fi

  log "Generating orderer genesis block at $GENESIS_BLOCK_FILE"
  # Note: For some unknown reason (at least for now) the block file can't be
  # named orderer.genesis.block or the orderer will fail to launch!
  configtxgen -profile lanOrdererGenesis -outputBlock $GENESIS_BLOCK_FILE
  if [ "$?" -ne 0 ]; then
    fatal "Failed to generate orderer genesis block"
  fi

  log "Generating channel configuration transaction at $CHANNEL_TX_FILE"
  configtxgen -profile ThreeOrgsChannel -outputCreateChannelTx $CHANNEL_TX_FILE -channelID $CHANNEL_NAME
  if [ "$?" -ne 0 ]; then
    fatal "Failed to generate channel configuration transaction"
  fi
  
}

function generateAnchorPeers() {
	
	     
	     log "Generating anchor peer update transaction for $ORG at $ANCHOR_TX_FILE"
	     configtxgen -profile ThreeOrgsChannel -outputAnchorPeersUpdate $ANCHOR_TX_FILE \
		         -channelID $CHANNEL_NAME -asOrg $ORG
	     if [ "$?" -ne 0 ]; then
		fatal "Failed to generate anchor peer update for $ORG"
	     fi
	 
}















function makeConfigTxYaml {
   {
   echo "
################################################################################
#
#   Section: Organizations
#
#   - This section defines the different organizational identities which will
#   be referenced later in the configuration.
#
################################################################################
Organizations:"
   echo "
  - &orderer

    Name: orderer

    # ID to load the MSP definition as
    ID: ordererMSP

    # MSPDir is the filesystem path which contains the MSP configuration
    MSPDir: "

   

   echo "
################################################################################
#
#   SECTION: Application
#
#   This section defines the values to encode into a config transaction or
#   genesis block for application related parameters
#
################################################################################
Application: &ApplicationDefaults

    # Organizations is the list of orgs which are defined as participants on
    # the application side of the network
    Organizations:
"
   echo "
################################################################################
#
#   Profile
#
#   - Different configuration profiles may be encoded here to be specified
#   as parameters to the configtxgen tool
#
################################################################################
Profiles:

  OrgsOrdererGenesis:
    Orderer:
      # Orderer Type: The orderer implementation to start
      # Available types are \"solo\" and \"kafka\"
      OrdererType: solo
      Addresses:"

   for ORG in $ORDERER_ORGS; do
      local COUNT=1
      while [[ "$COUNT" -le $NUM_ORDERERS ]]; do
         initOrdererVars $ORG $COUNT
         echo "        - $ORDERER_HOST:7050"
         COUNT=$((COUNT+1))
      done
   done

   echo "
      # Batch Timeout: The amount of time to wait before creating a batch
      BatchTimeout: 2s

      # Batch Size: Controls the number of messages batched into a block
      BatchSize:

        # Max Message Count: The maximum number of messages to permit in a batch
        MaxMessageCount: 10

        # Absolute Max Bytes: The absolute maximum number of bytes allowed for
        # the serialized messages in a batch.
        AbsoluteMaxBytes: 99 MB

        # Preferred Max Bytes: The preferred maximum number of bytes allowed for
        # the serialized messages in a batch. A message larger than the preferred
        # max bytes will result in a batch larger than preferred max bytes.
        PreferredMaxBytes: 512 KB

      Kafka:
        # Brokers: A list of Kafka brokers to which the orderer connects
        # NOTE: Use IP:port notation
        Brokers:
          - 127.0.0.1:9092

      # Organizations is the list of orgs which are defined as participants on
      # the orderer side of the network
      Organizations:"

   for ORG in $ORDERER_ORGS; do
      initOrgVars $ORG
      echo "        - *${ORG_CONTAINER_NAME}"
   done

   echo "
    Consortiums:

      SampleConsortium:

        Organizations:"

   for ORG in $PEER_ORGS; do
      initOrgVars $ORG
      echo "          - *${ORG_CONTAINER_NAME}"
   done

   echo "
  OrgsChannel:
    Consortium: SampleConsortium
    Application:
      <<: *ApplicationDefaults
      Organizations:"

   for ORG in $PEER_ORGS; do
      initOrgVars $ORG
      echo "        - *${ORG_CONTAINER_NAME}"
   done

   } > /etc/hyperledger/fabric/configtx.yaml
   # Copy it to the data directory to make debugging easier
   cp /etc/hyperledger/fabric/configtx.yaml /data
}




main


