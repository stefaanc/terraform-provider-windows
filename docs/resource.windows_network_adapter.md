## Resource: "windows_network_adapter"

> :bulb:  
> This resource is automatically created by the windows-computer, and cannot be destroyed.  
> 
> - Terraform's "Create" lifecycle-method imports the resource, saves the imported state so it can be reinstated at a later time, and updates the resource based on the attributes in the Terraform configuration. 
>  
> - Terraform's "Destroy" lifecycle-method reinstates the originally imported state. 

### Example Usage

```terraform
resource "windows_network_adapter" "my_network_adapter_1" {
    guid = "C42B1E6D-0856-4932-B06C-3085DA1B1978"

    mac_address = "00-10-18-57-1B-0D"

    dns_client {
        register_connection_address = true
        register_connection_suffix  = "staging.local"
    }
}
output "my_network_adapter_1_name" {
    value = windows_network_adapter.my_network_adapter_1.name
}
output "my_network_adapter_1_admin_status" {
    value = windows_network_adapter.my_network_adapter_1.admin_status
}
output "my_network_adapter_1_operational_status" {
    value = windows_network_adapter.my_network_adapter_1.operational_status
}
output "my_network_adapter_1_connection_status" {
    value = windows_network_adapter.my_network_adapter_1.connection_status
}
output "my_network_adapter_1_connection_speed" {
    value = windows_network_adapter.my_network_adapter_1.connection_speed
}
output "my_network_adapter_1_is_physical" {
    value = windows_network_adapter.my_network_adapter_1.is_physical
}
```

```terraform
resource "windows_network_adapter" "my_network_adapter_2" {
    name = "Ethernet"
}
output "my_network_adapter_2_guid" {
    value = windows_network_adapter.my_network_adapter_2.guid
}
```

```terraform
resource "windows_network_adapter" "my_network_adapter_3" {
    old_name = "Ethernet"
    new_name = "Staging"
}
output "my_network_adapter_3_name" {
    value = windows_network_adapter.my_network_adapter_3.name
}
```

```terraform
resource "windows_network_adapter" "my_network_adapter_4" {
    name = "Staging"

    x_lifecycle {
        ignore_error_if_not_exists = true
    }
}
output "my_network_adapter_4_exists" {
    value = windows_network_adapter.my_network_adapter_4.x_lifecycle[0].exists
}
```

<br/>

### Argument Attributes Reference

> :warning:  
> One of the identifying argument attributes is required.  Setting multiple identifying attributes will throw an error. 

- `guid` - (string, Optional, Identifying) -  The GUID of the network connection profile.  This GUID can be used to access the Windows registry.
 
- `name` - (string, Optional, Identifying) -  The name of the network adapter.  The name of the network adapter can be changed using the `new_name` attribute.

- `old_name` - (string, Optional, Identifying) -  The old name of the network adapter.  
When you also specify `new_name`, don't use `name` to identify the adapter but use `old_name` instead, or use any of the alternative identifying attributes.  And for downstream interpolation, use `name` to avoid unexpected issues.

- `new_name` - (string, Optional) -  The new name of the network adapter.  If the new name is different from the name, then the name will be changed.  
When specifying `new_name`, don't use `name` to identify the adapter but use `old_name` instead, or use any of the alternative identifying attributes.  And for downstream interpolation, use `name` to avoid unexpected issues.

- `mac_address` - (string, Optional) -  The MAC address of the network adapter.  

- `dns_client` - (resource, Optional) -  When the network adapter is not a DNS client, i.e. when it doesn't have an IP interface, these attributes throw an error when in config.

  - `register_connection_address` - (boolean, Optional) -  Indicates whether the IP address for this connection is to be registered by the DNS client.

  - `register_connection_suffix` - (string, Optional) -  Specifies the connection-specific suffixes to append. This attribute value is a per-connection DNS suffix to append to the computer name to construct a Fully Qualified Domain Name (FQDN). This FQDN is used as the host name for name resolution by the DNS client.

- `x_lifecycle` - (resource, Optional)

  - `ignore_error_if_not_exists` - (boolean, Optional, defaults to `false`) -  If the resource doesn't exist, the Terraform state contains zeroed attributes for this resource.  No error is thrown.

<br/>

### Exported Attributes Reference

```json
{
    "guid":               "E34DC156-C49F-42DB-A6F6-D5609648D274"

    "name":               "Staging"
    "old_name":           ""
    "new_name":           ""

    "mac_address":        "90-B1-1C-63-D3-82"

    "dns_client": [{
        "register_connection_address": true
        "register_connection_suffix":  "staging.local"
    }]

    "admin_status":       "Up"
    "operational_status": "Up"
    "connection_status":  "Connected"
    "connection_speed":   "100 Mbps"macAddress
    "is_physical":        true

    "x_lifecycle": [{
        "ignore_error_if_not_exists": true
        "exists":                     true
    }]      
}
```

- `admin_status` - (string) -  The administrative status of the network adapter.

- `operational_status` - (string) -  The operational status of the network adapter.  

- `connection_status` - (string) -  The status of the network adapter's connection.  

- `connection_speed` - (string) -  The speed of the network adapter's connection.

- `is_physical` - (boolean) -  Is the network adapter associated with a physical NIC?

- `x_lifecycle` - (resource)

  - `exists` - (boolean) -  The resource exists, and the Terraform state contains the attributes of the resource.

<br/>

### API Mapping

To help with debugging, the following provides an overview of where the attributes can be found, using shell commands.

###### Mapping of attributes on Powershell

attribute                             | command
:-------------------------------------|:------------
`guid`                                | `( Get-NetAdapter ).InstanceID.Trim("{}")`
`name`                                | `( Get-NetAdapter ).Name`
`old_name`                            | not mapped
`new_name`                            | not mapped
`mac_address`                         | `( Get-NetAdapter ).MacAddress`
`dns_client`                          | &nbsp;
 -&nbsp;`register_connection_address` | `( Get-DNSClient ).RegisterThisConnectionsAddress`
 -&nbsp;`register_connection_suffix`  | `if ( ( Get-DNSClient ).UseSuffixWhenRegistering ) { ( Get-DNSClient ).ConnectionSpecificSuffix } else { "" }`
`admin_status`                        | `( Get-NetAdapter ).AdminStatus`
`operational_status`                  | `( Get-NetAdapter ).ifOperStatus`
`connection_status`                   | `( Get-NetAdapter ).MediaConnectionState`
`connection_speed`                    | `( Get-NetAdapter ).LinkSpeed`
`is_physical`                         | `( Get-NetAdapter ).ConnectorPresent`

<br/>

