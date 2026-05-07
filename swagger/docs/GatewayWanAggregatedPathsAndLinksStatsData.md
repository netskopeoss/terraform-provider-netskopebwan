# GatewayWanAggregatedPathsAndLinksStatsData

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Link** | **string** | Link through which traffic is sent | [optional] [default to null]
**Isp** | **string** | Name of the ISP to which the public ip belongs | [optional] [default to null]
**Ip** | **string** | IP of local interface forwarding the traffic | [optional] [default to null]
**OverlayIp** | **string** | Overlay IP of the peer | [optional] [default to null]
**Peer** | **string** | Name of the edge forwarding the traffic for paths, &#x27;Link&#x27; for links | [optional] [default to null]
**LatestRxBandwidth** | **float32** | Receive capacity(bandwidth) of this link measured in bps | [optional] [default to null]
**LatestTxBandwidth** | **float32** | Transmission capacity(bandwidth) of this link measured in bps | [optional] [default to null]
**AvgTxThroughput** | **float32** | Average number of bytes transmitted per sample interval | [optional] [default to null]
**MaxTxThroughput** | **float32** | Max number of bytes transmitted in a sample interval | [optional] [default to null]
**AvgRxThroughput** | **float32** | Average number of bytes received per sample interval | [optional] [default to null]
**MaxRxThroughput** | **float32** | Max number of bytes received in a sample interval | [optional] [default to null]
**AvgLatency** | **float32** | Average latency measured in milli seconds | [optional] [default to null]
**MaxLatency** | **float32** | Max latency measured in milli seconds | [optional] [default to null]
**AvgJitter** | **float32** | Average jitter measurement in milli seconds | [optional] [default to null]
**MaxJitter** | **float32** | Max jitter measurement in milli seconds | [optional] [default to null]
**AvgLoss** | **float32** | Average traffic loss in percentage | [optional] [default to null]
**MaxLoss** | **float32** | Max traffic loss in percentage | [optional] [default to null]
**TotalRxBytes** | **float32** | Traffic received in this path/link in bytes | [optional] [default to null]
**TotalTxBytes** | **float32** | Traffic transmitted in this path/link in bytes | [optional] [default to null]
**TotalRxPackets** | **float32** | Traffic received in this path/link in packets | [optional] [default to null]
**TotalTxPackets** | **float32** | Traffic transmitted in this path/link in packets | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

