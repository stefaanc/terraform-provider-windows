## Resource: "windows_network_connection"

> :bulb:  
> This resource is automatically created by the windows-computer, and cannot be destroyed.
>   
> - Terraform's "Create" lifecycle-method imports the resource, saves the imported state so it can be reinstated at a later time, and updates the resource based on the attributes in the Terraform configuration. 
>  
> - Terraform's "Destroy" lifecycle-method reinstates the originally imported state. 

### Example Usage

```terraform
resource "windows_network_connection" "my_network_connection_1" {
    guid = "8598CC0D-18B8-4364-80D6-2F75C37ABCC0"

    connection_profile = "private"
}
output "my_network_connection_1_name" {
    value = windows_network_connection.my_network_connection_1.name
}
```

```terraform
resource "windows_network_connection" "my_network_connection_2" {
    ipv4_gateway_address = "192.168.2.1"
    allow_disconnect     = true
}
output "my_network_connection_2_name" {
    value = windows_network_connection.my_network_connection_2.name
}
output "my_network_connection_2_connectivity" {
    value = windows_network_connection.my_network_connection_2.ipv4_connectivity
}
```

```terraform
resource "windows_network_connection" "my_network_connection_3" {
    ipv6_gateway_address = "fe80::69c2:41ab:c6a1:1"
    allow_disconnect     = true
}
output "my_network_connection_3_name" {
    value = windows_network_connection.my_network_connection_3.name
}
output "my_network_connection_3_connectivity" {
    value = windows_network_connection.my_network_connection_3.ipv6_connectivity
}
```

```terraform
resource "windows_network_connection" "my_network_connection_4" {
    name             = "Network"
    allow_disconnect = true
}
output "my_network_connection_4_guid" {
    value = windows_network_connection.my_network_connection_4.guid
}
output "my_network_connection_4_ipv4_gateway" {
    value = windows_network_connection.my_network_connection_4.ipv4_gateway_address
}
output "my_network_connection_4_ipv6_gateway" {
    value = windows_network_connection.my_network_connection_4.ipv6_gateway_address
}
```

```terraform
resource "windows_network_connection" "my_network_connection_5" {
    old_name = "Network"
    new_name = "My-Network"
}
output "my_network_connection_5_name" {
    value = windows_network_connection.my_network_connection_5.name
}
```

```terraform
resource "windows_network_connection" "my_network_connection_6" {
    old_name = "Network 2"
    new_name = "My-Backup-Network"

    x_lifecycle {
        ignore_error_if_not_exists = true
    }
}
output "my_network_connection_6_exists" {
    value = windows_network_connection.my_network_connection_6.x_lifecycle[0].exists
}
```

<br/>

### Argument Attributes Reference

> :warning:  
> One of the identifying argument attributes is required.  Setting multiple identifying attributes will throw an error. 

- `guid` - (string, Optional, Identifying) -  The GUID of the network connection profile.  This GUID can be used to access the Windows registry.

- `ipv4_gateway_address` - (string, Optional, Identifying) -  The IPv4 gateway used for this connection.  If associated IP interfaces have multiple default gateways, you can set the `allow_disconnect` attribute to allow the provider to disconnect default gateways  one-by-one in order to find the gateway that belongs to this connection.

- `ipv6_gateway_address` - (string, Optional, Identifying) -  The IPv6 gateway used for this connection.  If associated IP interfaces have multiple default gateways, you can set the `allow_disconnect` attribute to allow the provider to disconnect default gateways  one-by-one in order to find the gateway that belongs to this connection.

- `name` - (string, Optional, Identifying) -  The name of the network for this connection.  This can be set, for instance based on the SSID of a WiFi network, otherwise the network will typically get a very generic name like `"Network"`.  The name of a connection can be changed using the `new_name` attribute.

- `old_name` - (string, Optional, Identifying) -  The old name of the network for this connection.  
When you also specify `new_name`, don't use `name` to identify the connection but use `old_name` instead, or use any of the alternative identifying attributes.  And for downstream interpolation, use `name` to avoid unexpected issues.

- `new_name` - (string, Optional) -  The new name of the network for this connection.  If the new name is different from the name, then the name will be changed.  
When specifying `new_name`, don't use `name` to identify the connection but use `old_name` instead, or use any of the alternative identifying attributes.  And for downstream interpolation, use `name` to avoid unexpected issues.

- `allow_disconnect` - (boolean, Optional, defaults to `false`) -  The provider will attempt to find the gateway belonging to this connection by looking for the default gateway of the IP interfaces that are associated to this connection.  If these associated IP interfaces have multiple default gateways, the provider has no way to find out which of their default gateways belong to this network-connection.  In this case, you can set the `allow_disconnect` attribute to allow the provider to disconnect/reconnect default gateways one-by-one in order to find the one that belongs to this connection.

- `connection_profile` - (string, Optional) -  The profile of this connection.  Accepted values are `"public"` or `"private"`.  The computer automatically sets the value `"DomainAuthenticated"` when the network is authenticated to a domain controller.  The value of this attribute is used by the firewall.

- `x_lifecycle` - (resource, Optional)

  - `ignore_error_if_not_exists` - (boolean, Optional, defaults to `false`) -  If the resource doesn't exist, the Terraform state contains zeroed attributes for this resource.  No error is thrown.

<br/>

### Exported Attributes Reference

```json
{
    "guid":                 "C42B1E6D-0856-4932-B06C-3085DA1B1978",

    "ipv4_gateway_address": "192.168.2.1",
    "ipv6_gateway_address": "fe80::69c2:41ab:c6a1:1",

    "name":                 "My-Network",
    "old_name":             "",
    "new_name":             "",

    "allow_disconnect":     true,

    "connection_profile":   "Private",

    "ipv4_connectivity":    "Internet",
    "ipv6_connectivity":    "LocalNetwork",

    "x_lifecycle": [{
        "ignore_error_if_not_exists": true,
        "exists":                     true
    }]      
}
```

In addition to the argument attributes:

- `ipv4_connectivity` - (string) -  The kind of connectivity this network connection provides when using IPv4.

- `ipv6_connectivity` - (string) -  The kind of connectivity this network connection provides when using IPv6.

- `x_lifecycle` - (resource)

  - `exists` - (boolean) -  The resource exists, and the Terraform state contains the attributes of the resource.

<br/>

### API Mapping

To help with debugging, the following provides an overview of where the attributes can be found, using shell commands.

###### Mapping of attributes on Powershell

attribute                             | command
:-------------------------------------|:------------
`guid`                                | `( Get-ChildItem -Path 'HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\NetworkList\Profiles' ).PSChildName.Trim("{}")`
`ipv4_gateway_address`                | function(`( Get-NetRoute -InterfaceIndex $( Get-NetConnectionProfile ).InterfaceIndex ).NextHop`)
`ipv6_gateway_address`                | function(`( Get-NetRoute -InterfaceIndex $( Get-NetConnectionProfile ).InterfaceIndex ).NextHop`)
`name`                                | `( Get-NetConnectionProfile ).Name`
`old_name`                            | not mapped
`new_name`                            | not mapped
`allow_disconnect`                    | not mapped
`connection_profile`                  | `( Get-NetConnectionProfile ).NetworkCategory`
`ipv4_connectivity`                   | `( Get-NetConnectionProfile ).IPv4Connectivity`
`ipv6_connectivity`                   | `( Get-NetConnectionProfile ).IPv6Connectivity`

> Remark:  
> A network connection is defined for every individual gateway in the routing table.
> 
> As far as we can tell, there is no direct way to find the gateway associated to a network connection, using a shell command.
> Network connections can be associated to multiple network adapters, and network adapters can have multiple default gateways.  For certain configurations, it can be necessary to delete each of the gateways one-by-one, while monitoring which network connection disappears, in order to find the gateway that is associated to a network connection.

<br/>
