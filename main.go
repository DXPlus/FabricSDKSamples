package main

import (
	"FabricSDKSamples/cli"
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"log"
	"time"
)

const (
	org1CfgPath = "./config/org1sdk-config.yaml"
	org2CfgPath = "./config/org2sdk-config.yaml"
)

var (
	peer0Org1 = "peer0.org1.example.com"
	peer0Org2 = "peer0.org2.example.com"
)

func main() {
	org1Client := cli.New(org1CfgPath, "Org1", "Admin", "User1")
	org2Client := cli.New(org2CfgPath, "Org2", "Admin", "User1")
	defer org1Client.Close()
	defer org2Client.Close()
	//org1Client.CCPath="FabricSDKSamples/chaincode/"
	//org2Client.CCPath="FabricSDKSamples/chaincode/"
	// 升级链码
	//cli.UpgradeChainCode(org1Client, org2Client,"v21")
	//org1Client.InvokeAdd([]string{peer0Org1,peer0Org2},"c","200")
	org1Client.QueryCC(peer0Org1,"a")
	org1Client.QueryCC(peer0Org1,"c")
	org1Client.InvokeTransfer([]string{peer0Org1,peer0Org2},"a","c","50")
	org1Client.QueryCC(peer0Org1,"a")
	org1Client.QueryCC(peer0Org1,"c")
	// chaincode event listen
	ec := cli.CreatEventClient(org2Client)
	chainCodeEventListener(org2Client, ec)
	time.Sleep(time.Second * 10)
}

func chainCodeEventListener(c *cli.Client, ec *event.Client) {
	eventName := ".*"
	log.Printf("Listen chaincode event: %v", eventName)
	ccReg, eventCh, err := ec.RegisterChaincodeEvent("example2", eventName)
	if err != nil {
		fmt.Println("failed to register chaincode event")
	}
	defer ec.Unregister(ccReg)
	fmt.Println("chaincode event registered successfully")
	c.InvokeSetEvent([]string{peer0Org1,peer0Org2},"1. this is a message from chaincode.")
    //接受通道消息
	e := <-eventCh
	log.Printf("Receive cc event, ccid: %v \neventName: %v\n"+ "payload: %v \ntxid: %v \nblock: %v \nsourceURL: %v\n",
					e.ChaincodeID, e.EventName, string(e.Payload), e.TxID, e.BlockNumber, e.SourceURL)
}