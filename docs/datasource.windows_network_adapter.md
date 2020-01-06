## Data Source: "windows_network_adapter"

### Example Usage

```terraform
data "windows_network_adapter" "my_network_adapter_A" {
    guid = "C42B1E6D-0856-4932-B06C-3085DA1B1978"
}
output "my_network_adapter_A_name" {
    value = data.windows_network_adapter.my_network_adapter_A.name
}
output "my_network_adapter_A_mac_address" {
    value = data.windows_network_adapter.my_network_adapter_A.mac_address
}
output "my_network_adapter_A_admin_status" {
    value = data.windows_network_adapter.my_network_adapter_A.admin_status
}
output "my_network_adapter_A_operational_status" {
    value = data.windows_network_adapter.my_network_adapter_A.operational_status
}
output "my_network_adapter_A_connection_status" {
    value = data.windows_network_adapter.my_network_adapter_A.connection_status
}
output "my_network_adapter_A_connection_speed" {
    value = data.windows_network_adapter.my_network_adapter_A.connection_speed
}
output "my_network_adapter_A_is_physical" {
    value = data.windows_network_adapter.my_network_adapter_A.is_physical
}
```

```terraform
data "windows_network_adapter" "my_network_adapter_B" {
    name = "Ethernet"
}
output "my_network_adapter_B_guid" {
    value = data.windows_network_adapter.my_network_adapter_B.guid
}
```

```terraform
data "windows_network_adapter" "my_network_adapter_C" {
    name = "Ethernet"

    x_lifecycle {
        ignore_error_if_not_exists = true
    }
}
output "my_network_adapter_C_exists" {
    value = data.windows_network_adapter.my_network_adapter_C.x_lifecycle[0].exists
}
```

<br/>

### Argument Attributes Reference

> :warning:  
> One of the identifying argument attributes is required.  Setting multiple identifying attributes will throw an error. 

- `guid` - (string, Optional, Identifying) -  The GUID of the network connection profile.  This GUID can be used to access the Windows registry.
 
- `name` - (string, Optional, Identifying) -  The name of the network adapter.

- `x_lifecycle` - (resource, Optional)

  - `ignore_error_if_not_exists` - (boolean, Optional, defaults to `false`) -  If the resource doesn't exist, the Terraform state contains zeroed attributes for this resource.  No error is thrown.

<br/>

### Exported Attributes Reference

```json
{
    "guid":                  "E34DC156-C49F-42DB-A6F6-D5609648D274"

    "name":                  "Staging"

    "mac_address":           "90-B1-1C-63-D3-82"
    "permanent_mac_address": "90-B1-1C-63-D3-82"

    "dns_client": [{
        "register_connection_address": true
        "register_connection_suffix":  "staging.local"
    }]

    "admin_status":          "Up"
    "operational_status":    "Up"
    "connection_status":     "Connected"
    "connection_speed":      "100 Mbps"macAddress
    "is_physical":           true

    "x_lifecycle": [{
        "ignore_error_if_not_exists": true
        "exists":                     true
    }]      
}
```

- `mac_address` - (string) -  The active MAC address of the network adapter.  This is the MAC address that is being used.

  > :warning:  
  > Changing a MAC address does cause a short disconnection from the network.
  
- `permanent_mac_address` - (string) -  The permanent MAC address of the network adapter.  This is the MAC address that is "hard-coded" for the network adapter.  This address is not used when `mac_address` was changed to a different value.

- `dns_client` - (resource)

  - `register_connection_address` - (boolean) -  Indicates whether the IP address for this connection is to be registered by the DNS client.

  - `register_connection_suffix` - (string) -  Specifies the connection-specific suffixes to append. This attribute value is a per-connection DNS suffix to append to the computer name to construct a Fully Qualified Domain Name (FQDN). This FQDN is used as the host name for name resolution by the DNS client.  

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
`permanent_mac_address`               | `( Get-NetAdapter ).PermanentAddress`
`dns_client`                          | &nbsp;
 -&nbsp;`register_connection_address` | `( Get-DNSClient ).RegisterThisConnectionsAddress`
 -&nbsp;`register_connection_suffix`  | `if ( ( Get-DNSClient ).UseSuffixWhenRegistering ) { ( Get-DNSClient ).ConnectionSpecificSuffix } else { "" }`
`admin_status`                        | `( Get-NetAdapter ).AdminStatus`
`operational_status`                  | `( Get-NetAdapter ).ifOperStatus`
`connection_status`                   | `( Get-NetAdapter ).MediaConnectionState`
`connection_speed`                    | `( Get-NetAdapter ).LinkSpeed`
`is_physical`                         | `( Get-NetAdapter ).ConnectorPresent`

<br/>

