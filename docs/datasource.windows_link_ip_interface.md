## Data Source: "windows_link_ip_interface"

### Example Usage

```terraform
data "windows_link_ip_interface" "my_ip_interface_1" {
    index = 7
}
output "my_ip_interface_1_index" {
    value = data.windows_link_ip_interface.my_ip_interface_1.index
}
output "my_ip_interface_1_network_connections" {
    value = data.windows_link_ip_interface.my_ip_interface_1.network_connection_names
}
output "my_ip_interface_1_computer" {
    value = data.windows_link_ip_interface.my_ip_interface_1.computer_name
}
```

```terraform
data "windows_link_ip_interface" "my_ip_interface_2" {
    alias = "Ethernet"
}
output "my_ip_interface_2_index" {
    value = data.windows_link_ip_interface.my_ip_interface_2.index
}
```

```terraform
data "windows_link_ip_interface" "my_ip_interface_3" {
    description = "Intel(R) 82579LM Gigabit Network Connection"
}
output "my_ip_interface_3_index" {
    value = data.windows_link_ip_interface.my_ip_interface_3.index
}
```

```terraform
data "windows_link_ip_interface" "my_ip_interface_4" {
    guid = "E34DC156-C49F-42DB-A6F6-D5609648D274"
}
output "my_ip_interface_4_index" {
    value = data.windows_link_ip_interface.my_ip_interface_4.index
}
```

```terraform
data "windows_link_ip_interface" "my_ip_interface_5" {
    mac_address = "90-B1-1C-63-D3-82"
}
output "my_ip_interface_5_index" {
    value = data.windows_link_ip_interface.my_ip_interface_5.index
}
```

```terraform
data "windows_link_ip_interface" "my_ip_interface_6" {
    network_adapter_name = "Ethernet"
}
output "my_ip_interface_6_index" {
    value = data.windows_link_ip_interface.my_ip_interface_6.index
}
```

```terraform
data "windows_link_ip_interface" "my_ip_interface_7" {
    vnetwork_adapter_name = "External Switch"
}
output "my_ip_interface_7_index" {
    value = data.windows_link_ip_interface.my_ip_interface_7.index
}
output "my_ip_interface_7_vswitch" {
    value = data.windows_link_ip_interface.my_ip_interface_7.vswitch_name
}
```

```terraform
data "windows_link_ip_interface" "my_ip_interface_8" {
    vnetwork_adapter_name = "Default Switch"

    x_lifecycle {
        ignore_error_if_not_exists = true
    }
}
output "my_ip_interface_8_index" {
    value = data.windows_link_ip_interface.my_ip_interface_8.index
}
output "my_ip_interface_8_exists" {
    value = data.windows_link_ip_interface.my_ip_interface_8.x_lifecycle[0].exists
}
```

<br/>

### Argument Attributes Reference

> :warning:  
> One of the identifying arguments is required.  Setting multiple will throw an error.. 

- `index` - (integer, Optional, Identifying) -  The index of the IP interface.  One and only one of the identifying arguments `index`, `alias`, `description`, `guid`, `mac_address`, `network_adapter_name` or `vnetwork_adapter_name` must be specified.

- `alias` - (string, Optional, Identifying) -  The alias of the IP interface.  This is typically the same as the `network_adapter_name`.

- `description` - (string, Optional, Identifying) -  The description of the IP interface.  This is typically a string that identifies the chip-set used by the network adapter, like `"Intel(R) 82579LM Gigabit Network Connection"`.  If the network adapter is associated to a virtual network adapter (and thus doesn't have a chip-set), then this is typically starting with `"Hyper-V Virtual Ethernet Adapter"`.

- `guid` - (string, Optional, Identifying) -  The GUID of the IP interface/network adapter.  This GUID can be used to access the Windows registry.

- `mac_address` - (string, Optional, Identifying) -  The MAC address of the network adapter.  If the network adapter is associated to a virtual network adapter, then this is typically the same as the MAC address of the virtual network adapter.

- `network_adapter_name` - (string, Optional, Identifying) -  The name of the network adapter.  This is typically the same as the `alias` of the IP interface.  Also, if the network adapter is associated to a virtual network adapter, then this is typically of the form `"vEthernet ($vna)"` where `$vna` is the name of the virtual network adapter. 

- `vnetwork_adapter_name` - (string, Optional, Identifying) -  The name of the virtual network adapter that is associated to the IP interface.  This attribute can only be used when the Hyper-V hypervisor is running on this windows-computer.

- `x_lifecycle` - (resource, Optional)

  - `ignore_error_if_not_exists` - (boolean, Optional, defaults to `false`) -  If the resource doesn't exist, the Terraform state contains zeroed attributes for this resource.  No error is thrown.

<br/>

### Exported Attributes Reference

```json
{
    "index":                    7,
    "alias":                    "Ethernet",
    "description":              "Intel(R) 82579LM Gigabit Network Connection",
    "guid":                     "E34DC156-C49F-42DB-A6F6-D5609648D274",
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

All argument attributes are set to the value of the existing resources (`vnetwork_adapter_name = ""` if there is no virtual network adapter associated to this IP interface).

In addition to the argument attributes:

- `network_connection_names` - (list[string]) -  The names of the network connections made via the default gateways for this IP interface.
 
- `vswitch_name` - (string, Optional) -  The name of the virtual switch that is associated to the IP interface.  This attribute is only provided when the Hyper-V hypervisor is running on this windows-computer (`vswitch_name = ""` if there is no virtual switch associated to this IP interface).

- `computer_name` - (string) -  The name of the windows-computer.

- `x_lifecycle` - (resource)

  - `exists` - (boolean) -  The resource exists, and the Terraform state contains the attributes of the resource.

<br/>
