# Tenant

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**DateCreated** | [**time.Time**](time.Time.md) | Time object record was created in ISO 8601 format. For example 2019-05-08T05:30:30.206Z | [optional] [default to null]
**DateModified** | [**time.Time**](time.Time.md) | Time object record was last modified in ISO 8601 format. For example &#x27;2019-05-08T05:30:30.206Z&#x27; | [optional] [default to null]
**CreatedBy** | [***UserRef**](UserRef.md) |  | [optional] [default to null]
**ModifiedBy** | [***UserRef**](UserRef.md) |  | [optional] [default to null]
**Id** | **string** |  | [optional] [default to null]
**Name** | **string** | The display name of the tenant. | [default to null]
**DomainNames** | **[]string** | One or more domain names that this tenant uses in the URL | [default to null]
**Description** | **string** | Additional notes about the tenant. | [optional] [default to null]
**IsDisabled** | **bool** | Tenant&#x27;s disabled status. If a tenant disabled no operations can be performed to it. | [optional] [default to false]
**TenantTypeInput** | [***TenantTypeInput**](TenantTypeInput.md) |  | [optional] [default to null]
**TenantType** | **string** |  | [optional] [default to null]
**ParentId** | **string** |  | [optional] [default to null]
**AncestorTenants** | **[]string** | A list of all ancestor tenants that have access to this one. The list is sorted with the most immidiate ancesort, the parent tenant, being the first and the most distant, the sys tenant, being the last. | [optional] [default to null]
**RestApiEndPoint** | **string** | The REST Endpoint for this tenant, use this URL when requesting access to resources under this tenant. | [optional] [default to null]
**Labels** | **[]string** | List of labels associated with the tenant | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

