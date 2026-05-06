# {{classname}}

All URIs are relative to *https://virtserver.swaggerhub.com/infiot/infiot-api/v2*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AddTenant**](TenantsApi.md#AddTenant) | **Post** /tenants | Create a new tenant
[**DeleteTenantById**](TenantsApi.md#DeleteTenantById) | **Delete** /tenants/{tenantId} | Delete tenant using tenant Id
[**GetAllTenants**](TenantsApi.md#GetAllTenants) | **Get** /tenants | List all tenants
[**GetTenantById**](TenantsApi.md#GetTenantById) | **Get** /tenants/{tenantId} | Get tenant details using tenant Id
[**UpdateTenantById**](TenantsApi.md#UpdateTenantById) | **Put** /tenants/{tenantId} | Update tenant details using tenant Id

# **AddTenant**
> Tenant AddTenant(ctx, body, optional)
Create a new tenant

Creates a new tenant with the given details <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed <br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Tenant creation would be under the tenant where the API token is created|    | Passed | Tenant creation would be under the tenant whose id is `childTenantId`|

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**Tenant**](Tenant.md)|  | Request Body Params      | Description |  |-----------------|-------------|  |&lt;b&gt;&lt;i&gt;name&lt;/i&gt;&lt;/b&gt; | Name of the tenant|  |&lt;b&gt;&lt;i&gt;domainNames&lt;/i&gt;&lt;/b&gt; | List of domain names to access the tenant|  | &lt;b&gt;&lt;i&gt;description&lt;/i&gt;&lt;/b&gt; | A suitable description for the tenant|  | &lt;b&gt;&lt;i&gt;isDisabled&lt;/i&gt;&lt;/b&gt; | If set to true the tenant would be disabled else enabled|  | &lt;b&gt;&lt;i&gt;tenantTypeInput&lt;/i&gt;&lt;/b&gt; | Tenant Type. Refer &#x60;TenantTypeInput&#x60; below for the schema definition | &lt;br&gt; | 
 **optional** | ***TenantsApiAddTenantOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a TenantsApiAddTenantOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **childTenantId** | **optional.**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**Tenant**](Tenant.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DeleteTenantById**
> Tenant DeleteTenantById(ctx, tenantId, optional)
Delete tenant using tenant Id

Deletes a specific tenant uniquely identified by the tenant ID <br><br> <b><i>Note - Please exercise caution before using this API, the call cannot be undone once its executed.</i></b><br> <b>Notice</b> the difference in behavior when `childTenantId` is passed/not passed <br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Tenant with the given id (for deletion) would be searched under the parent tenant (where the API token is created)|    | Passed | Tenant with given id (for deletion) would be searched under the parent tenant     (Tenant whose id is `childTenantId`|

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **tenantId** | **string**| Tenant Identifier | 
 **optional** | ***TenantsApiDeleteTenantByIdOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a TenantsApiDeleteTenantByIdOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**Tenant**](Tenant.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetAllTenants**
> TenantsList GetAllTenants(ctx, optional)
List all tenants

Retrieve details of all the child tenants (Master MSP, MSP, Organization) for the given tenant.<br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed <br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Details of tenants under the parent tenant (where the API token is created) would be listed|    | Passed | Details of tenants under the parent tenant     (tenant whose id is `childTenantId`) would be listed|

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***TenantsApiGetAllTenantsOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a TenantsApiGetAllTenantsOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **maxItems** | **optional.Int32**| Maximum number of items to return | [default to 20]
 **afterCursor** | **optional.String**| Start point | 
 **beforeCursor** | **optional.String**| End Point | 
 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**TenantsList**](TenantsList.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetTenantById**
> Tenant GetTenantById(ctx, tenantId, optional)
Get tenant details using tenant Id

Retrieve the details of a given tenant uniquely identified by the tenant ID <br>   <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed <br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Tenant with the given id would be searched under the parent tenant (where the API token is created)|    | Passed | Tenant with given id would be searched under the parent tenant (Tenant whose id is `childTenantId`|

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **tenantId** | **string**| Tenant Identifier | 
 **optional** | ***TenantsApiGetTenantByIdOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a TenantsApiGetTenantByIdOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**Tenant**](Tenant.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdateTenantById**
> Tenant UpdateTenantById(ctx, body, tenantId, optional)
Update tenant details using tenant Id

Update the details of a given tenant, uniquely identified by the tenant ID. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed <br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Tenant with the given id (for updating) would be searched under the parent tenant (where the API token is created)|    | Passed | Tenant with given id (for updating) would be searched under the parent tenant (Tenant whose id is `childTenantId`|

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**Tenant**](Tenant.md)| | Request Body Params      | Description |  |-----------------|-------------|  |  &lt;b&gt;&lt;i&gt;name&lt;/i&gt;&lt;/b&gt; | Name of the tenant|  | &lt;b&gt;&lt;i&gt;domainNames&lt;/i&gt;&lt;/b&gt; | List of domain names to access the tenant|  | &lt;b&gt;&lt;i&gt;description&lt;/i&gt;&lt;/b&gt; | A suitable description for the tenant|  | &lt;b&gt;&lt;i&gt;isDisabled&lt;/i&gt;&lt;/b&gt; | If set to true the tenant would be disabled else enabled| &lt;br&gt;  &lt;b&gt;Note&lt;/b&gt;: The tenant type parameter &#x60;tenantTypeInput&#x60; cannot be updated, so please either remove it from the request body or set it to old value while invoking this method. | 
  **tenantId** | **string**| Tenant Identifier | 
 **optional** | ***TenantsApiUpdateTenantByIdOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a TenantsApiUpdateTenantByIdOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **childTenantId** | **optional.**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**Tenant**](Tenant.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

