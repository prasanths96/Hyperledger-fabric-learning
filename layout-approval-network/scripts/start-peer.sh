function main() {
# Wait for other process to complete
sleep 10

# Although a peer may use the same TLS key and certificate file for both inbound and outbound TLS,
# we generate a different key and certificate for inbound and outbound TLS simply to show that it is permissible

# Generate server TLS cert and key pair for the peer
fabric-ca-client enroll -d --enrollment.profile tls -u $ENROLLMENT_URL -M /tmp/tls --csr.hosts $PEER_HOST

# Copy the TLS key and cert to the appropriate place
TLSDIR=$PEER_HOME/tls
mkdir -p $TLSDIR
cp /tmp/tls/signcerts/* $CORE_PEER_TLS_CERT_FILE
cp /tmp/tls/keystore/* $CORE_PEER_TLS_KEY_FILE
rm -rf /tmp/tls

# Generate client TLS cert and key pair for the peer
genClientTLSCert $PEER_NAME $CORE_PEER_TLS_CLIENTCERT_FILE $CORE_PEER_TLS_CLIENTKEY_FILE

# Generate client TLS cert and key pair for the peer CLI
genClientTLSCert $PEER_NAME /data/tls/$PEER_NAME-cli-client.crt /data/tls/$PEER_NAME-cli-client.key

# Enroll the peer to get an enrollment certificate and set up the core's local MSP directory
fabric-ca-client enroll -d -u $ENROLLMENT_URL -M $CORE_PEER_MSPCONFIGPATH

# Copy certs:
	
	# Common for copying tls and admincert
	ORG_MSP_DIR=/data/orgmspdirs/$ORG/msp

	# Copy tls certs from /data/.. folder to container local msp folder	
	mkdir $CORE_PEER_MSPCONFIGPATH/tlscacerts
	cp $ORG_MSP_DIR/tlscacerts/* $CORE_PEER_MSPCONFIGPATH/tlscacerts/

	# Copy admin certs like tls	
	mkdir $CORE_PEER_MSPCONFIGPATH/admincerts
	cp $ORG_MSP_DIR/admincerts/* $CORE_PEER_MSPCONFIGPATH/admincerts/


# Start the peer
log "Starting peer '$CORE_PEER_ID' with MSP at '$CORE_PEER_MSPCONFIGPATH'"
env | grep CORE
peer node start

}


function genClientTLSCert {
   if [ $# -ne 3 ]; then
      echo "Usage: genClientTLSCert <host name> <cert file> <key file>: $*"
      exit 1
   fi

   HOST_NAME=$1
   CERT_FILE=$2
   KEY_FILE=$3

   # Get a client cert
   fabric-ca-client enroll -d --enrollment.profile tls -u $ENROLLMENT_URL -M /tmp/tls --csr.hosts $HOST_NAME

   mkdir /data/tls || true
   cp /tmp/tls/signcerts/* $CERT_FILE
   cp /tmp/tls/keystore/* $KEY_FILE
   rm -rf /tmp/tls
}


main
