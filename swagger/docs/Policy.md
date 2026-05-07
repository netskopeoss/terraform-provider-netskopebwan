# Policy

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**DateCreated** | [**time.Time**](time.Time.md) | Time object record was created in ISO 8601 format. For example 2019-05-08T05:30:30.206Z | [optional] [default to null]
**DateModified** | [**time.Time**](time.Time.md) | Time object record was last modified in ISO 8601 format. For example &#x27;2019-05-08T05:30:30.206Z&#x27; | [optional] [default to null]
**CreatedBy** | [***UserRef**](UserRef.md) |  | [optional] [default to null]
**ModifiedBy** | [***UserRef**](UserRef.md) |  | [optional] [default to null]
**Id** | **string** |  | [optional] [default to null]
**Name** | **string** | The name of the policy | [optional] [default to null]
**Type_** | [***PolicyType**](PolicyType.md) |  | [optional] [default to null]
**AssignedEdges** | **[]string** |  | [optional] [default to null]
**Hubs** | [**[]EdgeRef**](EdgeRef.md) |  | [optional] [default to null]
**Config** | [***PolicyConfig**](PolicyConfig.md) |  | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

