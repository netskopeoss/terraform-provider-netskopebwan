# {{classname}}

All URIs are relative to *https://virtserver.swaggerhub.com/infiot/infiot-api/v2*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ActivateEdgeById**](EdgesApi.md#ActivateEdgeById) | **Post** /edges/{id}/activate | Create edge activation code
[**AddEdge**](EdgesApi.md#AddEdge) | **Post** /edges | Create a new edge
[**AddEdgeInterface**](EdgesApi.md#AddEdgeInterface) | **Post** /edges/{id}/interfaces | Create new interface
[**DeleteEdgeById**](EdgesApi.md#DeleteEdgeById) | **Delete** /edges/{id} | Delete Edge by Id value
[**DeleteEdgeIfByName**](EdgesApi.md#DeleteEdgeIfByName) | **Delete** /edges/{id}/interfaces/{name} | Delete Edge Interface by name
[**GetAllEdges**](EdgesApi.md#GetAllEdges) | **Get** /edges | List all edges
[**GetEdgeById**](EdgesApi.md#GetEdgeById) | **Get** /edges/{id} | Get edge by Id value
[**GetEdgeIfByName**](EdgesApi.md#GetEdgeIfByName) | **Get** /edges/{id}/interfaces/{name} | Get edge&#x27;s interface details by name
[**GetEdgeInterfaceList**](EdgesApi.md#GetEdgeInterfaceList) | **Get** /edges/{id}/interfaces | Get edge&#x27;s interface list
[**GetEdgeStatusById**](EdgesApi.md#GetEdgeStatusById) | **Get** /edges/{id}/status | Get edge last known status by Id value
[**UpdateEdgeById**](EdgesApi.md#UpdateEdgeById) | **Put** /edges/{id} | Update edge by Id value
[**UpdateEdgeIfByName**](EdgesApi.md#UpdateEdgeIfByName) | **Put** /edges/{id}/interfaces/{name} | Update edge interface details by name

# **ActivateEdgeById**
> EdgeActivationCode ActivateEdgeById(ctx, body, id, optional)
Create edge activation code

Create activation code for an Edge by passing Edge Id as a parameter. childTenantId that belongs to an Orgnaization tenant is required when the API is accessed using Master MSP/MSP tenant token. It can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token. <br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Error Message Code : `400`|    | Passed |Activation code for the Edge successfully generated and emailed <b> Mandatory values to be passed in Request body: </b> emailAddresses and timeoutInSeconds</br>

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**EdgeGenerateActivationCodeInput**](EdgeGenerateActivationCodeInput.md)|  | 
  **id** | **string**| Identifier | 
 **optional** | ***EdgesApiActivateEdgeByIdOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a EdgesApiActivateEdgeByIdOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **childTenantId** | **optional.**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**EdgeActivationCode**](EdgeActivationCode.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **AddEdge**
> Edge AddEdge(ctx, body, optional)
Create a new edge

Create a new Edge. childTenantId is a mandatory parameter to be passed if the API is executed using a MSP or Master MSP tenant token. childTenantId  should be an organization tenant's Id when accessed from MasterMSP/MSP tenant. It can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token.<br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Error Message Code : `400`|    | Passed |Edge is created successfully based on the Request body args provided.  There are two ways to create an Edge, <br><br> - Clone an edge from an existing edge template (assuming template edge is already provisioned by the MSP or the service provider) <br><br> - Create a new edge by passing all the necessary details. <br> Notice the difference in the mandatory request body parameters for both the options below <br>    | Creation Type      | Mandatory Request Body Parameters |    |--------------------|-------------|    |  Clone From Template  | `\"sourceTenantId\"` `\"sourceObjectId\"` `\"name\"`|    |  Create a new edge | `\"name\"` `\"model\"` `\"role\"`|

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**AddEdgeInput**](AddEdgeInput.md)|  | 
 **optional** | ***EdgesApiAddEdgeOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a EdgesApiAddEdgeOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **childTenantId** | **optional.**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**Edge**](Edge.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **AddEdgeInterface**
> Edge AddEdgeInterface(ctx, body, id, optional)
Create new interface

Create an new interface on specific edge by Passing Edge Id. childTenantId that belongs to an Orgnaization tenant is required when the API is accessed using Master MSP/MSP tenant token. It can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token. <br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Error Message Code : `400`|    | Passed |New interface successfully created based on the values provided in Request Body<br>

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**InterfaceSettings**](InterfaceSettings.md)|  | 
  **id** | **string**| Identifier | 
 **optional** | ***EdgesApiAddEdgeInterfaceOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a EdgesApiAddEdgeInterfaceOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **childTenantId** | **optional.**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**Edge**](Edge.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DeleteEdgeById**
> Edge DeleteEdgeById(ctx, id, optional)
Delete Edge by Id value

Delete an Edge by passing the Edge id as a parameter. childTenantId that belongs to an Orgnaization tenant is required when the API is accessed using Master MSP/MSP tenant token.It can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token. <br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Error Message Code : `400`|    | Passed | Edge is successfully deleted

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **id** | **string**| Identifier | 
 **optional** | ***EdgesApiDeleteEdgeByIdOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a EdgesApiDeleteEdgeByIdOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**Edge**](Edge.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DeleteEdgeIfByName**
> Edge DeleteEdgeIfByName(ctx, id, name, optional)
Delete Edge Interface by name

Delete a specific interface on an Edge by passing Edge Id and interface name as parameters. childTenantId that belongs to an Orgnaization tenant is required when the API is accessed using Master MSP/MSP tenant token. It can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token. <br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Error Message Code : `400`|    | Passed |specified Interface successfully deleted

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **id** | **string**| Identifier | 
  **name** | **string**| Name Identifier | 
 **optional** | ***EdgesApiDeleteEdgeIfByNameOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a EdgesApiDeleteEdgeIfByNameOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**Edge**](Edge.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetAllEdges**
> EdgesList GetAllEdges(ctx, optional)
List all edges

Display list of all the edges under a  tenant. childTenantId  is a mandatory parameter to be passed if the API is executed using a MSP or Master MSP tenant token. The Child Tenant Id should be an organization tenant's Id when accessed from MasterMSP/MSP tenant. It can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token. <br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Error Message Code : `400`|    | Passed |List of all edges under the Child tenant will be displayed

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***EdgesApiGetAllEdgesOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a EdgesApiGetAllEdgesOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **maxItems** | **optional.Int32**| Maximum number of items to return | [default to 20]
 **afterCursor** | **optional.String**| Start point | 
 **beforeCursor** | **optional.String**| End Point | 
 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 
 **getTemplates** | **optional.Bool**| if present, templates will be obtained | 

### Return type

[**EdgesList**](EdgesList.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetEdgeById**
> Edge GetEdgeById(ctx, id, optional)
Get edge by Id value

Get all the details of an Edge by passing the Edge id as a parameter. childTenantId that belongs to an Orgnaization tenant is required when the API is accessed using Master MSP/MSP tenant token. It can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token. <br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Error Message Code : `400`|    | Passed | Details of Edge sucessfully fetched <br>

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **id** | **string**| Identifier | 
 **optional** | ***EdgesApiGetEdgeByIdOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a EdgesApiGetEdgeByIdOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**Edge**](Edge.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetEdgeIfByName**
> InterfaceSettings GetEdgeIfByName(ctx, id, name, optional)
Get edge's interface details by name

Get details of a specific interface on an Edge by passing Edge Id and interface name as parameters. childTenantId that belongs to an Orgnaization tenant is required when the API is accessed using Master MSP/MSP tenant token. <br>It can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token. <br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Error Message Code : `400`|    | Passed |Details of the specfic interface on Edge successfully fetched

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **id** | **string**| Identifier | 
  **name** | **string**| Name Identifier | 
 **optional** | ***EdgesApiGetEdgeIfByNameOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a EdgesApiGetEdgeIfByNameOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**InterfaceSettings**](InterfaceSettings.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetEdgeInterfaceList**
> []InterfaceSettings GetEdgeInterfaceList(ctx, id, optional)
Get edge's interface list

Get list of all the interfaces on an Edge by passing Edge Id as a parameter. childTenantId that belongs to an Orgnaization tenant is required when the API is accessed using Master MSP/MSP tenant token.It can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token. <br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Error Message Code : `400`|    | Passed |List of all  the interfaces on an Edge is successfully fetched<br>

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **id** | **string**| Identifier | 
 **optional** | ***EdgesApiGetEdgeInterfaceListOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a EdgesApiGetEdgeInterfaceListOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**[]InterfaceSettings**](array.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetEdgeStatusById**
> EdgeStatusRef GetEdgeStatusById(ctx, id, optional)
Get edge last known status by Id value

Get the latest status of an Edge by passing the Edge id as a parameter. childTenantId that belongs to an Orgnaization tenant is required when the API is accessed using Master MSP/MSP tenant token. It can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token. <br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Error Message Code : `400`|    | Passed | Last known status of Edge successfully fetched

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **id** | **string**| Identifier | 
 **optional** | ***EdgesApiGetEdgeStatusByIdOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a EdgesApiGetEdgeStatusByIdOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**EdgeStatusRef**](EdgeStatusRef.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdateEdgeById**
> Edge UpdateEdgeById(ctx, body, id, optional)
Update edge by Id value

Update an Edge by passing the Edge id as a parameter. childTenantId that belongs to an Orgnaization tenant is required when the API is accessed using Master MSP/MSP tenant token. The new value/s of the edge should be populated on the Request body. childTenantId can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token. <br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Error Message Code : `400`|    | Passed |Edge is successfully updated using the values provided in Request body

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**UpdateEdgeInput**](UpdateEdgeInput.md)|  | 
  **id** | **string**| Identifier | 
 **optional** | ***EdgesApiUpdateEdgeByIdOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a EdgesApiUpdateEdgeByIdOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **childTenantId** | **optional.**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**Edge**](Edge.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdateEdgeIfByName**
> Edge UpdateEdgeIfByName(ctx, body, id, name, optional)
Update edge interface details by name

Update details of a specific interface on an Edge by passing Edge Id and interface name as parameters. childTenantId that belongs to an Orgnaization tenant is required when the API is accessed using Master MSP/MSP tenant token. It can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token. <br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Error Message Code : `400`|    | Passed |Specfied interface updated successfully on Edge

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**InterfaceSettings**](InterfaceSettings.md)|  | 
  **id** | **string**| Identifier | 
  **name** | **string**| Name Identifier | 
 **optional** | ***EdgesApiUpdateEdgeIfByNameOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a EdgesApiUpdateEdgeIfByNameOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **childTenantId** | **optional.**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**Edge**](Edge.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

