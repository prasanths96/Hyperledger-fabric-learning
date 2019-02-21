/*=======================================================================
Project name: Layout Approval Network
Developer: Prasanth Sundaravelu
Purpose: POC for Certification
Special features implemented: Attribute Based Access Control (ABAC),
							  Data Encryption Before storage using BCCSP,
							  Historian for all transactions (Similar to Hyperledger composer)
Abbrevations: BDA - Bangalore Development Authority,
			  FA - Forest Authority
			  LA - Lake Authority
			  lan - Layout Approval Network
Sample transactions:
	Invocation:
		(As BDA only: Identity certificate must hold attribute lan.role="bda")
		peer chaincode invoke -C $CHANNEL_NAME -n $CHAINCODE_NAME -c '{"Args":["createLayout","1","address"]}' $ORDERER_CONN_ARGS
		peer chaincode invoke -C $CHANNEL_NAME -n $CHAINCODE_NAME -c '{"Args":["requestNOC","1"]}' $ORDERER_CONN_ARGS

		ENCKEY=`openssl rand 32 -base64` && DECKEY=$ENCKEY
		peer chaincode invoke -C $CHANNEL_NAME -n $CHAINCODE_NAME -c '{"Args":["encryptAndCreateLayout","2","address"]}' --transient "{\"ENCKEY\":\"$ENCKEY\"}" $ORDERER_CONN_ARGS

		(As FA or LA only: Identity certificate must hold attribute lan.role="fa" or "la")
		peer chaincode invoke -C $CHANNEL_NAME -n $CHAINCODE_NAME -c '{"Args":["approveLayout","1"]}' $ORDERER_CONN_ARGS
		peer chaincode invoke -C $CHANNEL_NAME -n $CHAINCODE_NAME -c '{"Args":["rejectLayout","1"]}' $ORDERER_CONN_ARGS

	Query:
		peer chaincode query -C $CHANNEL_NAME -n $CHAINCODE_NAME -c '{"Args":["viewLayout","1"]}'
		peer chaincode query -C $CHANNEL_NAME -n $CHAINCODE_NAME -c '{"Args":["decryptAndViewLayout","2"]}' --transient "{\"DECKEY\":\"$DECKEY\"}"
		(Getting history of transaction based on layout ID)
		peer chaincode query -C $CHANNEL_NAME -n mycc -c '{"Args":["getHistory","1"]}'
		(Getting history of ALL transaction)
		peer chaincode query -C $CHANNEL_NAME -n mycc -c '{"Args":["getHistory","ALL_TRANSACTION_HISTORY"]}'
=======================================================================*/

package main

import (
	"encoding/json"
	"fmt"
	"time"

	//"strconv"
	"strings"

	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/bccsp/factory"
	"github.com/hyperledger/fabric/core/chaincode/lib/cid"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/core/chaincode/shim/ext/entities"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
)

// Define Status codes for the response
const (
	OK    = 200
	ERROR = 500
)

// Constants for Encryption
const DECKEY = "DECKEY"
const ENCKEY = "ENCKEY"
const IV = "IV"

const allTransactionKey = "ALL_TRANSACTION_HISTORY"

// Define the Smart Contract structure
type lacc struct {
	bccspInst bccsp.BCCSP
}

type layout struct {
	AssetType      string `json:"AssetType"`
	Id             string `json:"Id"`
	Address        string `json:"Address"`
	RequestedNOC   bool   `json:"requestedNOC"`
	LAStatus       string `json:"LAStatus"`
	FAStatus       string `json:"FAStatus"`
	ApprovalStatus string `json:"ApprovalStatus"`
}

type layoutHT struct {
	Id             string `json:"Id"`
	Address        string `json:"Address"`
	RequestedNOC   string `json:"requestedNOC"`
	FAStatus       string `json:"FAStatus"`
	LAStatus       string `json:"LAStatus"`
	ApprovalStatus string `json:"ApprovalStatus"`
}

type layoutHistory struct {
	TxId           string `json: "TxId"`
	Id             string `json:"Id"`
	Address        string `json:"Address"`
	RequestedNOC   bool   `json:"requestedNOC"`
	LAStatus       string `json:"LAStatus"`
	FAStatus       string `json:"FAStatus"`
	ApprovalStatus string `json:"ApprovalStatus"`
	TimeStamp      string `json:"TimeStamp"`
	IsDelete       bool   `json:"IsDelete"`
}

// Creating comparison functions for time variables
type timeVariables struct {
	seconds int64
	nanos   int32
}

/* Only required for using couchdb
// Creating objects for seperate fields
type faObj struct {
	Id           string `json:"Id"`
	RequestedNOC string `json:"requestedNOC"`
	FAStatus     string `json:"FAStatus"`
}

type laObj struct {
	Id           string `json:"Id"`
	RequestedNOC string `json:"requestedNOC"`
	LAStatus     string `json:"LAStatus"`
} */

// Creating generic object for internal use
type fieldObj struct {
	Id           string
	RequestedNOC string
	fieldVal     string
}

// Returns 1 if t1 is bigger/newer, -1 if t2 is bigger/newer, 0 if same
func (t1 *timeVariables) Compare(t2 *timeVariables) int32 {
	if t1.seconds > t2.seconds {
		return 1
	} else if t1.seconds < t2.seconds {
		return -1
	} else { // if seconds are same, check nano seconds
		if t1.nanos > t2.nanos {
			return 1
		} else if t1.nanos < t2.nanos {
			return -1
		} else { // if seconds and nanos are same
			return 0
		}
	}
}

// RangeQuery for FA or LA for solving PHANTOM_READ_CONFLICT problem
func seperateRangeQuery(stub shim.ChaincodeStubInterface, args []string) (*fieldObj, error) {
	var err error
	errObj := &fieldObj{}

	// ==== 1 =====   2
	// ==== Id ==== Field  {FA or LA}
	layoutId := args[0]
	field := strings.ToLower(args[1])
	// Set composite index name based on field
	compositeIndexName := "id~noc~" + field

	// ==== Get composite keys with same ID ====
	deltaResultsIterator, err := stub.GetStateByPartialCompositeKey(compositeIndexName, []string{layoutId})
	if err != nil {
		return errObj, err
	}
	defer deltaResultsIterator.Close()

	// Check if layout does not exist
	if !deltaResultsIterator.HasNext() {
		return errObj, errors.New(fmt.Sprintf("Layout: %s does not exist", layoutId))
	}

	// Creating variables for layoutHT object
	var noc, fieldVal string
	// Creating time variables for comparison for 3 fields
	oldTime := &timeVariables{}

	for i := 0; deltaResultsIterator.HasNext(); i++ {
		// Get the next state
		state, nextErr := deltaResultsIterator.Next()
		if nextErr != nil {
			return errObj, nextErr
		}
		// Split the composite key into its component parts
		_, keyParts, splitKeyErr := stub.SplitCompositeKey(state.Key)
		if splitKeyErr != nil {
			return errObj, splitKeyErr
		}
		// if not null, change value
		// ========= !! ASSUMING that GetStateByPartialCompositeKey returns chronological order  <<<<< FALSE!!! =========
		// ======== GetStateByPartialCompositeKey returns lexical order =======

		// Taking timestamp in seconds and nano seconds for comparison
		resultsIterator, err := stub.GetHistoryForKey(state.Key)
		if err != nil {
			return errObj, err
		}
		defer resultsIterator.Close()

		response, err := resultsIterator.Next()
		if err != nil {
			return errObj, err
		}
		// Current key's time variables
		seconds := response.Timestamp.Seconds
		nanos := response.Timestamp.Nanos
		timeStamp := &timeVariables{seconds, nanos}

		// These variables do not differ
		// id = keyParts[0]

		// Flags to denote if the "if" conditions has passed
		var ifFlag bool

		if keyParts[1] != "" && timeStamp.Compare(oldTime) == 1 {
			ifFlag = true
			noc = keyParts[1]
			oldTime = timeStamp
		}

		if keyParts[2] != "" && timeStamp.Compare(oldTime) == 1 {
			ifFlag = true
			fieldVal = keyParts[2]
			oldTime = timeStamp
		}

		// Logic when multiple transactions occurred during same time
		// Just displaying all the changes happened at same time - (UNSOLVED)
		if !ifFlag && keyParts[1] != "" && timeStamp.Compare(oldTime) == 0 {
			noc = noc + keyParts[1]
		}
		if !ifFlag && keyParts[2] != "" && timeStamp.Compare(oldTime) == 0 {
			fieldVal = fieldVal + keyParts[2]
		}

	}

	// Checking if variable is empty
	if fieldVal == "" {
		fieldVal = "NA"
	}
	if noc == "" {
		noc = "false"
	}

	// Creating object to send
	fieldObject := &fieldObj{layoutId, noc, fieldVal}
	return fieldObject, nil

}

func (t *lacc) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (t *lacc) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()

	tMap, err := stub.GetTransient()
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve transient, err %s", err))
	}

	fmt.Println("Invoke is running " + function)

	// Handle different functions
	if function == "requestNOC" { // Requests NOC only by BDA
		return t.requestNOC(stub, args)
	} else if function == "createLayout" { // only by BDA
		return t.createLayout(stub, args)
	} else if function == "viewLayout" {
		return t.viewLayout(stub, args)
	} else if function == "encryptAndCreateLayout" {
		return t.Encrypter(stub, args, tMap[ENCKEY], tMap[IV])
	} else if function == "decryptAndViewLayout" {
		return t.Decrypter(stub, args, tMap[DECKEY], tMap[IV])
	} else if function == "approveLayout" { // FA and LA can do this
		return t.approveLayout(stub, args)
	} else if function == "rejectLayout" { // FA and LA can do this
		return t.rejectLayout(stub, args)
	} else if function == "getHistory" {
		return t.getHistory(stub, args)
	} else if function == "createLayoutHT" {
		return t.createLayoutHT(stub, args) // only by BDA
	} else if function == "approveLayoutHT" {
		return t.approveLayoutHT(stub, args) // only by FA or LA
	} else if function == "rejectLayoutHT" {
		return t.rejectLayoutHT(stub, args) // only by FA or LA
	} else if function == "viewLayoutHT" {
		return t.viewLayoutHT(stub, args)
	} else if function == "requestNOCHT" {
		return t.requestNOCHT(stub, args) // only by BDA
	}

	fmt.Println("Invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

func (t *lacc) createLayout(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	//   0       1
	// "Id", "Address"
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2: Id and Address.")
	}

	// Check if the invoker is authorized
	// MSPID, err := cid.GetMSPID(stub)
	// fmt.Println(MSPID);

	//GetID returns the ID associated with the invoking identity.
	//This ID is guaranteed to be unique within the MSP.
	//ID, err := cid.GetID(stub)
	//fmt.Println("GetId: "+ID+"\n X509:\n")

	//creator, err := stub.GetCreator()
	//fmt.Println("Creator:")
	//fmt.Println(string(creator))

	// Check if the invoker's role is bda : lan.role=bda
	val, ok, err := cid.GetAttributeValue(stub, "lan.role")
	if ok == false {
		fmt.Println("Attribute id not found")
		return shim.Error("only lan.role = 'bda' attribute users can create layout.")
	}
	if err != nil {
		return shim.Error(err.Error())
	}
	if val != "bda" {
		fmt.Println("Attribute role: " + val)
		return shim.Error("only lan.role = 'bda' attribute users can create layout.")
	}
	fmt.Println("Attribute value:" + val)

	//func cid.GetAttributeValue(stub ChaincodeStubInterface, attrName string) (value string, found bool, err error)

	// ==== Input sanitation ====
	fmt.Println("- Layout creation started")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}

	layoutId := args[0]
	Address := strings.ToLower(args[1])

	// ==== Check if layout already exists ====
	layoutAsBytes, err := stub.GetState(layoutId)
	if err != nil {
		return shim.Error("Failed to get Layout: " + err.Error())
	} else if layoutAsBytes != nil {
		fmt.Println("This layout already exists: " + layoutId)
		return shim.Error("This layout already exists: " + layoutId)
	}

	// ==== Create layout object and marshal to JSON ====
	AssetType := "layout"
	RequestedNOC := false
	LAStatus := "NA"
	FAStatus := "NA"
	ApprovalStatus := "NA"

	layout := &layout{AssetType, layoutId, Address, RequestedNOC, LAStatus, FAStatus, ApprovalStatus}
	layoutJSONAsBytes, err := json.Marshal(layout)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save layout to state ===
	err = stub.PutState(layoutId, layoutJSONAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// ==== Also updating in all transactions world state ====
	err = stub.PutState(allTransactionKey, layoutJSONAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	// ==== Publish event ====
	eventPayload := "Created Layout: " + layoutId
	payloadAsBytes := []byte(eventPayload)
	err = stub.SetEvent("mainEvent", payloadAsBytes)

	// ==== Layout Saved . Return success ====
	fmt.Println("- End create layout")
	return shim.Success(nil)
}

// ======== HT Version ========

func (t *lacc) createLayoutHT(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	//       0         1
	// ==== id ==== address

	if len(args) != 2 {
		return shim.Error("Incorrect no. of arguments. Expected: 2 - {id, address}")
	}

	// ==== ABAC ====
	val, ok, err := cid.GetAttributeValue(stub, "lan.role")
	if ok == false {
		fmt.Println("Attribute id not found")
		return shim.Error("only lan.role = 'bda' attribute users can create layout.")
	}
	if err != nil {
		return shim.Error(err.Error())
	}
	if val != "bda" {
		fmt.Println("Attribute role: " + val)
		return shim.Error("only lan.role = 'bda' attribute users can create layout.")
	}

	// ==== Input sanitation ====
	fmt.Println("- Layout creation started")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}

	layoutId := args[0]
	Address := strings.ToLower(args[1])
	// Retrieve info needed for the update procedure
	txid := stub.GetTxID()
	compositeIndexName := "id~address~noc~fa~la~txID"

	// ==== Check if Layout already exist ====
	deltaResultsIterator, err := stub.GetStateByPartialCompositeKey(compositeIndexName, []string{layoutId})
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve value for %s: %s", layoutId, err.Error()))
	}
	defer deltaResultsIterator.Close()

	// Check if layout already exists
	if deltaResultsIterator.HasNext() {
		return shim.Error(fmt.Sprintf("Layout: %s already exists", layoutId))
	}

	// ==== Put to state process begin ==
	// Create the composite key that will allow us to query for all deltas on a particular variable
	compositeKey, compositeErr := stub.CreateCompositeKey(compositeIndexName, []string{layoutId, Address, "", "", "", txid})
	if compositeErr != nil {
		return shim.Error(fmt.Sprintf("Could not create a composite key for %s: %s", layoutId, compositeErr.Error()))
	}
	// faObj
	compositeIndexName = "id~noc~fa"
	compositeKeyfa, compositeErr := stub.CreateCompositeKey(compositeIndexName, []string{layoutId, "", ""})
	if compositeErr != nil {
		return shim.Error(fmt.Sprintf("Could not create a composite key for %s: %s", layoutId, compositeErr.Error()))
	}
	// laObj
	compositeIndexName = "id~noc~la"
	compositeKeyla, compositeErr := stub.CreateCompositeKey(compositeIndexName, []string{layoutId, "", ""})
	if compositeErr != nil {
		return shim.Error(fmt.Sprintf("Could not create a composite key for %s: %s", layoutId, compositeErr.Error()))
	}

	// Save the composite key index
	compositePutErr := stub.PutState(compositeKey, []byte{0x00})
	if compositePutErr != nil {
		return shim.Error(fmt.Sprintf("Could not put operation for %s in the ledger: %s", layoutId, compositePutErr.Error()))
	}
	// fa
	compositePutErr = stub.PutState(compositeKeyfa, []byte{0x00})
	if compositePutErr != nil {
		return shim.Error(fmt.Sprintf("Could not put operation for %s in the ledger: %s", layoutId, compositePutErr.Error()))
	}
	// la
	compositePutErr = stub.PutState(compositeKeyla, []byte{0x00})
	if compositePutErr != nil {
		return shim.Error(fmt.Sprintf("Could not put operation for %s in the ledger: %s", layoutId, compositePutErr.Error()))
	}

	// ==== Publish event ====
	eventPayload := "Created Layout: " + layoutId
	payloadAsBytes := []byte(eventPayload)
	err = stub.SetEvent("mainEvent", payloadAsBytes)

	return shim.Success([]byte(fmt.Sprintf("Successfully added layout: %s", layoutId)))

}

func (t *lacc) viewLayout(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	// 1
	// layoutId

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expected: 1 - LayoutId ")
	}

	fmt.Println("Process started..")

	layoutId := args[0]

	layoutAsBytes, err := stub.GetState(layoutId)
	if err != nil {
		//jsonResp = "{\"Error\":\"Failed to get state for " + layoutId + "\"}"
		return shim.Error(err.Error())
	} else if layoutAsBytes == nil {
		//jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return shim.Error("The layout with Id: " + layoutId + " does not exist.")
	}

	return shim.Success(layoutAsBytes)

}

// ======== HT Version ========

func (t *lacc) viewLayoutHT(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	// ==== 1
	// ==== ID

	if len(args) != 1 {
		return shim.Error("Incorrect no. of arguments. Expecting: 1 - {Id}")
	}

	layoutId := args[0]
	compositeIndexName := "id~address~noc~fa~la~txID"
	// ==== Get composite keys with same ID ====
	deltaResultsIterator, err := stub.GetStateByPartialCompositeKey(compositeIndexName, []string{layoutId})
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve value for %s: %s", layoutId, err.Error()))
	}
	defer deltaResultsIterator.Close()

	// Check if layout does not exist
	if !deltaResultsIterator.HasNext() {
		return shim.Error(fmt.Sprintf("Layout: %s does not exists", layoutId))
	}

	// Creating variables for layoutHT object
	var id, address, noc, fa, la, approvalStatus string
	// Creating time variables for comparison for 3 fields
	oldTimeNOC := &timeVariables{}
	oldTimeFA := &timeVariables{}
	oldTimeLA := &timeVariables{}

	for i := 0; deltaResultsIterator.HasNext(); i++ {
		// Get the next state
		state, nextErr := deltaResultsIterator.Next()
		if nextErr != nil {
			return shim.Error(nextErr.Error())
		}
		// Split the composite key into its component parts
		_, keyParts, splitKeyErr := stub.SplitCompositeKey(state.Key)
		if splitKeyErr != nil {
			return shim.Error(splitKeyErr.Error())
		}
		// if not null, change value
		// ========= !! ASSUMING that GetStateByPartialCompositeKey returns chronological order  <<<<< FALSE!!! =========
		// ======== GetStateByPartialCompositeKey returns lexical order =======

		// Taking timestamp in seconds and nano seconds for comparison
		resultsIterator, err := stub.GetHistoryForKey(state.Key)
		if err != nil {
			return shim.Error(err.Error())
		}
		defer resultsIterator.Close()

		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Current key's time variables
		seconds := response.Timestamp.Seconds
		nanos := response.Timestamp.Nanos
		timeStamp := &timeVariables{seconds, nanos}

		// These variables do not differ
		id = keyParts[0]
		address = keyParts[1]

		// Flags to denote if the "if" conditions has passed
		var nocFlag, faFlag, laFlag bool

		if keyParts[2] != "" && timeStamp.Compare(oldTimeNOC) == 1 {
			nocFlag = true
			noc = keyParts[2]
			oldTimeNOC = timeStamp
		}
		if keyParts[3] != "" && timeStamp.Compare(oldTimeFA) == 1 {
			faFlag = true
			fa = keyParts[3]
			oldTimeFA = timeStamp
		}
		if keyParts[4] != "" && timeStamp.Compare(oldTimeLA) == 1 {
			laFlag = true
			la = keyParts[4]
			oldTimeLA = timeStamp
		}

		// Logic when multiple transactions occurred during same time
		// Just displaying all the changes happened at same time - (UNSOLVED)
		if !nocFlag && keyParts[2] != "" && timeStamp.Compare(oldTimeNOC) == 0 {
			noc = noc + keyParts[2]
		}
		if !faFlag && keyParts[3] != "" && timeStamp.Compare(oldTimeFA) == 0 {
			fa = fa + keyParts[3]
		}
		if !laFlag && keyParts[4] != "" && timeStamp.Compare(oldTimeLA) == 0 {
			la = la + keyParts[4]
		}

	}
	// Setting Total approval status
	if fa == "APPROVED" && la == "APPROVED" {
		approvalStatus = "APPROVED"
	} else if fa == "REJECTED" || la == "REJECTED" {
		approvalStatus = "REJECTED"
	} else {
		approvalStatus = "NA"
	}

	// Checking if variables are empty
	if noc == "" {
		noc = "false"
	}
	if fa == "" {
		fa = "NA"
	}
	if la == "" {
		la = "NA"
	}

	// Create an object of type layoutHT
	layoutHT := &layoutHT{id, address, noc, fa, la, approvalStatus}
	layoutHTAsBytes, err := json.Marshal(layoutHT)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(layoutHTAsBytes)

}

func (t *lacc) requestNOC(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	// Check if the invoker's role is bda : lan.role=bda
	val, ok, err := cid.GetAttributeValue(stub, "lan.role")
	if ok == false {
		fmt.Println("Attribute id not found")
		return shim.Error("only lan.role = 'bda' attribute users can create layout.")
	}
	if err != nil {
		return shim.Error(err.Error())
	}
	if val != "bda" {
		fmt.Println("Attribute role: " + val)
		return shim.Error("only lan.role = 'bda' attribute users can create layout.")
	}

	// 1
	// Id
	if len(args) != 1 {
		return shim.Error("Expecting one arg: ID")
	}

	layoutId := args[0]
	layoutAsBytes, err := stub.GetState(layoutId)
	if err != nil {
		return shim.Error(err.Error())
	} else if layoutAsBytes == nil {
		return shim.Error("The layout with Id: " + layoutId + " does not exist.")
	}
	updatedLayout := layout{}
	err = json.Unmarshal(layoutAsBytes, &updatedLayout) //unmarshal it aka JSON.parse()
	if err != nil {
		return shim.Error(err.Error())
	}

	// Change Requested NOC status to true after checking
	if updatedLayout.RequestedNOC == true {
		return shim.Error("NOC already requested for layout: " + layoutId)
	}

	updatedLayout.RequestedNOC = true

	updatedLayoutAsBytes, _ := json.Marshal(updatedLayout)
	err = stub.PutState(layoutId, updatedLayoutAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Also updating in all transactions world state
	err = stub.PutState(allTransactionKey, updatedLayoutAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Publish event
	eventPayload := "NOC Requested for layout: " + layoutId
	payloadAsBytes := []byte(eventPayload)
	err = stub.SetEvent("mainEvent", payloadAsBytes)

	// FINALLY CREATE EVENT TO SEND TO FA AND LA
	return shim.Success(nil)
}

// ======== HT Version =========
func (t *lacc) requestNOCHT(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	// ==== 1
	// ==== ID

	if len(args) != 1 {
		return shim.Error("Incorrect no. of arguments. Expecting: 1 - {Id}")
	}

	// Check if the invoker's role is bda : lan.role=bda
	val, ok, err := cid.GetAttributeValue(stub, "lan.role")
	if ok == false {
		fmt.Println("Attribute id not found")
		return shim.Error("only lan.role = 'bda' attribute users can create layout.")
	}
	if err != nil {
		return shim.Error(err.Error())
	}
	if val != "bda" {
		fmt.Println("Attribute role: " + val)
		return shim.Error("only lan.role = 'bda' attribute users can create layout.")
	}

	// Check if NOC already requested.
	resultAsResponse := t.viewLayoutHT(stub, args)
	if resultAsResponse.Status == ERROR {
		return shim.Error(string(resultAsResponse.Payload)) // Check this
	}
	resultAsBytes := []byte(resultAsResponse.Payload)
	// Unmarshall JSON data stored as Bytes to Object
	result := layoutHT{}
	err = json.Unmarshal(resultAsBytes, &result)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Checking if layout is requested for NOC
	if result.RequestedNOC == "true" {
		return shim.Error("Layout is already requested for NOC")
	}

	// Begin process of putState
	layoutId := args[0]
	txid := stub.GetTxID()
	address := ""
	noc := "true"
	fa := ""
	la := ""

	compositeIndexName := "id~address~noc~fa~la~txID"

	compositeKey, compositeErr := stub.CreateCompositeKey(compositeIndexName, []string{layoutId, address, noc, fa, la, txid})
	if compositeErr != nil {
		return shim.Error(fmt.Sprintf("Could not create a composite key for %s: %s", layoutId, compositeErr.Error()))
	}
	// faObj
	compositeIndexName = "id~noc~fa"
	compositeKeyfa, compositeErr := stub.CreateCompositeKey(compositeIndexName, []string{layoutId, noc, ""})
	if compositeErr != nil {
		return shim.Error(fmt.Sprintf("Could not create a composite key for %s: %s", layoutId, compositeErr.Error()))
	}
	// laObj
	compositeIndexName = "id~noc~la"
	compositeKeyla, compositeErr := stub.CreateCompositeKey(compositeIndexName, []string{layoutId, noc, ""})
	if compositeErr != nil {
		return shim.Error(fmt.Sprintf("Could not create a composite key for %s: %s", layoutId, compositeErr.Error()))
	}

	// Save the composite key index
	compositePutErr := stub.PutState(compositeKey, []byte{0x00})
	if compositePutErr != nil {
		return shim.Error(fmt.Sprintf("Could not put operation for %s in the ledger: %s", layoutId, compositePutErr.Error()))
	}
	// fa
	compositePutErr = stub.PutState(compositeKeyfa, []byte{0x00})
	if compositePutErr != nil {
		return shim.Error(fmt.Sprintf("Could not put operation for %s in the ledger: %s", layoutId, compositePutErr.Error()))
	}
	// la
	compositePutErr = stub.PutState(compositeKeyla, []byte{0x00})
	if compositePutErr != nil {
		return shim.Error(fmt.Sprintf("Could not put operation for %s in the ledger: %s", layoutId, compositePutErr.Error()))
	}
	// Publish event
	eventPayload := "NOC Requested for layout: " + layoutId
	payloadAsBytes := []byte(eventPayload)
	err = stub.SetEvent("mainEvent", payloadAsBytes)

	return shim.Success([]byte("NOC Requested successfully."))

}

func (t *lacc) approveLayout(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	// Check if the invoker's role is bda : lan.role=bda
	attrVal, ok, err := cid.GetAttributeValue(stub, "lan.role")
	if ok == false {
		fmt.Println("Attribute id not found")
		return shim.Error("only lan.role = 'fa' or 'la' attribute users can approve layout.")
	}
	if err != nil {
		return shim.Error(err.Error())
	}
	if attrVal != "fa" && attrVal != "la" {
		fmt.Println("Attribute role: " + attrVal)
		return shim.Error("only lan.role = 'fa' or 'la' attribute users can approve layout.")
	}

	// 1
	// ID
	if len(args) != 1 {
		return shim.Error("Expecting one arg: ID")
	}

	// Get state from worldstate
	layoutId := args[0]
	layoutAsBytes, err := stub.GetState(layoutId)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Unmarshall JSON data stored as Bytes to Object
	updatedLayout := layout{}
	err = json.Unmarshal(layoutAsBytes, &updatedLayout)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Checking if layout is requested for NOC
	if updatedLayout.RequestedNOC == false {
		return shim.Error("Layout is not requested for NOC")
	}

	// Changing FAStatus or LAStatus to APPROVED respectively for the role invoked
	if attrVal == "fa" {
		if updatedLayout.FAStatus == "APPROVED" {
			return shim.Error("Layout is already approved by FA")
		}
		updatedLayout.FAStatus = "APPROVED"
	} else if attrVal == "la" {
		if updatedLayout.LAStatus == "APPROVED" {
			return shim.Error("Layout is already approved by LA")
		}
		updatedLayout.LAStatus = "APPROVED"
	}

	// Changing total ApprovalStatus respectively.
	if updatedLayout.FAStatus == "APPROVED" && updatedLayout.LAStatus == "APPROVED" {
		updatedLayout.ApprovalStatus = "APPROVED"
	} else if updatedLayout.FAStatus == "REJECTED" || updatedLayout.LAStatus == "REJECTED" {
		updatedLayout.ApprovalStatus = "REJECTED"
	} else {
		updatedLayout.ApprovalStatus = "NA"
	}

	// Marshaling back to json and storing in worldstate
	updatedLayoutAsBytes, _ := json.Marshal(updatedLayout)
	err = stub.PutState(layoutId, updatedLayoutAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Also updating in all transactions world state
	err = stub.PutState(allTransactionKey, updatedLayoutAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Publish event
	eventPayload := "Approved Layout: " + layoutId + " by " + attrVal
	payloadAsBytes := []byte(eventPayload)
	err = stub.SetEvent("mainEvent", payloadAsBytes)

	return shim.Success(nil)

}

// ======== HT Version ========

func (t *lacc) approveLayoutHT(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	// ====  1
	// ==== id

	if len(args) != 1 {
		return shim.Error("Incorrect no. of arguments. Expecting: 1 - {id}")
	}

	// ==== ABAC ====
	// Check if the invoker's role is bda : lan.role=bda
	attrVal, ok, err := cid.GetAttributeValue(stub, "lan.role")
	if ok == false {
		fmt.Println("Attribute id not found")
		return shim.Error("only lan.role = 'fa' or 'la' attribute users can approve layout.")
	}
	if err != nil {
		return shim.Error(err.Error())
	}
	if attrVal != "fa" && attrVal != "la" {
		fmt.Println("Attribute role: " + attrVal)
		return shim.Error("only lan.role = 'fa' or 'la' attribute users can approve layout.")
	}

	// ====
	layoutId := args[0]
	txid := stub.GetTxID()
	compositeIndexName := "id~address~noc~fa~la~txID"
	// ==== Get state ====
	/*deltaResultsIterator, err := stub.GetStateByPartialCompositeKey(compositeIndexName, []string{layoutId})
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve value for %s: %s", layoutId, err.Error()))
	}
	defer deltaResultsIterator.Close()

	// Check if layout already exists
	if !deltaResultsIterator.HasNext() {
		return shim.Error(fmt.Sprintf("Layout: %s does not exist.", layoutId))
	}*/

	// ==== Check if layout exists
	queryArgs := []string{layoutId, attrVal}
	result, err := seperateRangeQuery(stub, queryArgs) // Also checks if layout exists
	if err != nil {
		return shim.Error(err.Error())
	}

	// ==== Check if requested for NOC
	if result.RequestedNOC == "false" {
		return shim.Error(fmt.Sprintf("Layout: %s is not requested for NOC.", result.Id))
	}
	// ==== Check if already Approved ====
	if attrVal == "fa" && result.fieldVal == "APPROVED" {
		return shim.Error(fmt.Sprintf("Layout: %s already Approved by FA", result.Id))
	} else if attrVal == "la" && result.fieldVal == "APPROVED" {
		return shim.Error(fmt.Sprintf("Layout: %s already Approved by LA", result.Id))
	}

	// ==== Putstate process ==
	address := ""
	noc := ""
	fa := ""
	la := ""
	if attrVal == "fa" {
		fa = "APPROVED"
	} else if attrVal == "la" {
		la = "APPROVED"
	}

	compositeKey, compositeErr := stub.CreateCompositeKey(compositeIndexName, []string{layoutId, address, noc, fa, la, txid})
	if compositeErr != nil {
		return shim.Error(fmt.Sprintf("Could not create a composite key for %s: %s", layoutId, compositeErr.Error()))
	}

	// Building compositeIndexName based on who is invoking
	compositeIndexNameFAOrLA := "id~noc~" + attrVal
	// fa or la
	compositeKeyFaOrLa, compositeErr := stub.CreateCompositeKey(compositeIndexNameFAOrLA, []string{layoutId, address, noc, fa, la, txid})
	if compositeErr != nil {
		return shim.Error(fmt.Sprintf("Could not create a composite key for %s: %s", layoutId, compositeErr.Error()))
	}

	// Save the composite key index
	compositePutErr := stub.PutState(compositeKey, []byte{0x00})
	if compositePutErr != nil {
		return shim.Error(fmt.Sprintf("Could not put operation for %s in the ledger: %s", layoutId, compositePutErr.Error()))
	}

	compositePutErr = stub.PutState(compositeKeyFaOrLa, []byte{0x00})
	if compositePutErr != nil {
		return shim.Error(fmt.Sprintf("Could not put operation for %s in the ledger: %s", layoutId, compositePutErr.Error()))
	}

	// Publish event
	eventPayload := "Approved Layout: " + layoutId + " by " + attrVal
	payloadAsBytes := []byte(eventPayload)
	err = stub.SetEvent("mainEvent", payloadAsBytes)

	return shim.Success([]byte(fmt.Sprintf("Successfully approved layout: %s by %s", layoutId, attrVal)))

}

func (t *lacc) rejectLayout(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	// Check if the invoker's role is bda : lan.role=bda
	attrVal, ok, err := cid.GetAttributeValue(stub, "lan.role")
	if ok == false {
		fmt.Println("Attribute id not found")
		return shim.Error("only lan.role = 'fa' or 'la' attribute users can reject layout.")
	}
	if err != nil {
		return shim.Error(err.Error())
	}
	if attrVal != "fa" && attrVal != "la" {
		fmt.Println("Attribute role: " + attrVal)
		return shim.Error("only lan.role = 'fa' or 'la' attribute users can reject layout.")
	}

	// 1
	// ID
	if len(args) != 1 {
		return shim.Error("Expecting one arg: ID")
	}

	// Get state from worldstate
	layoutId := args[0]
	layoutAsBytes, err := stub.GetState(layoutId)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Unmarshall JSON data stored as Bytes to Object
	updatedLayout := layout{}
	err = json.Unmarshal(layoutAsBytes, &updatedLayout)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Checking if layout is requested for NOC
	if updatedLayout.RequestedNOC == false {
		return shim.Error("Layout is not requested for NOC")
	}

	// Changing FAStatus or LAStatus to APPROVED respectively for the role invoked
	if attrVal == "fa" {
		if updatedLayout.FAStatus == "REJECTED" {
			return shim.Error("Layout is already rejected by FA")
		}
		updatedLayout.FAStatus = "REJECTED"
	} else if attrVal == "la" {
		if updatedLayout.LAStatus == "REJECTED" {
			return shim.Error("Layout is already rejected by LA")
		}
		updatedLayout.LAStatus = "REJECTED"
	}

	// Changing total ApprovalStatus respectively.
	if updatedLayout.FAStatus == "APPROVED" && updatedLayout.LAStatus == "APPROVED" {
		updatedLayout.ApprovalStatus = "APPROVED"
	} else if updatedLayout.FAStatus == "REJECTED" || updatedLayout.LAStatus == "REJECTED" {
		updatedLayout.ApprovalStatus = "REJECTED"
	} else {
		updatedLayout.ApprovalStatus = "NA"
	}

	// Marshaling back to json and storing in worldstate
	updatedLayoutAsBytes, _ := json.Marshal(updatedLayout)
	err = stub.PutState(layoutId, updatedLayoutAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Also updating in all transactions world state
	err = stub.PutState(allTransactionKey, updatedLayoutAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Publish event
	eventPayload := "Rejected Layout: " + layoutId + " by " + attrVal
	payloadAsBytes := []byte(eventPayload)
	err = stub.SetEvent("mainEvent", payloadAsBytes)

	return shim.Success(nil)

}

// ======== HT Version ========

func (t *lacc) rejectLayoutHT(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	// ====  1
	// ==== id

	if len(args) != 1 {
		return shim.Error("Incorrect no. of arguments. Expecting: 1 - {id}")
	}

	// ==== ABAC ====
	// Check if the invoker's role is bda : lan.role=bda
	attrVal, ok, err := cid.GetAttributeValue(stub, "lan.role")
	if ok == false {
		fmt.Println("Attribute id not found")
		return shim.Error("only lan.role = 'fa' or 'la' attribute users can approve layout.")
	}
	if err != nil {
		return shim.Error(err.Error())
	}
	if attrVal != "fa" && attrVal != "la" {
		fmt.Println("Attribute role: " + attrVal)
		return shim.Error("only lan.role = 'fa' or 'la' attribute users can approve layout.")
	}

	// ====
	layoutId := args[0]
	txid := stub.GetTxID()
	compositeIndexName := "id~address~noc~fa~la~txID"

	// ==== Check if layout exists
	resultAsResponse := t.viewLayoutHT(stub, args) // Also checks if layout exists
	if resultAsResponse.Status == ERROR {
		return shim.Error("Problem retrieving from viewLayoutHT.")
	}

	resultAsBytes := resultAsResponse.Payload

	// Unmarshelling result
	result := &layoutHT{}
	json.Unmarshal(resultAsBytes, result)

	// ==== Check if requested for NOC
	if result.RequestedNOC == "false" {
		return shim.Error(fmt.Sprintf("Layout: %s is not requested for NOC.", result.Id))
	}
	// ==== Check if already Approved ====
	if attrVal == "fa" && result.FAStatus == "REJECTED" {
		return shim.Error(fmt.Sprintf("Layout: %s already Rejected by FA", result.Id))
	} else if attrVal == "la" && result.LAStatus == "REJECTED" {
		return shim.Error(fmt.Sprintf("Layout: %s already Rejected by LA", result.Id))
	}

	// ==== Putstate process ==
	address := ""
	noc := ""
	fa := ""
	la := ""
	if attrVal == "fa" {
		fa = "REJECTED"
	} else if attrVal == "la" {
		la = "REJECTED"
	}

	compositeKey, compositeErr := stub.CreateCompositeKey(compositeIndexName, []string{layoutId, address, noc, fa, la, txid})
	if compositeErr != nil {
		return shim.Error(fmt.Sprintf("Could not create a composite key for %s: %s", layoutId, compositeErr.Error()))
	}

	// Building compositeIndexName based on who is invoking
	compositeIndexNameFAOrLA := "id~noc~" + attrVal
	// fa or la
	compositeKeyFaOrLa, compositeErr := stub.CreateCompositeKey(compositeIndexNameFAOrLA, []string{layoutId, address, noc, fa, la, txid})
	if compositeErr != nil {
		return shim.Error(fmt.Sprintf("Could not create a composite key for %s: %s", layoutId, compositeErr.Error()))
	}

	// Save the composite key index
	compositePutErr := stub.PutState(compositeKey, []byte{0x00})
	if compositePutErr != nil {
		return shim.Error(fmt.Sprintf("Could not put operation for %s in the ledger: %s", layoutId, compositePutErr.Error()))
	}
	// fa or la
	compositePutErr = stub.PutState(compositeKeyFaOrLa, []byte{0x00})
	if compositePutErr != nil {
		return shim.Error(fmt.Sprintf("Could not put operation for %s in the ledger: %s", layoutId, compositePutErr.Error()))
	}

	// Publish event
	eventPayload := "Rejected Layout: " + layoutId + " by " + attrVal
	payloadAsBytes := []byte(eventPayload)
	err = stub.SetEvent("mainEvent", payloadAsBytes)

	return shim.Success([]byte(fmt.Sprintf("Successfully rejected layout: %s by %s", layoutId, attrVal)))

}

func (t *lacc) getHistory(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var err error
	// 1
	// ID
	if len(args) != 1 {
		return shim.Error("Expecting one argument: ID")
	}

	layoutId := args[0]
	resultsIterator, err := stub.GetHistoryForKey(layoutId)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the layout
	// var buffer bytes.Buffer
	// buffer.WriteString("[")
	layoutHistoryResult := []layoutHistory{}

	// bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		/*
			// Add a comma before array members, suppress it for the first array member
			if bArrayMemberAlreadyWritten == true {
				buffer.WriteString(",")
			}
			buffer.WriteString("{\"TxId\":")
			buffer.WriteString("\"")
			buffer.WriteString(response.TxId)
			//buffer.WriteString("\"\n")

			buffer.WriteString(", \"Value\":")
			// if it was a delete operation on given key, then we need to set the
			//corresponding value null. Else, we will write the response.Value
			//as-is (as the Value itself a JSON marble)
			if response.IsDelete {
				buffer.WriteString("null")
			} else {
				buffer.WriteString(string(response.Value))
			}

			buffer.WriteString(", \"Timestamp\":")
			buffer.WriteString("\"")
			buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
			buffer.WriteString("\"")

			buffer.WriteString(", \"IsDelete\":")
			buffer.WriteString("\"")
			buffer.WriteString(strconv.FormatBool(response.IsDelete))
			buffer.WriteString("\"")

			buffer.WriteString("}")
			bArrayMemberAlreadyWritten = true
		*/

		//
		temp := layoutHistory{}
		temp.TxId = response.TxId
		if response.IsDelete {
			temp.Id = "null"
			temp.FAStatus = "null"
			temp.LAStatus = "null"
			// temp.RequestedNOC = "null"
			temp.ApprovalStatus = "null"
			temp.Address = "null"
		} else {
			tempLayout := layout{}
			err = json.Unmarshal(response.Value, &tempLayout) //unmarshal it aka JSON.parse()
			if err != nil {
				return shim.Error(err.Error())
			}
			temp.Id = tempLayout.Id
			temp.Address = tempLayout.Address
			temp.FAStatus = tempLayout.FAStatus
			temp.LAStatus = tempLayout.LAStatus
			temp.RequestedNOC = tempLayout.RequestedNOC
			temp.ApprovalStatus = tempLayout.ApprovalStatus
		}
		temp.TimeStamp = time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String()
		temp.IsDelete = response.IsDelete

		layoutHistoryResult = append(layoutHistoryResult, temp)

	}
	// buffer.WriteString("]")
	layoutHistoryResultAsBytes, _ := json.Marshal(layoutHistoryResult)

	return shim.Success(layoutHistoryResultAsBytes)
}

// Encrypter exposes how to write state to the ledger after having
// encrypted it with an AES 256 bit key that has been provided to the chaincode through the
// transient field
func (t *lacc) Encrypter(stub shim.ChaincodeStubInterface, args []string, encKey, IV []byte) pb.Response {
	// create the encrypter entity - we give it an ID, the bccsp instance, the key and (optionally) the IV
	ent, err := entities.NewAES256EncrypterEntity("ID", t.bccspInst, encKey, IV)
	if err != nil {
		return shim.Error(fmt.Sprintf("entities.NewAES256EncrypterEntity failed, err %s", err))
	}

	// Check if the invoker's role is bda : lan.role=bda
	val, ok, err := cid.GetAttributeValue(stub, "lan.role")
	if ok == false {
		fmt.Println("Attribute id not found")
		return shim.Error("only lan.role = 'bda' attribute users can create layout.")
	}
	if err != nil {
		return shim.Error(err.Error())
	}
	if val != "bda" {
		fmt.Println("Attribute role: " + val)
		return shim.Error("only lan.role = 'bda' attribute users can create layout.")
	}
	fmt.Println("Attribute value:" + val)

	//   0       1
	// "Id", "Address"
	if len(args) != 2 {
		return shim.Error("Expecting two args: ID and Address")
	}

	// ==== Input sanitation ====
	fmt.Println("- Layout creation started")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}

	layoutId := args[0]
	Address := strings.ToLower(args[1])

	// ==== Check if layout already exists ====
	layoutAsBytes, err := stub.GetState(layoutId)
	if err != nil {
		return shim.Error("Failed to get Layout: " + err.Error())
	} else if layoutAsBytes != nil {
		fmt.Println("This layout already exists: " + layoutId)
		return shim.Error("This layout already exists: " + layoutId)
	}

	// ==== Create layout object and marshal to JSON ====
	AssetType := "layout"
	RequestedNOC := false
	LAStatus := "NA"
	FAStatus := "NA"
	ApprovalStatus := "NA"

	layout := &layout{AssetType, layoutId, Address, RequestedNOC, LAStatus, FAStatus, ApprovalStatus}
	layoutJSONAsBytes, err := json.Marshal(layout)
	if err != nil {
		return shim.Error(err.Error())
	}

	// here, we encrypt cleartextValue and assign it to key
	err = encryptAndPutState(stub, ent, layoutId, layoutJSONAsBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("encryptAndPutState failed, err %+v", err))
	}

	// ==== Layout Saved . Return success ====
	fmt.Println("- End create layout")
	return shim.Success(nil)
}

// Decrypter exposes how to read from the ledger and decrypt using an AES 256
// bit key that has been provided to the chaincode through the transient field.
func (t *lacc) Decrypter(stub shim.ChaincodeStubInterface, args []string, decKey, IV []byte) pb.Response {
	// create the encrypter entity - we give it an ID, the bccsp instance, the key and (optionally) the IV
	ent, err := entities.NewAES256EncrypterEntity("ID", t.bccspInst, decKey, IV)
	if err != nil {
		return shim.Error(fmt.Sprintf("entities.NewAES256EncrypterEntity failed, err %s", err))
	}

	// 1
	// layoutId

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expected: 1 - LayoutId ")
	}

	fmt.Println("Process started..")

	layoutId := args[0]
	if err != nil {
		return shim.Error(err.Error())
	}

	// here we decrypt the state associated to key
	layoutAsBytes, err := getStateAndDecrypt(stub, ent, layoutId)
	if err != nil {
		return shim.Error(fmt.Sprintf("getStateAndDecrypt failed, err %+v", err))
	}

	// here we return the decrypted value as a result
	return shim.Success(layoutAsBytes)
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	factory.InitFactories(nil)

	err := shim.Start(&lacc{factory.GetDefault()})
	// Create a new Smart Contract
	//err := shim.Start(new(lacc))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}

}
