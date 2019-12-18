## Resource: "windows_network_adapter"

> :bulb:  
> This resource is automatically created by the windows-computer, and cannot be destroyed.  
> 
> - Terraform's "Create" lifecycle-method imports the resource, saves the imported state so it can be reinstated at a later time, and updates the resource based on the attributes in the Terraform configuration. 
>  
> - Terraform's "Destroy" lifecycle-method reinstates the originally imported state. 

### Example Usage

<br/>

### Argument Attributes Reference

- `name` - (string, Required) -  

- `mac_address` - (string, Optional) -  

- `ipv4` - (resource, Optional)  

  - `interface_metric` - (integer, Optional) -  

  - `dhcp_enabled` - (boolean, Optional) -  

  - `ip_address` - (set[resource], Optional)
 
    - `address` - (string, Optional) -  

    - `prefix_length` - (integer, Optional) -  

    - `skip_as_source` - (boolean, Optional) -

  - `gateway` - (set[resource], Optional)

    - `address` - (string, Optional) -  

    - `route_metric` - (integer, Optional) -  

  - `dns_addresses` - (list[string], Optional)

- `ipv6` - (resource, Optional)  

  - `interface_metric` - (integer, Optional) -  

  - `dhcp_enabled` - (boolean, Optional) -  

  - `ip_address` - (resource, Optional)
 
    - `address` - (string, Optional) -  

    - `prefix_length` - (integer, Optional) -  

    - `skip_as_source` - (boolean, Optional) -  

- `dns_client` - (resource, Optional)  

  - `register_connection_address` - (boolean, Optional) -  

  - `register_connection_suffix` - (string, Optional) -  

- `x_lifecycle` - (resource, Optional)

  - `ignore_error_if_not_exists` - (boolean, Optional, Computed) -  

<br/>

### Exported Attributes Reference

```json
{
    
}
```

- `guid` - (string, Required) -  

- `ipv4` - (resource)  

  - `interface_metric_obtained_automatically` - (boolean) -  

  - `ip_address_obtained_through_automatically` - (boolean) -  

  - `gateway` - (resource)

    - `route_metric_obtained_automatically` - (boolean) -  

  - `gateway_obtained_automatically` - (boolean) -  

  - `dns_addresses_obtained_automatically` - (boolean) -  

  - `connection_status` - (string) -  

  - `connectivity` - (string) -  

- `ipv6` - (resource)  

  - `interface_metric_obtained_automatically` - (boolean) -  

  - `ip_address_obtained_through_automatically` - (boolean) -  

  - `gateway` - (resource)

    - `route_metric_obtained_automatically` - (boolean) -  

  - `gateway_obtained_automatically` - (boolean) -  

  - `dns_addresses_obtained_automatically` - (boolean) -  

  - `connection_status` - (string) -  

  - `connectivity` - (string) -  

- `dns` - (resource)  

- `admin_status` - (string) -  

- `operational_status` - (string) -  

- `connection_status` - (string) -  

- `connection_speed` - (string) -  

- `is_physical` - (boolean) -  

- `x_lifecycle` - (resource)

  - `exists` - (boolean) -  

<br/>
