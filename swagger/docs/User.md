# User

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**DateCreated** | [**time.Time**](time.Time.md) | Time object record was created in ISO 8601 format. For example 2019-05-08T05:30:30.206Z | [optional] [default to null]
**DateModified** | [**time.Time**](time.Time.md) | Time object record was last modified in ISO 8601 format. For example &#x27;2019-05-08T05:30:30.206Z&#x27; | [optional] [default to null]
**CreatedBy** | [***UserRef**](UserRef.md) |  | [optional] [default to null]
**ModifiedBy** | [***UserRef**](UserRef.md) |  | [optional] [default to null]
**Id** | **string** | A unique ID assigned to the user. | [optional] [default to null]
**Name** | **string** | Usernames may contain lowercase latin characters, numbers, dots, or underscores | [default to null]
**Email** | **string** | Email ID of the user | [default to null]
**Roles** | [**[]UserRole**](UserRole.md) | User roles. | [default to null]
**IsDisabled** | **bool** | If true user is disabled | [optional] [default to false]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

