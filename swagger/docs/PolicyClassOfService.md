# PolicyClassOfService

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CosTrafficClass** | **string** | the cos class e.g. Voice, Broadcast etc | [optional] [default to null]
**CosMinGuaranteeBwPercent** | **int32** |  | [optional] [default to null]
**CosLatencyMs** | **int32** | latency SLA for the cos class | [optional] [default to null]
**CosJitterMs** | **int32** | jitter SLA for the cos class | [optional] [default to null]
**CosLossPercent** | **int32** |  | [optional] [default to null]
**CosLlq** | **bool** | choose to use a low latency queue (llq) | [optional] [default to null]
**CosPriority** | **string** | priority level | [optional] [default to null]
**CosLastResort** | **bool** | control specific traffic types to pass through if the only\\ available link is the metered/standby link | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

