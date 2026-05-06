# {{classname}}

All URIs are relative to *https://virtserver.swaggerhub.com/infiot/infiot-api/v2*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetGatewayAggregatedDeviceFlowStatsByID**](MonitoringApi.md#GetGatewayAggregatedDeviceFlowStatsByID) | **Get** /monitoring/gateways/{id}/devices/device_flows_totals | Get gateway device flows aggregated metrics
[**GetGatewayAggregatedDeviceStatsByID**](MonitoringApi.md#GetGatewayAggregatedDeviceStatsByID) | **Get** /monitoring/gateways/{id}/devices/devices_totals | Get gateway devices aggregated metrics
[**GetGatewayAggregatedPathAndLinkStatsByID**](MonitoringApi.md#GetGatewayAggregatedPathAndLinkStatsByID) | **Get** /monitoring/gateways/{id}/wan/paths_links_totals | Get gateway paths and links aggregated metrics
[**GetGatewayLTESignalStatsByID**](MonitoringApi.md#GetGatewayLTESignalStatsByID) | **Get** /monitoring/gateways/{id}/system/lte | Get gateway LTE signal time series statistics
[**GetGatewayLatestInterfaceStatsByID**](MonitoringApi.md#GetGatewayLatestInterfaceStatsByID) | **Get** /monitoring/gateways/{id}/overview/interfaces_latest | Get gateway interfaces latest metrics
[**GetGatewayLatestPathStatsByID**](MonitoringApi.md#GetGatewayLatestPathStatsByID) | **Get** /monitoring/gateways/{id}/overview/paths_latest | Get gateway paths latest metrics
[**GetGatewayLatestRouteStatsByID**](MonitoringApi.md#GetGatewayLatestRouteStatsByID) | **Get** /monitoring/gateways/{id}/overview/routes_latest | Get gateway routes latest metrics
[**GetGatewayLoadStatsByID**](MonitoringApi.md#GetGatewayLoadStatsByID) | **Get** /monitoring/gateways/{id}/system/load | Get gateway system load time series statistics
[**GetGatewayMemoryStatsByID**](MonitoringApi.md#GetGatewayMemoryStatsByID) | **Get** /monitoring/gateways/{id}/system/memory | Get gateway memory time series statistics
[**GetGatewayPathAndLinkStatsByID**](MonitoringApi.md#GetGatewayPathAndLinkStatsByID) | **Get** /monitoring/gateways/{id}/wan/paths_links | Get gateway paths and links time series metrics
[**GetGatewayUptimeStatsByID**](MonitoringApi.md#GetGatewayUptimeStatsByID) | **Get** /monitoring/gateways/{id}/system/uptime | Get gateway uptime time series statistics
[**GetGatewayWiFiStrengthStatsByID**](MonitoringApi.md#GetGatewayWiFiStrengthStatsByID) | **Get** /monitoring/gateways/{id}/system/wifi | Get gateway WiFi strength time series statistics

# **GetGatewayAggregatedDeviceFlowStatsByID**
> GatewayDevicesAggregatedDeviceFlowStats GetGatewayAggregatedDeviceFlowStatsByID(ctx, id, startDatetime, endDatetime, ip, optional)
Get gateway device flows aggregated metrics

Retrieve the aggregated device flows statistics for a gateway ID and timeframe given as parameters. childTenantId that belongs to an Orgnaization tenant is required when the API is accessed using Master MSP/MSP tenant token. The new value/s of the edge should be populated on the Request body. childTenantId can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token. <br>     | `childTenantId` | Behavior                  |    |-----------------|---------------------------|    |  Not passed     | Error Message Code : `400`|    |  Passed         | Aggregated device flows statistics are succesfully returned   |

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **id** | **string**| Identifier | 
  **startDatetime** | **time.Time**| Timestamp indicating the start of the queried time range, in ISO 8601 format. For example &#x27;2019-05-08T05:30:30.206Z&#x27; | 
  **endDatetime** | **time.Time**| Timestamp indicating the end of the queried time range, in ISO 8601 format. For example &#x27;2019-05-08T05:30:30.206Z&#x27; | 
  **ip** | **string**| IP of the device whose flow stats are to be obtained | 
 **optional** | ***MonitoringApiGetGatewayAggregatedDeviceFlowStatsByIDOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a MonitoringApiGetGatewayAggregatedDeviceFlowStatsByIDOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**GatewayDevicesAggregatedDeviceFlowStats**](GatewayDevicesAggregatedDeviceFlowStats.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetGatewayAggregatedDeviceStatsByID**
> GatewayDevicesAggregatedDeviceStats GetGatewayAggregatedDeviceStatsByID(ctx, id, startDatetime, endDatetime, optional)
Get gateway devices aggregated metrics

Retrieve the aggregated devices statistics for a gateway ID and timeframe given as parameters. childTenantId that belongs to an Orgnaization tenant is required when the API is accessed using Master MSP/MSP tenant token. The new value/s of the edge should be populated on the Request body. childTenantId can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token. <br>     | `childTenantId` | Behavior                  |    |-----------------|---------------------------|    |  Not passed     | Error Message Code : `400`|    |  Passed         | Aggregated devices statistics are succesfully returned   |

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **id** | **string**| Identifier | 
  **startDatetime** | **time.Time**| Timestamp indicating the start of the queried time range, in ISO 8601 format. For example &#x27;2019-05-08T05:30:30.206Z&#x27; | 
  **endDatetime** | **time.Time**| Timestamp indicating the end of the queried time range, in ISO 8601 format. For example &#x27;2019-05-08T05:30:30.206Z&#x27; | 
 **optional** | ***MonitoringApiGetGatewayAggregatedDeviceStatsByIDOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a MonitoringApiGetGatewayAggregatedDeviceStatsByIDOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**GatewayDevicesAggregatedDeviceStats**](GatewayDevicesAggregatedDeviceStats.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetGatewayAggregatedPathAndLinkStatsByID**
> GatewayWanAggregatedPathsAndLinksStats GetGatewayAggregatedPathAndLinkStatsByID(ctx, id, startDatetime, endDatetime, optional)
Get gateway paths and links aggregated metrics

Retrieve the aggregated paths and links statistics for a gateway ID and timeframe given as parameters. childTenantId that belongs to an Orgnaization tenant is required when the API is accessed using Master MSP/MSP tenant token. The new value/s of the edge should be populated on the Request body. childTenantId can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token. <br>     | `childTenantId` | Behavior                  |    |-----------------|---------------------------|    |  Not passed     | Error Message Code : `400`|    |  Passed         | Aggregated paths and links statistics are succesfully returned   |

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **id** | **string**| Identifier | 
  **startDatetime** | **time.Time**| Timestamp indicating the start of the queried time range, in ISO 8601 format. For example &#x27;2019-05-08T05:30:30.206Z&#x27; | 
  **endDatetime** | **time.Time**| Timestamp indicating the end of the queried time range, in ISO 8601 format. For example &#x27;2019-05-08T05:30:30.206Z&#x27; | 
 **optional** | ***MonitoringApiGetGatewayAggregatedPathAndLinkStatsByIDOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a MonitoringApiGetGatewayAggregatedPathAndLinkStatsByIDOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**GatewayWanAggregatedPathsAndLinksStats**](GatewayWANAggregatedPathsAndLinksStats.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetGatewayLTESignalStatsByID**
> GatewaySysLteSignalStats GetGatewayLTESignalStatsByID(ctx, id, startDatetime, endDatetime, optional)
Get gateway LTE signal time series statistics

Retrieve the LTE signal statistics for a gateway ID and timeframe given as parameters. childTenantId that belongs to an Orgnaization tenant is required when the API is accessed using Master MSP/MSP tenant token. The new value/s of the edge should be populated on the Request body. childTenantId can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token. <br>     | `childTenantId` | Behavior                  |    |-----------------|---------------------------|    |  Not passed     | Error Message Code : `400`|    |  Passed         | LTE signal statistics are succesfully returned   |

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **id** | **string**| Identifier | 
  **startDatetime** | **time.Time**| Timestamp indicating the start of the queried time range, in ISO 8601 format. For example &#x27;2019-05-08T05:30:30.206Z&#x27; | 
  **endDatetime** | **time.Time**| Timestamp indicating the end of the queried time range, in ISO 8601 format. For example &#x27;2019-05-08T05:30:30.206Z&#x27; | 
 **optional** | ***MonitoringApiGetGatewayLTESignalStatsByIDOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a MonitoringApiGetGatewayLTESignalStatsByIDOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**GatewaySysLteSignalStats**](GatewaySysLTESignalStats.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetGatewayLatestInterfaceStatsByID**
> GatewayInterfacesLatestStats GetGatewayLatestInterfaceStatsByID(ctx, id, optional)
Get gateway interfaces latest metrics

Retrieve the latest interfaces statistics for a gateway ID and timeframe given as parameters. childTenantId that belongs to an Orgnaization tenant is required when the API is accessed using Master MSP/MSP tenant token. The new value/s of the edge should be populated on the Request body. childTenantId can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token. <br>     | `childTenantId` | Behavior                  |    |-----------------|---------------------------|    |  Not passed     | Error Message Code : `400`|    |  Passed         | Latest interfaces statistics are succesfully returned   |

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **id** | **string**| Identifier | 
 **optional** | ***MonitoringApiGetGatewayLatestInterfaceStatsByIDOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a MonitoringApiGetGatewayLatestInterfaceStatsByIDOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**GatewayInterfacesLatestStats**](GatewayInterfacesLatestStats.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetGatewayLatestPathStatsByID**
> GatewayPathsLatestStats GetGatewayLatestPathStatsByID(ctx, id, optional)
Get gateway paths latest metrics

Retrieve the latest paths statistics for a gateway ID and timeframe given as parameters. childTenantId that belongs to an Orgnaization tenant is required when the API is accessed using Master MSP/MSP tenant token. The new value/s of the edge should be populated on the Request body. childTenantId can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token. <br>     | `childTenantId` | Behavior                  |    |-----------------|---------------------------|    |  Not passed     | Error Message Code : `400`|    |  Passed         | Latest paths statistics are succesfully returned   |

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **id** | **string**| Identifier | 
 **optional** | ***MonitoringApiGetGatewayLatestPathStatsByIDOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a MonitoringApiGetGatewayLatestPathStatsByIDOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**GatewayPathsLatestStats**](GatewayPathsLatestStats.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetGatewayLatestRouteStatsByID**
> GatewayRoutesLatestStats GetGatewayLatestRouteStatsByID(ctx, id, optional)
Get gateway routes latest metrics

Retrieve the latest routes statistics for a gateway ID and timeframe given as parameters. childTenantId that belongs to an Orgnaization tenant is required when the API is accessed using Master MSP/MSP tenant token. The new value/s of the edge should be populated on the Request body. childTenantId can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token. <br>     | `childTenantId` | Behavior                  |    |-----------------|---------------------------|    |  Not passed     | Error Message Code : `400`|    |  Passed         | Latest routes statistics are succesfully returned   |

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **id** | **string**| Identifier | 
 **optional** | ***MonitoringApiGetGatewayLatestRouteStatsByIDOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a MonitoringApiGetGatewayLatestRouteStatsByIDOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**GatewayRoutesLatestStats**](GatewayRoutesLatestStats.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetGatewayLoadStatsByID**
> GatewaySysLoadStats GetGatewayLoadStatsByID(ctx, id, startDatetime, endDatetime, optional)
Get gateway system load time series statistics

Retrieve the system load statistics for a gateway ID and timeframe given as parameters. childTenantId that belongs to an Orgnaization tenant is required when the API is accessed using Master MSP/MSP tenant token. The new value/s of the edge should be populated on the Request body. childTenantId can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token. <br>     | `childTenantId` | Behavior                  |    |-----------------|---------------------------|    |  Not passed     | Error Message Code : `400`|    |  Passed         | System load statistics are succesfully returned   |

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **id** | **string**| Identifier | 
  **startDatetime** | **time.Time**| Timestamp indicating the start of the queried time range, in ISO 8601 format. For example &#x27;2019-05-08T05:30:30.206Z&#x27; | 
  **endDatetime** | **time.Time**| Timestamp indicating the end of the queried time range, in ISO 8601 format. For example &#x27;2019-05-08T05:30:30.206Z&#x27; | 
 **optional** | ***MonitoringApiGetGatewayLoadStatsByIDOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a MonitoringApiGetGatewayLoadStatsByIDOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**GatewaySysLoadStats**](GatewaySysLoadStats.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetGatewayMemoryStatsByID**
> GatewaySysMemoryStats GetGatewayMemoryStatsByID(ctx, id, startDatetime, endDatetime, optional)
Get gateway memory time series statistics

Retrieve the memory statistics for a gateway ID and timeframe given as parameters. childTenantId that belongs to an Orgnaization tenant is required when the API is accessed using Master MSP/MSP tenant token. The new value/s of the edge should be populated on the Request body. childTenantId can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token. <br>     | `childTenantId` | Behavior                  |    |-----------------|---------------------------|    |  Not passed     | Error Message Code : `400`|    |  Passed         | Memory statistics are succesfully returned   |

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **id** | **string**| Identifier | 
  **startDatetime** | **time.Time**| Timestamp indicating the start of the queried time range, in ISO 8601 format. For example &#x27;2019-05-08T05:30:30.206Z&#x27; | 
  **endDatetime** | **time.Time**| Timestamp indicating the end of the queried time range, in ISO 8601 format. For example &#x27;2019-05-08T05:30:30.206Z&#x27; | 
 **optional** | ***MonitoringApiGetGatewayMemoryStatsByIDOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a MonitoringApiGetGatewayMemoryStatsByIDOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**GatewaySysMemoryStats**](GatewaySysMemoryStats.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetGatewayPathAndLinkStatsByID**
> GatewayWanPathsAndLinksStats GetGatewayPathAndLinkStatsByID(ctx, id, startDatetime, endDatetime, metric, optional)
Get gateway paths and links time series metrics

Retrieve the paths and links statistics for a gateway ID and timeframe given as parameters. childTenantId that belongs to an Orgnaization tenant is required when the API is accessed using Master MSP/MSP tenant token. The new value/s of the edge should be populated on the Request body. childTenantId can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token. <br>     | `childTenantId` | Behavior                  |    |-----------------|---------------------------|    |  Not passed     | Error Message Code : `400`|    |  Passed         | paths and links statistics are succesfully returned   |

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **id** | **string**| Identifier | 
  **startDatetime** | **time.Time**| Timestamp indicating the start of the queried time range, in ISO 8601 format. For example &#x27;2019-05-08T05:30:30.206Z&#x27; | 
  **endDatetime** | **time.Time**| Timestamp indicating the end of the queried time range, in ISO 8601 format. For example &#x27;2019-05-08T05:30:30.206Z&#x27; | 
  **metric** | **string**| String indicating what metric is to be returned by wan statistics | 
 **optional** | ***MonitoringApiGetGatewayPathAndLinkStatsByIDOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a MonitoringApiGetGatewayPathAndLinkStatsByIDOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**GatewayWanPathsAndLinksStats**](GatewayWANPathsAndLinksStats.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetGatewayUptimeStatsByID**
> GatewaySysUptimeStats GetGatewayUptimeStatsByID(ctx, id, startDatetime, endDatetime, optional)
Get gateway uptime time series statistics

Retrieve the uptime statistics for a gateway ID and timeframe given as parameters. childTenantId that belongs to an Orgnaization tenant is required when the API is accessed using Master MSP/MSP tenant token. The new value/s of the edge should be populated on the Request body. childTenantId can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token. <br>     | `childTenantId` | Behavior                  |    |-----------------|---------------------------|    |  Not passed     | Error Message Code : `400`|    |  Passed         | Uptime statistics are succesfully returned   |

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **id** | **string**| Identifier | 
  **startDatetime** | **time.Time**| Timestamp indicating the start of the queried time range, in ISO 8601 format. For example &#x27;2019-05-08T05:30:30.206Z&#x27; | 
  **endDatetime** | **time.Time**| Timestamp indicating the end of the queried time range, in ISO 8601 format. For example &#x27;2019-05-08T05:30:30.206Z&#x27; | 
 **optional** | ***MonitoringApiGetGatewayUptimeStatsByIDOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a MonitoringApiGetGatewayUptimeStatsByIDOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**GatewaySysUptimeStats**](GatewaySysUptimeStats.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetGatewayWiFiStrengthStatsByID**
> GatewaySysWiFiStrengthStats GetGatewayWiFiStrengthStatsByID(ctx, id, startDatetime, endDatetime, optional)
Get gateway WiFi strength time series statistics

Retrieve the WiFi strength statistics for a gateway ID and timeframe given as parameters. childTenantId that belongs to an Orgnaization tenant is required when the API is accessed using Master MSP/MSP tenant token. The new value/s of the edge should be populated on the Request body. childTenantId can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token. <br>     | `childTenantId` | Behavior                  |    |-----------------|---------------------------|    |  Not passed     | Error Message Code : `400`|    |  Passed         | WiFi strength statistics are succesfully returned   |

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **id** | **string**| Identifier | 
  **startDatetime** | **time.Time**| Timestamp indicating the start of the queried time range, in ISO 8601 format. For example &#x27;2019-05-08T05:30:30.206Z&#x27; | 
  **endDatetime** | **time.Time**| Timestamp indicating the end of the queried time range, in ISO 8601 format. For example &#x27;2019-05-08T05:30:30.206Z&#x27; | 
 **optional** | ***MonitoringApiGetGatewayWiFiStrengthStatsByIDOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a MonitoringApiGetGatewayWiFiStrengthStatsByIDOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**GatewaySysWiFiStrengthStats**](GatewaySysWiFiStrengthStats.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

