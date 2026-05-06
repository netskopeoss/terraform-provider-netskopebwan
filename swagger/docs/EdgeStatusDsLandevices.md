# EdgeStatusDsLandevices

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**DslIpv4** | **string** |  | [optional] [default to null]
**DslName** | **string** | Name of the connected LAN device | [optional] [default to null]
**DslProtocol** | **string** | Protocol of the connected LAN device | [optional] [default to dhcp]
**DslState** | **bool** | State of the device, true indicates device is online | [optional] [default to true]
**DslMacAddr** | **string** | Mac address of this device | [optional] [default to 00:00:00:00:00:00]
**DslTxbytes** | **int32** | Total amount of traffic transmitted by this device in bytes | [optional] [default to 0]
**DslTxpkts** | **int32** | Total amount of traffic transmitted by this device in packets | [optional] [default to 0]
**DslRxbytes** | **int32** | Total amount of traffic received by this device in bytes | [optional] [default to 0]
**DslRxpkts** | **int32** | Total amount of traffic received by this device in packets | [optional] [default to 0]
**DslDhcpFp** | **string** | String containing DHCP finger print. | [optional] [default to null]
**DslDeviceType** | **string** | Indicates whether the device is a LAN device (internal) WAN device (external), special devices added by admin (managed) or the edge itself (infiotedge) | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

