# {{classname}}

All URIs are relative to *https://virtserver.swaggerhub.com/infiot/infiot-api/v2*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AddUser**](UsersApi.md#AddUser) | **Post** /users | Create a new user
[**DeleteUserById**](UsersApi.md#DeleteUserById) | **Delete** /users/{id} | Delete User using user Id
[**GetAllUsers**](UsersApi.md#GetAllUsers) | **Get** /users | List all users
[**GetUserById**](UsersApi.md#GetUserById) | **Get** /users/{id} | Get user details using user Id
[**UpdateUserById**](UsersApi.md#UpdateUserById) | **Put** /users/{id} | Update user details using user Id

# **AddUser**
> User AddUser(ctx, body, optional)
Create a new user

Creates a new user in the given tenant <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed <br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | User will be created inside the parent tenant (where the API token is created)|    | Passed | User will be created inside the parent tenant (Tenant whose id is `childTenantId`)|

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**User**](User.md)|  | Request Body Params      | Description |  |-----------------|-------------|  |&lt;b&gt;&lt;i&gt;name&lt;/i&gt;&lt;/b&gt; | Name of the User|  |&lt;b&gt;&lt;i&gt;email&lt;/i&gt;&lt;/b&gt; | User email Id |  | &lt;b&gt;&lt;i&gt;roles&lt;/i&gt;&lt;/b&gt; | User role (though the parameter type is a list - a user can have only 1 role assigned). Refer schema &#x60;UserRole&#x60; for different types of acceptable user roles|  | &lt;b&gt;&lt;i&gt;isDisabled&lt;/i&gt;&lt;/b&gt; | If set to true the user would be disabled else enabled| | 
 **optional** | ***UsersApiAddUserOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a UsersApiAddUserOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **childTenantId** | **optional.**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**User**](User.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DeleteUserById**
> User DeleteUserById(ctx, id, optional)
Delete User using user Id

Deletes a specific user uniquely identified by the user Id <br><br> <b><i>Note - Please exercise caution before using this API, the call cannot be undone once its executed.</i></b><br> <b>Notice</b> the difference in behavior when `childTenantId` is passed/not passed <br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | User with the given id (for deletion) would be searched under the parent tenant (where the API token is created)|    | Passed | User with given id (for deletion) would be searched under the parent tenant (Tenant whose id is `childTenantId`|

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **id** | **string**| Identifier | 
 **optional** | ***UsersApiDeleteUserByIdOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a UsersApiDeleteUserByIdOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**User**](User.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetAllUsers**
> []User GetAllUsers(ctx, optional)
List all users

Retrieves all the users (and their details) who have access to the given tenant.<br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed <br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | All users under the parent tenant (where the API token is created) would be listed|    | Passed | All users under the parent tenant (Tenant whose id is `childTenantId`) would be listed|

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***UsersApiGetAllUsersOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a UsersApiGetAllUsersOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**[]User**](array.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetUserById**
> User GetUserById(ctx, id, optional)
Get user details using user Id

Retrieve the details of a given user uniquely identified by the user ID <br>   <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed <br>     | `childTenantId`      | Behavior |    |-----------------|-------------|    |  Not passed  | User with the given id would be searched under the parent tenant (where the API token is created)|    | Passed | User with given id would be searched under the parent tenant (Tenant whose id is `childTenantId`|

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **id** | **string**| Identifier | 
 **optional** | ***UsersApiGetUserByIdOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a UsersApiGetUserByIdOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **childTenantId** | **optional.String**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**User**](User.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdateUserById**
> User UpdateUserById(ctx, body, id, optional)
Update user details using user Id

Update the details of a given user, uniquely identified by the user ID. <br> <b>Note</b> the difference in behavior when `childTenantId` is passed/not passed <br>    | `childTenantId`      | Behavior |   |-----------------|-------------|   |  Not passed  | User with the given id (for updating) would be searched under the parent tenant (where the API token is created)|   | Passed | User with given id (for updating) would be searched under the parent tenant (Tenant whose id is `childTenantId`|

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**User**](User.md)| | Request Body Params      | Description |  |-----------------|-------------|  |&lt;b&gt;&lt;i&gt;name&lt;/i&gt;&lt;/b&gt; | Name of the User|  |&lt;b&gt;&lt;i&gt;email&lt;/i&gt;&lt;/b&gt; | User email Id |  | &lt;b&gt;&lt;i&gt;roles&lt;/i&gt;&lt;/b&gt; | User role (though the parameter type is a list - a user can have only 1 role assigned). Refer schema &#x60;UserRole&#x60; for different types of acceptable user roles|  | &lt;b&gt;&lt;i&gt;isDisabled&lt;/i&gt;&lt;/b&gt; | If set to true the user would be disabled else enabled| | 
  **id** | **string**| Identifier | 
 **optional** | ***UsersApiUpdateUserByIdOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a UsersApiUpdateUserByIdOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **childTenantId** | **optional.**| Tenant Id where the resource exists (Use this parameter if you wish to execute you query to a specific tenant). Make sure the Tenant should be a child of the Tenant where the API token is created | 

### Return type

[**User**](User.md)

### Authorization

[AuthToken](../README.md#AuthToken)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

