# PolicyConfigPcfgUrlFilter

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**PcfgUfEnabled** | **bool** |  | [optional] [default to false]
**PcfgUfBlocklist** | **[]string** | List of URLs to be blocked irrespective of category/reputation | [optional] [default to null]
**PcfgUfAllowlist** | **[]string** | List of URLs to be allowed irrespective of category/reputation | [optional] [default to null]
**PcfgUfReputationThreshold** | [***WebReputation**](WebReputation.md) |  | [optional] [default to null]
**PcfgUfBlockedCategories** | **[]int32** | List of categories to block (Infiot category specifier) | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

