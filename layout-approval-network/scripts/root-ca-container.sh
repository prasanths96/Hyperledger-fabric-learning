FABRIC_CA_SERVER_HOME=/etc/hyperledger/fabric-ca
FABRIC_CA_SERVER_CSR_CN=rca-orderer
FABRIC_CA_SERVER_CSR_HOSTS=rca-orderer


# Initialize the root CA
fabric-ca-server init -b $BOOTSTRAP_USER_PASS
######################### ^^ Careful about this variable, setup-fabric script isnt generic

# Copy ca cert into shared dir
cp $FABRIC_CA_SERVER_HOME/ca-cert.pem /data/ca-certs/rca-orderer.pem

# Add custom orgs:

aff="orderer: []\n   bda: []\n   fa: []\n   la: []" 

aff="${aff#\\n   }"

sed -i "/affiliations:/a \\   $aff" \
   $FABRIC_CA_SERVER_HOME/fabric-ca-server-config.yaml


# Start the root CA
fabric-ca-server start

