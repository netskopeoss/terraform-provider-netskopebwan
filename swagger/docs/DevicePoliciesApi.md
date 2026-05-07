# {{classname}}

All URIs are relative to *https://virtserver.swaggerhub.com/infiot/infiot-api/v2*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AddDevicePolicy**](DevicePoliciesApi.md#AddDevicePolicy) | **Post** /devicepolicies | Create new Device Policy
[**DeleteDevicePolicyById**](DevicePoliciesApi.md#DeleteDevicePolicyById) | **Delete** /devicepolicies/{id} | Delete Device policy using policy Id
[**GetAllDevicePolicies**](DevicePoliciesApi.md#GetAllDevicePolicies) | **Get** /devicepolicies | List all device policies
[**GetDevicePolicyById**](DevicePoliciesApi.md#GetDevicePolicyById) | **Get** /devicepolicies/{id} | Get device policy using policy Id
[**UpdateDevicePolicyById**](DevicePoliciesApi.md#UpdateDevicePolicyById) | **Put** /devicepolicies/{id} | Update device policy details using device policy Id

# **AddDevicePolicy**
> ManagedDevicePolicy AddDevicePolicy(ctx, body, optional)
Create new Device Policy

Create a new device policy in an organization tenant. childTenantId is a mandatory parameter to be passed if the API is executed using a MSP or Master MSP tenant token. childTenantId should be organization tenant's Id when accessed from MasterMSP/MSP tenant. It can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token.<br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Error Message Code : `400`|    | Passed |Edge is created successfully based on the Request body args provided. There are two ways to create a Policy, <br><br> - Clone a Policy from an existing edge template (assuming template edge is already provisioned by the MSP or the service provider) <br><br> - Create a new edge by passing all the necessary details. <br> Notice the difference in the mandatory request body parameters for both the options below <br>    | Creation Type      | Mandatory Request Body Parameters |    |--------------------|-------------|    |  Clone From Template  | `\"sourceTenantId\"` `\"sourceObjectId\"` `\"mdplc_name\"`|    |  Create a new Policy | `\"mdplc_name\"` |

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**AddDevicePolicyInput**](AddDevicePolicyInput.md)|  | 
 **optional** | ***DevicePoliciesApiAddDevicePolicyOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a DevicePoliciesApiAddDevicePolicyOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **childTenantId** | **optional.**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**ManagedDevicePolicy**](ManagedDevicePolicy.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DeleteDevicePolicyById**
> ManagedDevicePolicy DeleteDevicePolicyById(ctx, id, optional)
Delete Device policy using policy Id

Delete a Device Policy by passing the Device Policy id as a parameter. childTenantId that belongs to an Orgnaization tenant is required when the API is accessed using Master MSP/MSP tenant token. The new value/s of the edge should be populated on the Request body. childTenantId can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token. <br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Error Message Code : `400`|    | Passed |Policy is successfully deleted

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **id** | **string**| Identifier | 
 **optional** | ***DevicePoliciesApiDeleteDevicePolicyByIdOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a DevicePoliciesApiDeleteDevicePolicyByIdOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**ManagedDevicePolicy**](ManagedDevicePolicy.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetAllDevicePolicies**
> ManagedDevicePolicyList GetAllDevicePolicies(ctx, optional)
List all device policies

List all device policies under an organization tenant. childTenantId  is a mandatory parameter to be passed if the API is executed using a MSP or Master MSP tenant token. The Child Tenant Id should be an organization tenant's Id when accessed from MasterMSP/MSP tenant. It can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token. <br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Error Message Code : `400`|    | Passed |List of all policies configured in the Child tenant will be displayed

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***DevicePoliciesApiGetAllDevicePoliciesOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a DevicePoliciesApiGetAllDevicePoliciesOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**ManagedDevicePolicyList**](ManagedDevicePolicyList.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetDevicePolicyById**
> ManagedDevicePolicy GetDevicePolicyById(ctx, id, optional)
Get device policy using policy Id

Get all the details of a specific device policy by passing the Device Policy id as a parameter. childTenantId that belongs to an Orgnaization tenant is required when the API is accessed using Master MSP/MSP tenant token.It can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token.<br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Error Message Code : `400`|    | Passed |Details of Policy successfully fetched <br>

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **id** | **string**| Identifier | 
 **optional** | ***DevicePoliciesApiGetDevicePolicyByIdOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a DevicePoliciesApiGetDevicePolicyByIdOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**ManagedDevicePolicy**](ManagedDevicePolicy.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdateDevicePolicyById**
> ManagedDevicePolicy UpdateDevicePolicyById(ctx, body, id, optional)
Update device policy details using device policy Id

Updates a Device Policy by passing the Device Policy id as a parameter. childTenantId that belongs to an Orgnaization tenant is required when the API is accessed using Master MSP/MSP tenant token. The new value/s of the edge should be populated on the Request body. childTenantId can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token. <br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Error Message Code : `400`|    | Passed |Policy is successfully updated using the values provided in Request body

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**ManagedDevicePolicy**](ManagedDevicePolicy.md)|  | 
  **id** | **string**| Identifier | 
 **optional** | ***DevicePoliciesApiUpdateDevicePolicyByIdOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a DevicePoliciesApiUpdateDevicePolicyByIdOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **childTenantId** | **optional.**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**ManagedDevicePolicy**](ManagedDevicePolicy.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

