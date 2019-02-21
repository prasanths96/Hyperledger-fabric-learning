# Wait for other process to complete
sleep 15

# Enroll to get orderer's TLS cert (using the "tls" profile)
fabric-ca-client enroll -d --enrollment.profile tls -u $ENROLLMENT_URL -M /tmp/tls --csr.hosts $ORDERER_HOST

# Copy the TLS key and cert to the appropriate place
TLSDIR=$ORDERER_HOME/tls
mkdir -p $TLSDIR
cp /tmp/tls/keystore/* $ORDERER_GENERAL_TLS_PRIVATEKEY
cp /tmp/tls/signcerts/* $ORDERER_GENERAL_TLS_CERTIFICATE
rm -rf /tmp/tls

# Enroll again to get the orderer's enrollment certificate (default profile)
fabric-ca-client enroll -d -u $ENROLLMENT_URL -M $ORDERER_GENERAL_LOCALMSPDIR

# Copy certs:
	
	# Common for copying tls and admincert
	ORG_MSP_DIR=/data/orgmspdirs/orderer/msp

	# Copy tls certs from /data/.. folder to container local msp folder	
	mkdir $ORDERER_GENERAL_LOCALMSPDIR/tlscacerts
	cp $ORG_MSP_DIR/tlscacerts/* $ORDERER_GENERAL_LOCALMSPDIR/tlscacerts/

	# Copy admin certs like tls	
	mkdir $ORDERER_GENERAL_LOCALMSPDIR/admincerts
	cp $ORG_MSP_DIR/admincerts/* $ORDERER_GENERAL_LOCALMSPDIR/admincerts/

# Wait for the genesis block to be created
sleep 15

# Start the orderer
env | grep ORDERER
orderer
