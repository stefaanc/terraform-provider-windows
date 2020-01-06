## Data Source: "windows_network_connection"

### Example Usage

```terraform
data "windows_network_connection" "my_network_connection_A" {
    guid = "8598CC0D-18B8-4364-80D6-2F75C37ABCC0"
}
output "my_network_connection_A_name" {
    value = data.windows_network_connection.my_network_connection_A.name
}
```

```terraform
data "windows_network_connection" "my_network_connection_B" {
    ipv4_gateway_address = "192.168.2.1"
    allow_disconnect     = true
}
output "my_network_connection_B_name" {
    value = data.windows_network_connection.my_network_connection_B.name
}
output "my_network_connection_B_connectivity" {
    value = data.windows_network_connection.my_network_connection_B.ipv4_connectivity
}
```

```terraform
data "windows_network_connection" "my_network_connection_C" {
    ipv6_gateway_address = "fe80::69c2:41ab:c6a1:1"
    allow_disconnect     = true
}
output "my_network_connection_C_name" {
    value = data.windows_network_connection.my_network_connection_C.name
}
output "my_network_connection_C_connectivity" {
    value = data.windows_network_connection.my_network_connection_C.ipv6_connectivity
}
```

```terraform
data "windows_network_connection" "my_network_connection_D" {
    name             = "Network"
    allow_disconnect = true
}
output "my_network_connection_D_guid" {
    value = data.windows_network_connection.my_network_connection_D.guid
}
output "my_network_connection_D_ipv4_gateway" {
    value = data.windows_network_connection.my_network_connection_D.ipv4_gateway_address
}
output "my_network_connection_D_ipv6_gateway" {
    value = data.windows_network_connection.my_network_connection_D.ipv6_gateway_address
}
```

```terraform
data "windows_network_connection" "my_network_connection_E" {
    name = "Network 2"

    x_lifecycle {
        ignore_error_if_not_exists = true
    }
}
output "my_network_connection_E_exists" {
    value = data.windows_network_connection.my_network_connection_E.x_lifecycle[0].exists
}
```

<br/>

### Argument Attributes Reference

> :warning:  
> One of the identifying argument attributes is required.  Setting multiple identifying attributes will throw an error. 

- `guid` - (string, Optional, Identifying) -  The GUID of the network connection profile.  This GUID can be used to access the Windows registry.

- `ipv4_gateway_address` - (string, Optional, Identifying) -  The IPv4 gateway used for this connection.  If associated IP interfaces have multiple default gateways, you can set the `allow_disconnect` attribute to allow the provider to disconnect default gateways  one-by-one in order to find the gateway that belongs to this connection.

- `ipv6_gateway_address` - (string, Optional, Identifying) -  The IPv6 gateway used for this connection.  If associated IP interfaces have multiple default gateways, you can set the `allow_disconnect` attribute to allow the provider to disconnect default gateways  one-by-one in order to find the gateway that belongs to this connection.

- `name` - (string, Optional, Identifying) -  The name of the network for this connection.  This can be set, for instance based on the SSID of a WiFi network, otherwise the network will typically get a very generic name like `"Network"`.

- `allow_disconnect` - (boolean, Optional, defaults to `false`) -  The provider will attempt to find the gateway belonging to this connection by looking for the default gateway of the IP interfaces that are associated to this connection.  If these associated IP interfaces have multiple default gateways, the provider has no way to find out which of their default gateways belong to this network-connection.  In this case, you can set the `allow_disconnect` attribute to allow the provider to disconnect/reconnect default gateways one-by-one in order to find the one that belongs to this connection.

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

    "allow_disconnect":     true,

    "connection_profile":   "Private",

    "ipv4_connectivity":    "Internet",
    "ipv6_connectivity":    "LocalNetwork",

    "network_adapter_names": [ "Ethernet" ],

    "x_lifecycle": [{
        "ignore_error_if_not_exists": true,
        "exists":                     true
    }]      
}
```

In addition to the argument attributes:

- `connection_profile` - (string) -  The profile of this connection.  Values are `"Public"`, `"Private"`, or `"DomainAuthenticated"` when the network is authenticated to a domain controller.  The value of this attribute is used by the firewall.

- `ipv4_connectivity` - (string) -  The kind of connectivity this network connection provides when using IPv4.

- `ipv6_connectivity` - (string) -  The kind of connectivity this network connection provides when using IPv6.

- `network_adapter_names` - (list[string]) -  The names of the network adapters associated to this network connection.

- `x_lifecycle` - (resource)

  - `exists` - (boolean) -  The resource exists, and the Terraform state contains the attributes of the resource.

<br/>

### API Mapping

To help with debugging, the following provides an overview of where the attributes can be found, using shell commands.

###### Mapping of attributes on Powershell

attribute                             | command
:-------------------------------------|:------------
`guid`                                | `( Get-ChildItem -Path 'HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\NetworkList\Profiles' ).PSChildName.Trim("{}")`
`ipv4_gateway_address`                | function(`( Get-NetConnectionProfile ).InterfaceIndex) \| ( Get-NetRoute -InterfaceIndex $_ ).NextHop`)
`ipv6_gateway_address`                | function(`( Get-NetConnectionProfile ).InterfaceIndex \| ( Get-NetRoute  -InterfaceIndex $_ ).NextHop`)
`name`                                | `( Get-NetConnectionProfile ).Name`
`old_name`                            | not mapped
`new_name`                            | not mapped
`allow_disconnect`                    | not mapped
`connection_profile`                  | `( Get-NetConnectionProfile ).NetworkCategory`
`ipv4_connectivity`                   | `( Get-NetConnectionProfile ).IPv4Connectivity`
`ipv6_connectivity`                   | `( Get-NetConnectionProfile ).IPv6Connectivity`
`network_adapter_names`               | `( Get-NetConnectionProfile ).InterfaceAlias`

> Remark:  
> A network connection is defined for every individual gateway in the routing table.
> 
> As far as we can tell, there is no direct way to find the gateway associated to a network connection, using a shell command.
> Network connections can be associated to multiple network adapters, and network adapters can have multiple default gateways.  For certain configurations, it can be necessary to delete each of the gateways one-by-one, while monitoring which network connection disappears, in order to find the gateway that is associated to a network connection.

<br/>
