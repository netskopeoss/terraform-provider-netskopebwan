# {{classname}}

All URIs are relative to *https://virtserver.swaggerhub.com/infiot/infiot-api/v2*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AddDenySourceIPRuleToPolicy**](PoliciesApi.md#AddDenySourceIPRuleToPolicy) | **Post** /policies/{id}/block-src-ip | Create new firewall block rule for a specific source IP
[**AddPolicy**](PoliciesApi.md#AddPolicy) | **Post** /policies | Create new policy
[**AddQoSSourceIPRuleToPolicy**](PoliciesApi.md#AddQoSSourceIPRuleToPolicy) | **Post** /policies/{id}/qos-rule-src-ip | Create new QoS rule for a specific source IP
[**DeleteDenySourceIPRuleFromPolicy**](PoliciesApi.md#DeleteDenySourceIPRuleFromPolicy) | **Delete** /policies/{id}/block-src-ip/{ip} | Delete a policy firewall rule using match source IP
[**DeletePolicyById**](PoliciesApi.md#DeletePolicyById) | **Delete** /policies/{id} | Delete policy using policy Id
[**DeleteQoSSourceIPRuleFromPolicy**](PoliciesApi.md#DeleteQoSSourceIPRuleFromPolicy) | **Delete** /policies/{id}/qos-rule-src-ip/{ip} | Delete a policy QoS rule using match source IP
[**GetAllPolicies**](PoliciesApi.md#GetAllPolicies) | **Get** /policies | List all policies
[**GetPolicyById**](PoliciesApi.md#GetPolicyById) | **Get** /policies/{id} | Get policy using policy Id
[**UpdatePolicyById**](PoliciesApi.md#UpdatePolicyById) | **Put** /policies/{id} | Update policy details using policy Id

# **AddDenySourceIPRuleToPolicy**
> Policy AddDenySourceIPRuleToPolicy(ctx, body, id, optional)
Create new firewall block rule for a specific source IP

Create a new firewall block rule in a policy of an organizant tenant. childTenantId is a mandatory parameter to be passed if the API is executed using a MSP or Master MSP tenant token. childTenantId should be organization tenant's Id when accessed from MasterMSP/MSP tenant. It can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token.<br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Error Message Code : `400`|    | Passed |Policy rule is created successfully based on the Request body args provided.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**AddPolicyFirewallRuleInput**](AddPolicyFirewallRuleInput.md)|  | 
  **id** | **string**| Identifier | 
 **optional** | ***PoliciesApiAddDenySourceIPRuleToPolicyOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a PoliciesApiAddDenySourceIPRuleToPolicyOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **childTenantId** | **optional.**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**Policy**](Policy.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **AddPolicy**
> Policy AddPolicy(ctx, body, optional)
Create new policy

Create a new policy in an organizant tenant. childTenantId is a mandatory parameter to be passed if the API is executed using a MSP or Master MSP tenant token. childTenantId should be organization tenant's Id when accessed from MasterMSP/MSP tenant. It can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token.<br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Error Message Code : `400`|    | Passed |Edge is created successfully based on the Request body args provided. There are two ways to create a Policy, <br><br> - Clone a Policy from an existing edge template (assuming template edge is already provisioned by the MSP or the service provider) <br><br> - Create a new edge by passing all the necessary details. <br> Notice the difference in the mandatory request body parameters for both the options below <br>    | Creation Type      | Mandatory Request Body Parameters |    |--------------------|-------------|    |  Clone From Template  | `\"sourceTenantId\"` `\"sourceObjectId\"` `\"name\"`|    |  Create a new Policy | `\"name\"` |

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**AddPolicyInput**](AddPolicyInput.md)|  | 
 **optional** | ***PoliciesApiAddPolicyOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a PoliciesApiAddPolicyOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **childTenantId** | **optional.**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**Policy**](Policy.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **AddQoSSourceIPRuleToPolicy**
> Policy AddQoSSourceIPRuleToPolicy(ctx, body, id, optional)
Create new QoS rule for a specific source IP

Create a new QoS rule in a policy of an organizant tenant. childTenantId is a mandatory parameter to be passed if the API is executed using a MSP or Master MSP tenant token. childTenantId should be organization tenant's Id when accessed from MasterMSP/MSP tenant. It can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token.<br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Error Message Code : `400`|    | Passed |Policy rule is created successfully based on the Request body args provided.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**AddPolicyQoSRuleInput**](AddPolicyQoSRuleInput.md)|  | 
  **id** | **string**| Identifier | 
 **optional** | ***PoliciesApiAddQoSSourceIPRuleToPolicyOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a PoliciesApiAddQoSSourceIPRuleToPolicyOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **childTenantId** | **optional.**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**Policy**](Policy.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DeleteDenySourceIPRuleFromPolicy**
> Policy DeleteDenySourceIPRuleFromPolicy(ctx, id, ip, optional)
Delete a policy firewall rule using match source IP

Delete a policy firewall rule by passing the Policy id and source IP as a parameters. childTenantId that belongs to an Orgnaization tenant is required when the API is accessed using Master MSP/MSP tenant token. The new value/s of the edge should be populated on the Request body. childTenantId can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token. <br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Error Message Code : `400`|    | Passed |Policy rule is successfully deleted

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **id** | **string**| Identifier | 
  **ip** | **string**| IP | 
 **optional** | ***PoliciesApiDeleteDenySourceIPRuleFromPolicyOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a PoliciesApiDeleteDenySourceIPRuleFromPolicyOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**Policy**](Policy.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DeletePolicyById**
> Policy DeletePolicyById(ctx, id, optional)
Delete policy using policy Id

Delete a Policy by passing the Policy id as a parameter. childTenantId that belongs to an Orgnaization tenant is required when the API is accessed using Master MSP/MSP tenant token. The new value/s of the edge should be populated on the Request body. childTenantId can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token. <br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Error Message Code : `400`|    | Passed |Policy is successfully deleted

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **id** | **string**| Identifier | 
 **optional** | ***PoliciesApiDeletePolicyByIdOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a PoliciesApiDeletePolicyByIdOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**Policy**](Policy.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DeleteQoSSourceIPRuleFromPolicy**
> Policy DeleteQoSSourceIPRuleFromPolicy(ctx, id, ip, optional)
Delete a policy QoS rule using match source IP

Delete a policy QoS rule by passing the Policy id and source IP as a parameters. childTenantId that belongs to an Orgnaization tenant is required when the API is accessed using Master MSP/MSP tenant token. The new value/s of the edge should be populated on the Request body. childTenantId can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token. <br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Error Message Code : `400`|    | Passed |Policy rule is successfully deleted

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **id** | **string**| Identifier | 
  **ip** | **string**| IP | 
 **optional** | ***PoliciesApiDeleteQoSSourceIPRuleFromPolicyOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a PoliciesApiDeleteQoSSourceIPRuleFromPolicyOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**Policy**](Policy.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetAllPolicies**
> []Policy GetAllPolicies(ctx, optional)
List all policies

List all policies under an organization tenant. childTenantId  is a mandatory parameter to be passed if the API is executed using a MSP or Master MSP tenant token. The Child Tenant Id should be an organization tenant's Id when accessed from MasterMSP/MSP tenant. It can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token. <br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Error Message Code : `400`|    | Passed |List of all policies configured in the Child tenant will be displayed

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***PoliciesApiGetAllPoliciesOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a PoliciesApiGetAllPoliciesOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**[]Policy**](array.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetPolicyById**
> Policy GetPolicyById(ctx, id, optional)
Get policy using policy Id

Get all the details of a specific policy by passing the Policy id as a parameter. childTenantId that belongs to an Orgnaization tenant is required when the API is accessed using Master MSP/MSP tenant token.It can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token.<br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Error Message Code : `400`|    | Passed |Details of Policy successfully fetched <br>

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **id** | **string**| Identifier | 
 **optional** | ***PoliciesApiGetPolicyByIdOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a PoliciesApiGetPolicyByIdOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**Policy**](Policy.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdatePolicyById**
> Policy UpdatePolicyById(ctx, body, id, optional)
Update policy details using policy Id

Updates a Policy by passing the Policy id as a parameter. childTenantId that belongs to an Orgnaization tenant is required when the API is accessed using Master MSP/MSP tenant token. The new value/s of the edge should be populated on the Request body. childTenantId can be empty when using an Organization tenant token. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed when using a MSP/Master MSP token. <br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | Error Message Code : `400`|    | Passed |Policy is successfully updated using the values provided in Request body

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**Policy**](Policy.md)|  | 
  **id** | **string**| Identifier | 
 **optional** | ***PoliciesApiUpdatePolicyByIdOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a PoliciesApiUpdatePolicyByIdOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **childTenantId** | **optional.**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**Policy**](Policy.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

