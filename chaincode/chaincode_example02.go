package main

import (
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"strconv"
)

type MyChainCode struct {}

func (t *MyChainCode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("MyChainCode Init")
	_, args := stub.GetFunctionAndParameters()
	var A, B string    // Entities
	var Aval, Bval int // Asset holdings
	var err error
	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}
	// Initialize the chaincode
	A = args[0]
	Aval, err = strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	B = args[2]
	Bval, err = strconv.Atoi(args[3])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	fmt.Printf("Aval = %d, Bval = %d\n", Aval, Bval)

	// Write the state to the ledger
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	var as []byte
	for _, a := range args {
		as = append(as, []byte(a)...)
	}

	return shim.Success(nil)
}

func (t *MyChainCode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("MyChainCode Invoke")
	function, args := stub.GetFunctionAndParameters()
	if function == "query" {
		return t.query(stub, args)
	} else if function == "transfer" {
		return t.transfer(stub, args)
	} else if function == "add" {
		return t.add(stub, args)
	} else if function == "delete" {
		return t.delete(stub, args)
	} else if function == "setEvent" {
		return t.setEvent(stub, args)
	}
	return shim.Error("Invalid invoke function name. ")
}

func (t *MyChainCode) add(stub shim.ChaincodeStubInterface, args []string) pb.Response{
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	userName,userBalance := args[0],args[1]
	err:= stub.PutState(userName, []byte(userBalance))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (t *MyChainCode) transfer(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var X int          // Transaction value
	var err error
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}
	// Get the Key A and B
	A:= args[0]
	B:= args[1]
	// Get the state from A
	AvalBytes, err := stub.GetState(A)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if AvalBytes == nil {
		return shim.Error("Entity not found")
	}
	Aval, _ := strconv.Atoi(string(AvalBytes))//string to int
	// Get the state from B
	BvalBytes, err := stub.GetState(B)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if BvalBytes == nil {
		return shim.Error("Entity not found")
	}
	Bval, _ := strconv.Atoi(string(BvalBytes))//string to int
	// Perform the execution
	X, err = strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Invalid transaction amount, expecting a integer value")
	}
	Aval = Aval - X
	Bval = Bval + X
	// Write the A state
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return shim.Error(err.Error())
	}
	// Write the B state
	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (t *MyChainCode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	key := args[0]
	value , err := stub.GetState(key)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(value)
}

func (t *MyChainCode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	key := args[0]
	err := stub.DelState(key)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (t *MyChainCode) setEvent(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	payload := args[0]
	EventName := "myevent"
	if err := stub.SetEvent(EventName, []byte(payload)); err != nil {
		return shim.Error(fmt.Errorf("set event: %w", err).Error())
	}
	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(MyChainCode))
	if err != nil {
		fmt.Printf("Error starting MyChainCode: %s", err)
	}
}


