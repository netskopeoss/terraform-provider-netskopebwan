# {{classname}}

All URIs are relative to *https://virtserver.swaggerhub.com/infiot/infiot-api/v2*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AddManagedDevice**](DevicesApi.md#AddManagedDevice) | **Post** /edges/{id}/devices | Create new managed device
[**DeleteManagedDeviceById**](DevicesApi.md#DeleteManagedDeviceById) | **Delete** /edges/{id}/devices/{deviceId} | Delete Managed Device by Id
[**GetManagedDeviceById**](DevicesApi.md#GetManagedDeviceById) | **Get** /edges/{id}/devices/{deviceId} | Get Managed Device details by Id
[**GetManagedDeviceList**](DevicesApi.md#GetManagedDeviceList) | **Get** /edges/{id}/devices | Get list of managed devices
[**GetOnlineDeviceList**](DevicesApi.md#GetOnlineDeviceList) | **Get** /edges/{id}/onlinedevices | Get list of devices that are currently online
[**UpdateManagedDeviceById**](DevicesApi.md#UpdateManagedDeviceById) | **Put** /edges/{id}/devices/{deviceId} | Update Managed Device details by Id

# **AddManagedDevice**
> ManagedDevice AddManagedDevice(ctx, body, id, optional)
Create new managed device

Create a Managed device connected to the specified edge. childTenantId is a mandatory parameter to be passed if the API is executed using a MSP or Master MSP tenant token. childTenantId  should be an organization tenant's Id when accessed from MasterMSP/MSP tenant. It can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token.<br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Error Message Code : `400`|    | Passed |Managed device connected to the edge is  created successfully based on the Request body args provided.<br> Mandatory values to be passed in Request body: `\"name\"` `\"props\":{\"md_schemaver\"}` `\"props\":{\"md_labels\"}` `\"props\":{\"md_ip\"}` `\"props\":{\"md_mac_address\"}` These values can be retrieved using API `GET /edges/{id}/onlinedevices` if the device is already connected to the edge and is online

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**ManagedDevice**](ManagedDevice.md)|  | 
  **id** | **string**| Identifier | 
 **optional** | ***DevicesApiAddManagedDeviceOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a DevicesApiAddManagedDeviceOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **childTenantId** | **optional.**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**ManagedDevice**](ManagedDevice.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DeleteManagedDeviceById**
> ManagedDevice DeleteManagedDeviceById(ctx, id, deviceId, optional)
Delete Managed Device by Id

Delete a specific managed device on an Edge by passing Edge Id and device Id as parameters. childTenantId that belongs to an Orgnaization tenant is required when the API is accessed using Master MSP/MSP tenant token. It can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token. <br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Error Message Code : `400`|    | Passed |device successfully deleted

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **id** | **string**| Identifier | 
  **deviceId** | **string**| Name Identifier | 
 **optional** | ***DevicesApiDeleteManagedDeviceByIdOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a DevicesApiDeleteManagedDeviceByIdOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**ManagedDevice**](ManagedDevice.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetManagedDeviceById**
> ManagedDevice GetManagedDeviceById(ctx, id, deviceId, optional)
Get Managed Device details by Id

Get Managed device details connected to an Edge by passing Edge Id and device id as  parameters. childTenantId that belongs to an Orgnaization tenant is required when the API is accessed using Master MSP/MSP tenant token.It can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token. <br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Error Message Code : `400`|    | Passed |Details of Managed device on the specified edge is successfully fetched<br>

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **id** | **string**| Identifier | 
  **deviceId** | **string**| Name Identifier | 
 **optional** | ***DevicesApiGetManagedDeviceByIdOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a DevicesApiGetManagedDeviceByIdOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**ManagedDevice**](ManagedDevice.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetManagedDeviceList**
> ManagedDeviceList GetManagedDeviceList(ctx, id, optional)
Get list of managed devices

Display list of all the managed devices connected to an edge. childTenantId  is a mandatory parameter to be passed if the API is executed using a MSP or Master MSP tenant token. The Child Tenant Id should be an organization tenant's Id when accessed from MasterMSP/MSP tenant. It can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token. <br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Error Message Code : `400`|    | Passed |List of all devices connected to the specified edge under the Child tenant will be displayed

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **id** | **string**| Identifier | 
 **optional** | ***DevicesApiGetManagedDeviceListOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a DevicesApiGetManagedDeviceListOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **maxItems** | **optional.Int32**| Maximum number of items to return | [default to 20]
 **afterCursor** | **optional.String**| Start point | 
 **beforeCursor** | **optional.String**| End Point | 
 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**ManagedDeviceList**](ManagedDeviceList.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetOnlineDeviceList**
> ManagedDeviceList GetOnlineDeviceList(ctx, id, optional)
Get list of devices that are currently online

Display list of all the managed and un managed devices connected to an edge that are currently online. childTenantId  is a mandatory parameter to be passed if the API is executed using a MSP or Master MSP tenant token. The Child Tenant Id should be an organization tenant's Id when accessed from MasterMSP/MSP tenant. It can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token. <br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Error Message Code : `400`|    | Passed |List of all devices that are currently online on the specified edge under the Child tenant will be displayed

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **id** | **string**| Identifier | 
 **optional** | ***DevicesApiGetOnlineDeviceListOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a DevicesApiGetOnlineDeviceListOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 
 **tenantMgmtProxyId** | **optional.String**| Id of the edge used as management proxy. | 
 **unClaimedDevices** | **optional.Bool**| Indicate whether only unclaimed devices should be returned or not. | [default to false]

### Return type

[**ManagedDeviceList**](ManagedDeviceList.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdateManagedDeviceById**
> ManagedDevice UpdateManagedDeviceById(ctx, body, id, deviceId, optional)
Update Managed Device details by Id

Update details of a specific Managed device on an Edge by passing Edge Id and device Id as parameters. childTenantId that belongs to an Orgnaization tenant is required when the API is accessed using Master MSP/MSP tenant token. It can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token. <br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Error Message Code : `400`|    | Passed |Specfied device on Edge successfully on Edge

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**ManagedDevice**](ManagedDevice.md)|  | 
  **id** | **string**| Identifier | 
  **deviceId** | **string**| Name Identifier | 
 **optional** | ***DevicesApiUpdateManagedDeviceByIdOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a DevicesApiUpdateManagedDeviceByIdOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **childTenantId** | **optional.**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**ManagedDevice**](ManagedDevice.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

