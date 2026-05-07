# SyslogServer

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ServerIp** | **string** |  | [optional] [default to null]
**SourceInterface** | **string** |  | [optional] 
**Protocol** | **string** |  | [optional] [default to PROTOCOL.UDP]
**Port** | **int32** |  | [optional] [default to 514]
**Facility** | **string** |  | [optional] [default to local7]
**Tag** | **string** |  | [optional] [default to infiot]
**Applications** | **[]string** | application list whose logs are expected.eg.urlfilter, all | [optional] [default to null]
**Format** | **string** | output format type | [optional] [default to FORMAT.STRING_]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

