# Terraform Provider for Netskope Borderless SD-WAN 
## Requirements

-	[Terraform](https://www.terraform.io/downloads.html) >= 1.3


## Using The Provider
The Netskope Borderless SD-WAN Terraform Provider Repo includes sample plans to get you started. 
Here are the following steps to be done before executing the terraform plans.

### Netskope Borderless SD-WAN Tenant Tasks

1. Identify the "Base URL" for your Netskope Borderless SD-WAN tenant.

2. Login to Tenant URL and and create a Token to use it as credentials.
    - Token can be created by navigating the page "Administration --> Token --> (+) Button"
    - Token Permissions will be like as follows (Its just an example):
	```
            [
              {
                "rap_resource": "",
                "rap_privs": [
                  "privSiteCreate",
                  "privSiteRead",
                  "privSiteWrite",
                  "privSiteDelete",
                  "privSiteToken",
                  "privSiteRestart",
                  "privSiteOpsRead",
                  "privSiteOpsWrite",
                  "privTokenRead",
                  "privPolicyCreate",
                  "privPolicyRead",
                  "privPolicyWrite",
                  "privPolicyDelete",
                  "privAppRead",
                  "privAuditRecordCreate"
                ]
              }
            ]

	```

### Terraform Configuration

1. Setup Required Providers in TF file. File is already included (examples/version.tf)
	```
        terraform {
          required_version = ">=0.0.1"
          required_providers {
            netskopebwan = {
              source  = "netskopeoss/netskopebwan"
              version = "0.0.1"
            }
          }
        }
	```

### Credentials Input

1. Identify the "Base API URL" for your Netskope Borderless SD-WAN tenant.
    - This will generally have a ".api." sub-domain in your tenant URL.
    - For example, if your tenant URL is "https://acme01.infiot.net", then the API URL will be "https://acme01.api.infiot.net"

2. There are two ways, you can feed the credentials.
   * Through Environment Variables
       - NS_SDWAN_MGMT_URL - Tenant API URL
       - NS_SDWAN_MGMT_TOKEN - Token created in the first step.

   * Configure it in the Provider Block (example file at examples/provider.tf)

### Execution

   ![Docs](docs/) Directory has detailed documentations for each resource / data source directive. 
