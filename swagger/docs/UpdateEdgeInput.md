# UpdateEdgeInput

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The display name of the edge | [optional] [default to null]
**Description** | **string** | Additional notes about the edge | [optional] [default to null]
**Role** | [***EdgeRole**](EdgeRole.md) |  | [optional] [default to null]
**Serialnumber** | **string** | Serial number of the edge | [optional] [default to null]
**Swversion** | **string** | Version of the software manifest assigned to this edge | [optional] [default to null]
**Swmanifest** | **string** | URL of the software manifest assined to this edge | [optional] [default to null]
**Psk** | **string** |  | [optional] [default to null]
**BgpConfiguration** | [**[]EdgeBgpConfiguration**](EdgeBGPConfiguration.md) |  | [optional] [default to null]
**StaticRoutes** | [**[]StaticRoute**](StaticRoute.md) |  | [optional] [default to null]
**AssignedPolicy** | [***PolicyRef**](PolicyRef.md) |  | [optional] [default to null]
**One2OneNatRules** | [**[]InboundNatRule**](InboundNatRule.md) |  | [optional] [default to null]
**PortForwardingNatRules** | [**[]InboundNatRule**](InboundNatRule.md) |  | [optional] [default to null]
**Interfaces** | [***[]InterfaceSettings**](array.md) |  | [optional] [default to null]
**ClientConfiguration** | [***ClientConfiguration**](ClientConfiguration.md) |  | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

