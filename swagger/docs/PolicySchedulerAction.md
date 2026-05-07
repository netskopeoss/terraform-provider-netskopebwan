# PolicySchedulerAction

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SchRateLimitEnable** | **bool** | toggle rate limiting | [optional] [default to false]
**SchTxRateLimitKbps** | **int32** | kbps value of the uplink capacity to be used | [optional] [default to null]
**SchRxRateLimitKbps** | **int32** | kbps value of the downlink capacity to be used | [optional] [default to null]
**SchTxRateLimitType** | **string** | choose between policing and shaping for rate limiting | [optional] [default to SCH_TX_RATE_LIMIT_TYPE.POLICER]
**SchDropAlgo** | **string** | drop strategy for policing and when shaping queue is full | [optional] [default to SCH_DROP_ALGO.TAIL_DROP]
**SchQueueLimitBytes** | **int32** | capacity of the shaping queue | [optional] [default to 1024]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

