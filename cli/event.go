package cli

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient/seek"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"log"
)

func CreatEventClient(orgClient *Client) *event.Client{
	//clientChannelContext := orgClient.SDK.ChannelContext(orgClient.ChannelID, fabsdk.WithOrg(orgClient.OrgName), fabsdk.WithUser(orgClient.OrgUser))
	// New event client
	cp := orgClient.SDK.ChannelContext(orgClient.ChannelID, fabsdk.WithUser(orgClient.OrgUser))
	ec, err := event.New(
		cp,
		event.WithBlockEvents(), // 如果没有，会是filtered
		// event.WithBlockNum(1), // 从指定区块获取，需要此参数
		event.WithSeekType(seek.Newest))
	if err != nil {
		log.Printf("Create event client error: %v", err)
	}
	log.Printf("event client created")
	return ec
}