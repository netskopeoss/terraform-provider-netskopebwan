# TrafficMatchCriteria

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**MtchSrcZone** | **string** | source zone to match | [optional] [default to trusted]
**MtchDestZone** | **string** | destination zone to match | [optional] [default to trusted]
**MtchSrcIp** | **string** | source ip address to match | [optional] [default to 255.255.255.255/32]
**MtchSrcMac** | **string** | source mac address to match | [optional] [default to 00:00:00:00:00:00]
**MtchSrcPort** | **string** | source port range to match | [optional] [default to null]
**MtchSrcVlan** | **int32** | source vlan to match | [optional] [default to null]
**MtchDstVlan** | **int32** | destination vlan to match | [optional] [default to null]
**MtchDestIp** | **string** | destination ip address to match | [optional] [default to null]
**MtchDestInternet** | **bool** | match all internet bound client traffic | [optional] [default to false]
**MtchDestPort** | **string** | destination ports to match | [optional] [default to null]
**MtchAppId** | **[]int32** | application id from a list of predefined or user defined applications ids. | [optional] [default to null]
**MtchL4Protocol** | [***L4Protocols**](L4Protocols.md) |  | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

