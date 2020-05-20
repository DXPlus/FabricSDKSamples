package cli

import (
	"log"
	"net/http"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/pkg/errors"
)
var (
	peer0Org1 = "peer0.org1.example.com"
	peer0Org2 = "peer0.org2.example.com"
)

// 安装链码到peer节点
func (c *Client) InstallCC(v string, peer string) error {
	targetPeer := resmgmt.WithTargetEndpoints(peer)

	// pack the chaincode
	ccPkg, err := gopackager.NewCCPackage(c.CCPath, c.CCGoPath)
	if err != nil {
		return errors.WithMessage(err, "pack chaincode error")
	}

	// new request of installing chaincode
	req := resmgmt.InstallCCRequest{
		Name:    c.CCID,
		Path:    c.CCPath,
		Version: v,
		Package: ccPkg,
	}

	resps, err := c.rc.InstallCC(req, targetPeer)
	if err != nil {
		return errors.WithMessage(err, "installCC error")
	}

	// check other errors
	var errs []error
	for _, resp := range resps {
		log.Printf("Install  response status: %v", resp.Status)
		if resp.Status != http.StatusOK {
			errs = append(errs, errors.New(resp.Info))
		}
		if resp.Info == "already installed" {
			log.Printf("Chaincode %s already installed on peer: %s.\n",
				c.CCID+"-"+v, resp.Target)
			return nil
		}
	}

	if len(errs) > 0 {
		log.Printf("InstallCC errors: %v", errs)
		return errors.WithMessage(errs[0], "installCC first error")
	}
	return nil
}

// 实例化链码到peer节点
func (c *Client) InstantiateCC(v string, peer string) (fab.TransactionID,
	error) {
	// endorser policy
	org1OrOrg2 := "OR('Org1MSP.member','Org2MSP.member')"
	ccPolicy, err := c.genPolicy(org1OrOrg2)
	if err != nil {
		return "", errors.WithMessage(err, "gen policy from string error")
	}
	// new request
	// Attention: args should include `init` for Request not
	// have a method term to call init
	args := packArgs([]string{"init", "a", "100", "b", "200"})
	req := resmgmt.InstantiateCCRequest{
		Name:    c.CCID,
		Path:    c.CCPath,
		Version: v,
		Args:    args,
		Policy:  ccPolicy,
	}

	// send request and handle response
	reqPeers := resmgmt.WithTargetEndpoints(peer)
	resp, err := c.rc.InstantiateCC(c.ChannelID, req, reqPeers)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return "", nil
		}
		return "", errors.WithMessage(err, "instantiate chaincode error")
	}

	log.Printf("Instantitate chaincode tx: %s", resp.TransactionID)
	return resp.TransactionID, nil
}

func (c *Client) genPolicy(p string) (*common.SignaturePolicyEnvelope, error) {
	// TODO bug, this any leads to endorser invalid
	if p == "ANY" {
		return cauthdsl.SignedByAnyMember([]string{c.OrgName}), nil
	}
	return cauthdsl.FromString(p)
}

// 重新安装链码
func ReinstallChainCode(cli1, cli2 *Client) {
	log.Println("=================== Reinstall chaincode ===================")
	defer log.Println("=================== Reinstall chaincode end ===============")

	if err := cli1.InstallCC("v1", peer0Org1); err != nil {
		log.Panicf("Intall chaincode error: %v", err)
	}
	log.Println("Chaincode has been installed on org1's peers")

	if err := cli2.InstallCC("v1", peer0Org2); err != nil {
		log.Panicf("Intall chaincode error: %v", err)
	}
	log.Println("Chaincode has been installed on org2's peers")

	// InstantiateCC chaincode only need once for each channel
	if _, err := cli1.InstantiateCC("v1", peer0Org1); err != nil {
		log.Panicf("Instantiated chaincode error: %v", err)
	}
	log.Println("Chaincode has been instantiated")
}

// 升级链码
func UpgradeChainCode(cli1, cli2 *Client,version string) {
	log.Println("=================== Upgrade chaincode ===================")
	defer log.Println("================= Upgrade chaincode end =================")
	v := version
	// Install new version chaincode on org1's peers
	if err := cli1.InstallCC(v, peer0Org1); err != nil {
		log.Panicf("Intall chaincode error: %v", err)
	}
	log.Println("Chaincode has been installed on org1's peers")
	// Install new version chaincode on org2's peers
	if err := cli2.InstallCC(v, peer0Org2); err != nil {
		log.Panicf("Intall chaincode error: %v", err)
	}
	log.Println("Chaincode has been installed on org2's peers")
	// Upgrade chaincode only need once for each channel, this is
	if err := cli1.UpgradeCC(v, peer0Org1); err != nil {
		log.Panicf("Upgrade chaincode error: %v", err)
	}
	log.Println("Upgrade chaincode success for channel")
}

// add
func (c *Client) InvokeAdd(peers []string,key string,value string) (fab.TransactionID, error) {
	log.Println("Invoke add")
	// new channel request for invoke
	args := packArgs([]string{key,value})
	req := channel.Request{
		ChaincodeID: c.CCID,
		Fcn:         "add",
		Args:        args,
	}
	// send request and handle response
	// peers is needed
	reqPeers := channel.WithTargetEndpoints(peers...)
	resp, err := c.cc.Execute(req, reqPeers)
	log.Printf("Invoke chaincode response:\n"+
		"id: %v\nvalidate: %v\nchaincode status: %v\n\n",
		resp.TransactionID,
		resp.TxValidationCode,
		resp.ChaincodeStatus)
	if err != nil {
		return "", errors.WithMessage(err, "invoke chaincode error")
	}
	return resp.TransactionID, nil
}

//delete
func (c *Client) InvokeCCDelete(peers []string,key string) (fab.TransactionID, error) {
	log.Println("Invoke delete")
	// new channel request for invoke
	args := packArgs([]string{key})
	req := channel.Request{
		ChaincodeID: c.CCID,
		Fcn:         "delete",
		Args:        args,
	}
	// send request and handle response
	// peers is needed
	reqPeers := channel.WithTargetEndpoints(peers...)
	resp, err := c.cc.Execute(req, reqPeers)
	log.Printf("Invoke chaincode delete response:\n"+
		"id: %v\nvalidate: %v\nchaincode status: %v\n\n",
		resp.TransactionID,
		resp.TxValidationCode,
		resp.ChaincodeStatus)
	if err != nil {
		return "", errors.WithMessage(err, "invoke chaincode error")
	}
	return resp.TransactionID, nil
}

// transfer
func (c *Client) InvokeTransfer(peers []string,key1 string,key2 string,value string) (fab.TransactionID, error) {
	log.Println("Invoke transfer")
	// new channel request for invoke
	args := packArgs([]string{key1,key2,value})
	req := channel.Request{
		ChaincodeID: c.CCID,
		Fcn:         "transfer",
		Args:        args,
	}
	// send request and handle response
	// peers is needed
	reqPeers := channel.WithTargetEndpoints(peers...)
	resp, err := c.cc.Execute(req, reqPeers)
	log.Printf("Invoke chaincode response:\n"+
		"id: %v\nvalidate: %v\nchaincode status: %v\n\n",
		resp.TransactionID,
		resp.TxValidationCode,
		resp.ChaincodeStatus)
	if err != nil {
		return "", errors.WithMessage(err, "invoke chaincode error")
	}
	return resp.TransactionID, nil
}

// query
func (c *Client) QueryCC(peer, key string) error {
	log.Println("Invoke query")
	// new channel request for query
	req := channel.Request{
		ChaincodeID: c.CCID,
		Fcn:         "query",
		Args:        packArgs([]string{key}),
	}
	// send request and handle response
	reqPeers := channel.WithTargetEndpoints(peer)
	resp, err := c.cc.Query(req, reqPeers)
	if err != nil {
		return errors.WithMessage(err, "query chaincode error")
	}

	log.Printf("Query chaincode tx response:\ntx: %s\nresult: %v\n\n",
		resp.TransactionID,
		string(resp.Payload))
	return nil
}

// setEvent
func (c *Client) InvokeSetEvent(peers []string,value string) (fab.TransactionID, error) {
	log.Println("Invoke setEvent")
	// new channel request for invoke
	args := packArgs([]string{value})
	req := channel.Request{
		ChaincodeID: c.CCID,
		Fcn:         "setEvent",
		Args:        args,
	}
	// send request and handle response
	// peers is needed
	reqPeers := channel.WithTargetEndpoints(peers...)
	resp, err := c.cc.Execute(req, reqPeers)
	log.Printf("Invoke chaincode response:\n"+
		"id: %v\nvalidate: %v\nchaincode status: %v\n\n",
		resp.TransactionID,
		resp.TxValidationCode,
		resp.ChaincodeStatus)
	if err != nil {
		return "", errors.WithMessage(err, "invoke chaincode error")
	}
	return resp.TransactionID, nil
}


func (c *Client) UpgradeCC(v string, peer string) error {
	// endorser policy
	org1AndOrg2 := "AND('Org1MSP.member','Org2MSP.member')"
	ccPolicy, err := c.genPolicy(org1AndOrg2)
	if err != nil {
		return errors.WithMessage(err, "gen policy from string error")
	}
	// new request
	// Attention: args should include `init` for Request not
	// have a method term to call init
	// Reset a b's value to test the upgrade
	args := packArgs([]string{"init", "a", "100", "b", "100"})
	req := resmgmt.UpgradeCCRequest{
		Name:    c.CCID,
		Path:    c.CCPath,
		Version: v,
		Args:    args,
		Policy:  ccPolicy,
	}
	// send request and handle response
	reqPeers := resmgmt.WithTargetEndpoints(peer)
	resp, err := c.rc.UpgradeCC(c.ChannelID, req, reqPeers)
	if err != nil {
		return errors.WithMessage(err, "instantiate chaincode error")
	}

	log.Printf("Instantitate chaincode tx: %s", resp.TransactionID)
	return nil
}

func (c *Client) Close() {
	c.SDK.Close()
}

func packArgs(paras []string) [][]byte {
	var args [][]byte
	for _, k := range paras {
		args = append(args, []byte(k))
	}
	return args
}
