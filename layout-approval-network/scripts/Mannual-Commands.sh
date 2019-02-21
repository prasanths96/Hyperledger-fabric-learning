## Layout approval network example invocation flow:
# Register a BDA user: (arg1: bda or fa or la, arg2: username, arg3: password):
# This can also be used to switch to that user. "(dot)<space>" is very important while using this script to export variables to current shell:
. registerUser bda user1 userpw

# Invoke chaincode by createLayout using bda (Only bda can create layout):
# The variables used in the commands are already exported when using "registerUser" script:
peer chaincode invoke -C $CHANNEL_NAME -n mycc -c '{"Args":["createLayout","1","address"]}' $ORDERER_CONN_ARGS

# Invoke by requestNOC as bda (Only BDA can access):
peer chaincode invoke -C $CHANNEL_NAME -n mycc -c '{"Args":["requestNOC","1"]}' $ORDERER_CONN_ARGS

# Register a FA user: 
. registerUser fa user1 userpw 

# Invoke approveLayout or rejectLayout:
peer chaincode invoke -C $CHANNEL_NAME -n mycc -c '{"Args":["approveLayout","1"]}' $ORDERER_CONN_ARGS
#or
#peer chaincode invoke -C $CHANNEL_NAME -n mycc -c '{"Args":["rejectLayout","1"]}' $ORDERER_CONN_ARGS

# Register a LA user:
. registerUser la user1 userpw

# Invoke approveLayout or rejectLayout: 
peer chaincode invoke -C $CHANNEL_NAME -n mycc -c '{"Args":["approveLayout","1"]}' $ORDERER_CONN_ARGS
#or
#peer chaincode invoke -C $CHANNEL_NAME -n mycc -c '{"Args":["rejectLayout","1"]}' $ORDERER_CONN_ARGS

# Query: 
peer chaincode query -C $CHANNEL_NAME -n mycc -c '{"Args":["viewLayout","1"]}'

# Query Layout history: (Only provided layoutId's history)
peer chaincode query -C $CHANNEL_NAME -n mycc -c '{"Args":["getHistory","1"]}'

# Query all transactions history: (All layout's history) 
peer chaincode query -C $CHANNEL_NAME -n mycc -c '{"Args":["getHistory","ALL_TRANSACTION_HISTORY"]}'

## Example flow to demonstrate the ability of "All transaction history":
	
# Same process as above but, with new layout with ID "2"	
. registerUser bda user1 userpw	
peer chaincode invoke -C $CHANNEL_NAME -n mycc -c '{"Args":["createLayout","2","address"]}' $ORDERER_CONN_ARGS
peer chaincode invoke -C $CHANNEL_NAME -n mycc -c '{"Args":["requestNOC","2"]}' $ORDERER_CONN_ARGS
. registerUser fa user1 userpw 	
peer chaincode invoke -C $CHANNEL_NAME -n mycc -c '{"Args":["approveLayout","2"]}' $ORDERER_CONN_ARGS
. registerUser la user1 userpw
peer chaincode invoke -C $CHANNEL_NAME -n mycc -c '{"Args":["approveLayout","2"]}' $ORDERER_CONN_ARGS
	
# Query Layout history: (Only provided layoutId's history)
peer chaincode query -C $CHANNEL_NAME -n mycc -c '{"Args":["getHistory","2"]}'

# Query all transactions history: (All layout's history) 
peer chaincode query -C $CHANNEL_NAME -n mycc -c '{"Args":["getHistory","ALL_TRANSACTION_HISTORY"]}'


## Example invocation to demonstrate BCCSP encrypt and store data:
 
# Switch back to bda identity to be able to create layout:
. registerUser bda user1 userpw

# Create an Encryption key: (Here we use Encryption key as Decryption key aswell. (Symmetric key cryptography))
ENCKEY=`openssl rand 32 -base64` && DECKEY=$ENCKEY

# Invoke encryptAndCreateLayout: (Encryption key passed in Transient field of command makes sure key is not stored in database)
peer chaincode invoke -C $CHANNEL_NAME -n mycc -c '{"Args":["encryptAndCreateLayout","999","address"]}' --transient "{\"ENCKEY\":\"$ENCKEY\"}" $ORDERER_CONN_ARGS

# Checking if the data is really encrypted by querying normally:
peer chaincode query -C $CHANNEL_NAME -n mycc -c '{"Args":["viewLayout","999"]}'

# Now querying using decryptAndViewLayout: 
peer chaincode query -C $CHANNEL_NAME -n mycc -c '{"Args":["decryptAndViewLayout","999"]}' --transient "{\"DECKEY\":\"$DECKEY\"}"


	
