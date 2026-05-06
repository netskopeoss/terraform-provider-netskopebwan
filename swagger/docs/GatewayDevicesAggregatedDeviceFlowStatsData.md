# GatewayDevicesAggregatedDeviceFlowStatsData

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SrcIp** | **string** | IP address of the connected device | [optional] [default to null]
**SrcPort** | **string** | Source port in the first packet of the flow | [optional] [default to null]
**DestIp** | **string** | Destination IP in the first packet of the flow | [optional] [default to null]
**DestPort** | **string** | Destination port in the first packet of the flow | [optional] [default to null]
**FlowFirstPacketTs** | **string** | Timestamp indicating the first packet arrival for this flow, in RFC 3339 format | [optional] [default to null]
**ConcatPolicies** | **string** | Comma separated policies that were applied to this flow | [optional] [default to null]
**ConcatAppIds** | **string** | Comma separated application ids of this flow | [optional] [default to null]
**ConcatAppDisplaynames** | **string** | Comma separated application display names of this flow | [optional] [default to null]
**TxBytes** | **float32** | Total bytes transmitted in this flow | [optional] [default to null]
**RxBytes** | **float32** | Total bytes received in this flow | [optional] [default to null]
**InterfaceRx** | **string** | Name of the interface through which first packet of this flow was received. | [optional] [default to null]
**InterfaceTx** | **string** | Name of the interface through which first packet of this flow was sent out. | [optional] [default to null]
**OverlayIpTx** | **string** | The overlay ip of the remote peer to which first packet of the flow was sent out. | [optional] [default to null]
**LinkName** | **string** | Name of the link used by the flow | [optional] [default to null]
**SchedulerQueue** | **string** | The scheduler queue assigned to this flow | [optional] [default to null]
**AvgAppLatency** | **float32** | Average application latency measured in milliseconds | [optional] [default to null]
**ProtocolNumber** | **float32** | The value of the protocol number in the IP packet header | [optional] [default to null]
**FlowLastPacketTs** | **string** | Timestamp indicating the last packet arrival for this flow, in RFC 3339 format | [optional] [default to null]
**DestUrls** | **[]string** | Destination URLs of this flow | [optional] [default to null]
**ConcatPeers** | **string** | Comma separated names of the remote peers to which packets of the flow were sent out. | [optional] [default to null]
**VnatIps** | **[]string** | VPN NAT IPs used by this flow | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

