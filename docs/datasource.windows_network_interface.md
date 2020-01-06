## Data Source: "windows_network_interface"

### Example Usage

```terraform
data "windows_network_interface" "my_network_interface_A" {
    guid = "E34DC156-C49F-42DB-A6F6-D5609648D274"
}
output "my_network_interface_A_index" {
    value = data.windows_network_interface.my_network_interface_A.index
}
output "my_network_interface_A_network_connections" {
    value = data.windows_network_interface.my_network_interface_A.network_connection_names
}
output "my_network_interface_A_computer" {
    value = data.windows_network_interface.my_network_interface_A.computer_name
}
```

```terraform
data "windows_network_interface" "my_network_interface_B" {
    index = 7
}
output "my_network_interface_B_index" {
    value = data.windows_network_interface.my_network_interface_B.index
}
```

```terraform
data "windows_network_interface" "my_network_interface_C" {
    alias = "Ethernet"
}
output "my_network_interface_C_index" {
    value = data.windows_network_interface.my_network_interface_C.index
}
```

```terraform
data "windows_network_interface" "my_network_interface_D" {
    description = "Intel(R) 82579LM Gigabit Network Connection"
}
output "my_network_interface_D_index" {
    value = data.windows_network_interface.my_network_interface_D.index
}
```

```terraform
data "windows_network_interface" "my_network_interface_E" {
    mac_address = "90-B1-1C-63-D3-82"
}
output "my_network_interface_E_index" {
    value = data.windows_network_interface.my_network_interface_E.index
}
```

```terraform
data "windows_network_interface" "my_network_interface_F" {
    network_adapter_name = "Ethernet"
}
output "my_network_interface_F_index" {
    value = data.windows_network_interface.my_network_interface_F.index
}
```

```terraform
data "windows_network_interface" "my_network_interface_G" {
    vnetwork_adapter_name = "External Switch"
}
output "my_network_interface_G_index" {
    value = data.windows_network_interface.my_network_interface_G.index
}
output "my_network_interface_G_vswitch" {
    value = data.windows_network_interface.my_network_interface_G.vswitch_name
}
```

```terraform
data "windows_network_interface" "my_network_interface_H" {
    alias = "vEthernet (Default Switch)"

    x_lifecycle {
        ignore_error_if_not_exists = true
    }
}
output "my_network_interface_H_exists" {
    value = data.windows_network_interface.my_network_interface_H.x_lifecycle[0].exists
}
```

<br/>

### Argument Attributes Reference

> :warning:  
> One of the identifying arguments is required.  Setting multiple will throw an error.. 

- `index` - (integer, Optional, Identifying) -  The index of the network interface.  One and only one of the identifying arguments `index`, `alias`, `description`, `guid`, `mac_address`, `network_adapter_name` or `vnetwork_adapter_name` must be specified.

- `alias` - (string, Optional, Identifying) -  The alias of the network interface.  This is typically the same as the `network_adapter_name`.

- `description` - (string, Optional, Identifying) -  The description of the network interface.  This is typically a string that identifies the chip-set used by the network adapter, like `"Intel(R) 82579LM Gigabit Network Connection"`.  If the network adapter is associated to a virtual network adapter (and thus doesn't have a chip-set), then this is typically starting with `"Hyper-V Virtual Ethernet Adapter"`.

- `guid` - (string, Optional, Identifying) -  The GUID of the network interface.  This GUID can be used to access the Windows registry.

- `mac_address` - (string, Optional, Identifying) -  The MAC address of the associated network adapter.  If the network adapter is associated to a virtual network adapter, then this is typically the same as the MAC address of the virtual network adapter.  
Remark that it is possible for two interfaces to have the same MAC address, typically when using the Hyper-V hypervisor: the network adapter of an external or internal switch and the network adapter of the management OS may have the same MAC.  When trying to identify a network interface using a MAC address and when there are multiple network interfaces with that same MAC address, the provider will throw an error.

- `network_adapter_name` - (string, Optional, Identifying) -  The name of the network adapter.  This is typically the same as the `alias` of the network interface.  Also, if the network adapter is associated to a virtual network adapter, then this is typically of the form `"vEthernet ($vna)"` where `$vna` is the name of the virtual network adapter. 

- `vnetwork_adapter_name` - (string, Optional, Identifying) -  The name of the virtual network adapter that is associated to the network interface.  This attribute can only be used when the Hyper-V hypervisor is running on this windows-computer.

- `x_lifecycle` - (resource, Optional)

  - `ignore_error_if_not_exists` - (boolean, Optional, defaults to `false`) -  If the resource doesn't exist, the Terraform state contains zeroed attributes for this resource.  No error is thrown.

<br/>

### Exported Attributes Reference

```json
{
    "guid":                     "E34DC156-C49F-42DB-A6F6-D5609648D274",
    "index":                    7,
    "alias":                    "Ethernet",
    "description":              "Intel(R) 82579LM Gigabit Network Connection",
    "mac_address":              "90-B1-1C-63-D3-82",
    "network_adapter_name":     "Ethernet",
    "vnetwork_adapter_name":    "",

    "network_connection_names": [ "Network", "WiFi" ],
    "vswitch_name":             "",
    "computer_name":            "MY-COMPUTER",

    "x-lifecycle": [{
        "ignore_error_if_not_exists": false,
        "exists":                     false
    }]
}
```

In addition to the argument attributes:

- `network_connection_names` - (list[string]) -  The names of the network connections via the default gateways for this network interface.
 
- `vswitch_name` - (string, Optional) -  The name of the virtual switch that is associated to the network interface.  This attribute is only provided when the Hyper-V hypervisor is running on this windows-computer (`vswitch_name = ""` if there is no virtual switch associated to this network interface).

- `computer_name` - (string) -  The name of the windows-computer.

- `x_lifecycle` - (resource)

  - `exists` - (boolean) -  The resource exists, and the Terraform state contains the attributes of the resource.

<br/>

### API Mapping

To help with debugging, the following provides an overview of where the attributes can be found, using shell commands.

###### Mapping of attributes on Powershell

attribute                  | command
:--------------------------|:------------
`guid`                     | `( Get-NetAdapter ).InterfaceGUID.Trim("{}")`
`index`                    | `( Get-NetAdapter ).InterfaceIndex`    
`alias`                    | `( Get-NetAdapter ).InterfaceAlias`
`description`              | `( Get-NetAdapter ).InterfaceDescription`
`mac_address`              | `( Get-NetAdapter ).MacAddress`
`network_adapter_name`     | `( Get-NetAdapter ).Name`
`vnetwork_adapter_name`    | `( Get-VMNetworkAdapter -ManagementOS ).Name`
`network_connection_names` | `( Get-NetConnectionProfile ).Name`
`vswitch_name`             | `( Get-VMNetworkAdapter -ManagementOS ).SwitchName`
`computer_name`            | `( Get-NetAdapter ).SystemName`

<br/>
