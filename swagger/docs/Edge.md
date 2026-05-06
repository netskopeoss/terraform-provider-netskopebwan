# Edge

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**DateCreated** | [**time.Time**](time.Time.md) | Time object record was created in ISO 8601 format. For example 2019-05-08T05:30:30.206Z | [optional] [default to null]
**DateModified** | [**time.Time**](time.Time.md) | Time object record was last modified in ISO 8601 format. For example &#x27;2019-05-08T05:30:30.206Z&#x27; | [optional] [default to null]
**CreatedBy** | [***UserRef**](UserRef.md) |  | [optional] [default to null]
**ModifiedBy** | [***UserRef**](UserRef.md) |  | [optional] [default to null]
**Id** | **string** |  | [optional] [default to null]
**Model** | [***EdgeModel**](EdgeModel.md) |  | [optional] [default to null]
**Name** | **string** | The display name of the edge | [optional] [default to null]
**Description** | **string** | Additional notes about the edge | [optional] [default to null]
**Activated** | **bool** | True if edge is activated | [optional] [default to null]
**Role** | [***EdgeRole**](EdgeRole.md) |  | [optional] [default to null]
**Serialnumber** | **string** | Serial number of the edge | [optional] [default to null]
**Swversion** | **string** | Version of the software manifest assigned to this edge | [optional] [default to null]
**Swmanifest** | **string** | URL of the software manifest assined to this edge | [optional] [default to null]
**OverlayConfiguration** | [***EdgeOverlayConfiguration**](EdgeOverlayConfiguration.md) |  | [optional] [default to null]
**Psk** | **string** |  | [optional] [default to null]
**PublicKey** | **string** |  | [optional] [default to null]
**BgpConfiguration** | [**[]EdgeBgpConfiguration**](EdgeBGPConfiguration.md) |  | [optional] [default to null]
**StaticRoutes** | [**[]StaticRoute**](StaticRoute.md) |  | [optional] [default to null]
**MqttConfiguration** | [***MqttgcpConfiguration**](MQTTGCPConfiguration.md) |  | [optional] [default to null]
**AssignedPolicy** | [***PolicyRef**](PolicyRef.md) |  | [optional] [default to null]
**One2OneNatRules** | [**[]InboundNatRule**](InboundNatRule.md) |  | [optional] [default to null]
**PortForwardingNatRules** | [**[]InboundNatRule**](InboundNatRule.md) |  | [optional] [default to null]
**Interfaces** | [**[]InterfaceSettings**](InterfaceSettings.md) |  | [optional] [default to null]
**IsTemplate** | **bool** |  | [optional] [default to null]
**ClientConfiguration** | [***ClientConfiguration**](ClientConfiguration.md) |  | [optional] [default to null]
**Source** | [***ObjectRef**](ObjectRef.md) |  | [optional] [default to null]
**Managed** | **bool** |  | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

