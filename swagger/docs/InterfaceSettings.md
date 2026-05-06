# InterfaceSettings

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Var8021xMab** | **bool** |  | [optional] [default to false]
**Radius** | [**[]RadiusSetting**](RadiusSetting.md) |  | [optional] [default to null]
**Mode** | **string** |  | [optional] [default to MODE.ROUTED]
**AllowedVlans** | **[]int32** |  | [optional] [default to null]
**Zone** | **string** |  | [optional] [default to trusted]
**BridgeMembers** | **[]string** |  | [optional] [default to null]
**WifiProps** | [***WiFiInterfaceSetting**](WiFiInterfaceSetting.md) |  | [optional] [default to null]
**LteProps** | [***LteInterfaceSetting**](LteInterfaceSetting.md) |  | [optional] [default to null]
**Name** | **string** |  | [optional] [default to null]
**Type_** | **string** |  | [optional] [default to null]
**MacAddr** | **string** |  | [optional] [default to 00:00:00:00:00:00]
**Mtu** | **int32** |  | [optional] [default to 1500]
**MtuDiscovery** | **string** |  | [optional] [default to MTU_DISCOVERY.AUTO]
**IsDisabled** | **bool** |  | [optional] [default to false]
**Vlan** | **int32** |  | [optional] [default to 0]
**EnableNat** | **bool** |  | [optional] [default to false]
**ProxyArpSettings** | [**[]ProxyArpSetting**](ProxyArpSetting.md) |  | [optional] [default to null]
**OverlaySetting** | [***OverlaySetting**](OverlaySetting.md) |  | [optional] [default to null]
**Vrrp** | [***Vrrp**](Vrrp.md) |  | [optional] [default to null]
**DhcpServerSetting** | [***DhcpServerSettings**](DhcpServerSettings.md) |  | [optional] [default to null]
**DhcpRelayServerSetting** | [***[]string**](array.md) |  | [optional] [default to null]
**DoAdvertise** | **bool** | advertise the interface subnet | [optional] [default to false]
**Addresses** | [**[]InterfaceSettingsAddresses**](InterfaceSettings_addresses.md) |  | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

