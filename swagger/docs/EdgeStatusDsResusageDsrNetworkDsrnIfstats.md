# EdgeStatusDsResusageDsrNetworkDsrnIfstats

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**DsrnSfreq** | **int32** | Sampling frequency | [optional] [default to null]
**DsrnScount** | **int32** | Sampling count | [optional] [default to null]
**DsrnSindex** | **int32** | Current sample index | [optional] [default to null]
**DsrnPhysicalname** | **string** | Physical name of the interface | [optional] [default to null]
**DsrnIftype** | **string** | Type of the interface | [optional] [default to null]
**DsrnCellularStats** | [***EdgeStatusDsResusageDsrNetworkDsrnCellularStats**](EdgeStatus_ds_resusage_dsr_network_dsrn_cellular_stats.md) |  | [optional] [default to null]
**DsrnLogicalname** | **string** | Logical name of the interface. | [optional] [default to null]
**DsrnIsoverlay** | **bool** | Indicates whether this interface is a overlay interface or not. | [optional] [default to null]
**DsrnIfstate** | **string** | Indicates whether this interface is up or down. | [optional] [default to null]
**DsrnAddress** | **[]string** |  | [optional] [default to null]
**DsrnPublicip** | **string** |  | [optional] [default to null]
**DsrnIfIspname** | **string** | Name of the ISP to which the public ip belongs | [optional] [default to null]
**DsrnIfGeolocation** | [***EdgeStatusDsResusageDsrNetworkDsrnIfGeolocation**](EdgeStatus_ds_resusage_dsr_network_dsrn_if_geolocation.md) |  | [optional] [default to null]
**DsrnTxCapacity** | **string** | Transmission capacity (bandwidth) of this interface measured in bps | [optional] [default to null]
**DsrnRxCapacity** | **string** | Receive capacity (bandwidth) of this interface measured in bps | [optional] [default to null]
**DsrnTrafficstats** | [**[]EdgeStatusDsResusageDsrNetworkDsrnTrafficstats**](EdgeStatus_ds_resusage_dsr_network_dsrn_trafficstats.md) | Snapshot of stats at the start of collection interval (5s for example) | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

