# Terraform Provider Windows

**A terraform provider to work with a Windows computer**

The Windows provider is used to interact with the resources supported by a Windows computer.  
The provider has been developed and tested on a Windows 10 machine, but may be used with other versions of Windows with some limitations.



<br/>

# !!! UNDER CONSTRUCTION !!!!!!!!



<br/>

## Prerequisites

To build:
- [GNU make](https://www.gnu.org/software/make/)
- [Golang](https://golang.org/) >= v1.13
- [Terraform plugin SDK](https://github.com/hashicorp/terraform-plugin-sdk) ~= v1.4.0

To use:
- [Terraform](https://terraform.io) >= v0.12.17



<br>

## Building The Provider

1. Clone the git-repository on your machine

   ```shell
   mkdir -p $my_repositories
   cd $my_repositories
   git clone git@github.com:stefaanc/terraform-provider-windows
   ```
   > `$my_repositories` must point to the directory where you want to clone the repository
   
2. Build the provider

   ```shell
   cd $my_repositories/terraform-provider-windows
   make release
   ```

   This will build the provider and put it in 
   - `%AppData%\terraform.d\plugins` on Windows
   - `$HOME\.terraform.d\plugins` on Linux
<br/>

 > :bulb:  
 > The makefile provides more commands: `tidy`, `test`, `log`, `report`, `testacc`, `build`, ...
    


<br>

## Installing The Provider

1. Download the provider to your machine

   - go to [the releases tab on github](https://github.com/stefaanc/terraform-provider-hyperv/releases)
   - download the file that is appropriate for your machine

2. Move the provider from your `Downloads` folder to

   - `%AppData%\terraform.d\plugins` on Windows
   - `$HOME\.terraform.d\plugins` on Linux

<br/>
> :bulb:  
> Alternatively, you can try our latest release-in-progress under the `releases` folder.  No guarantee though this will be a fully working provider.



<br>

## Using The Provider

### Example Usage
 
```terraform
provider "windows" {}
```
 
```terraform
provider "windows" {
    version = "~> 0.0"

    type = "local"
}
```

```terraform
provider "windows" {
    version = "~> 0.0"

    type     = "ssh"
    host     = "localhost"
    port     = 22
    user     = "me"
    password = "my-password"
    insecure = false
}
```

<br/>

### Argument Attributes Reference

- `type` - (Optional, defaults to `"local"`) -  The type of connection to the windows computer: `"local"` or `"ssh"`.

For `type = "local"`

- Any other arguments are ignored.

For `type = "ssh"` 

- `host` - (Optional, defaults to `"localhost"`) -  The name or IP-address of the windows computer.

- `port` - (Optional, defaults to `22`) -  The port for SSH-communication with the windows computer.

- `user` - (Required) -  The user name for communication with the windows computer.

- `password` - (Required) -  The user password for communication with the windows computer.

- `insecure` - (Optional, defaults to `false`) -  Allow insecure communication.  When `insecure = false`, the certificate of the windows computer is checked against the user's known hosts on the machine that runs Terraform, as specified by the file `~/.ssh/known_hosts`.  When `insecure = true`, this check is disabled.

<br/>
> :bulb:  
> The provider's API needs elevated credentials ("Run as Administrator") for most methods.
> When using `type = "local"`, you need to run terraform from an elevated shell.
> When using `type = "ssh"`, terraform will always use the most elevated credentials available to the configured user.

<br/>
> :bulb:  
> The easiest way to add a host to `~/.ssh/known_hosts` is by using the command: `ssh-keyscan $host >> ~/.ssh/known_hosts`, where `$host` is `localhost` or the name or IP address of the host.  
> You can also hash the hosts names and addresses by using the `-H` switch: `ssh-keyscan -H $host >> ~/.ssh/known_hosts`
>  
> When this file is newly created, make sure that this file uses UTF-8 encoding without byte-order mark (BOM).  In Powershell you can set the default encoding for `>`, `>>` and `Out-File` by using the command: `$PSDefaultParameterValues['Out-File:Encoding'] = 'utf8'`.  Alternatively, you can set the encoding on a "per-command" basis using: `ssh-keyscan $host | Out-File ~/.ssh/known_hosts -Encoding 'utf8'` 

<br/>

### Data Sources

- [**windows_computer**](docs/datasource.windows_computer.md) -  Exports the attributes of a windows computer.  This includes it's DNS-client attributes, and reboot-pending status.

- [**windows_network_adapter**](docs/datasource.windows_network_adapter.md) -  Exports the attributes of a network-adapter.  This includes it's MAC address, DNS-client attributes, and statusses.

- [**windows_network_connection**](docs/datasource.windows_network_connection.md) -  Exports the attributes of a network-connection.  This includes it's IPv4 and IPv6 gateways, connection-profile, and connectivity-status.

- [**windows_network_interface**](docs/datasource.windows_network_interface.md) -  Exports the attributes of a network interface.  This provides identifying attributes of other resources that are associated to this network interface.  This includes it's GUID, index, alias, description, MAC address, associated network-adapter name, associated vnetwork-adapter name, associated network-connection names, associated vswitch name and associated computer name. 

<br/>

### Resources

- [**windows_computer**](docs/resource.windows_computer.md) -  Provides access to the attributes of a windows computer.  This includes it's DNS-client attributes, and reboot-pending status.

- [**windows_network_adapter**](docs/resource.windows_network_adapter.md) -  Provides access to the attributes of a network-adapter.  This includes it's MAC address, DNS-client attributes, and statusses.

- [**windows_network_connection**](docs/resource.windows_network_connection.md) -  Provides access to the attributes of a network-connection.  This includes it's IPv4 and IPv6 gateways, connection-profile, and connectivity-status.



<br/>

## Extended Lifecycle

This providers supports a number of extensions to the standard Terraform lifecycle.

### Data Sources

Sometimes one doesn't know if a data source exists or not.  An example is the `Default Switch` in Hyper-V.  This was introduced in some version of Hyper-V.  One cannot read the data source in Terraform, unless one is absolutely sure it exists, because Terraform will throw an error when it is not there.

Using the extended lifecycle attributes, one can read such data sources, without throwing an error when they don't exist.  This allows to implement dynamic Terraform behaviour depending on the existence of the data source.

###### Example Usage

```terraform
data "data_source" "my_data_source" {
    x_lifecycle {
        ignore_error_if_not_exists = true 
    }
}

output "my_data_source_exists" {
    value = data_source.my_data_source.x_lifecycle[0].exists
}
```

###### Argument Attributes Reference

- `ignore_error_if_not_exists` - (boolean, Optional, defaults to `false`) -  If the data source doesn't exist, the Terraform state contains zeroed attributes for this data source.  No error is thrown. 

###### Exported Attributes Reference

- `exists` - (boolean) -  The data source exists, and the Terraform state contains the attributes of the data source.

<br/>

### Resources

Terraform support importing resources using `Terraform import`.  However, this requires a manual (or externally scripted) action.  Using the extended lifecycle attributes, this can be automated in Terraform.

###### Example Usage

```terraform
resource "resource" "my_resource" {
    x_lifecycle {
        import_if_exists    = true 
        destroy_if_imported = true
    }
}

output "my_resource_imported" {
    value = resource.my_resource.x_lifecycle[0].imported
}
```

###### Argument Attributes Reference

- `import_if_exists` - (boolean, Optional, defaults to `false`) -  If the resource exists, it is imported into the Terraform state, it's original attributes are saved so they can be reinstated at a later time, and the resource is updated based on the attributes in the Terraform configuration.  No error is thrown.

- `destroy_if_imported` - (boolean, Optional, defaults to `false`) - If the resource is imported and if this attribute is set to `false`, the resource's original attributes are restored when calling `Terraform destroy`.  If the resource is imported and if this attribute is set to `true` the resource is destroyed when calling `Terraform destroy`.

###### Exported Attributes Reference

- `imported` - (boolean) -  The resource is imported.  Remark that this attribute is not set when the resource was imported using `Terraform import`.

<br/>

### Persistent Resources

A special class of resources are resources that cannot be created using Terraform, and cannot be destroyed using Terraform.  In this respect they behave similar to data sources.  However, unlike data sources but like other resources, one can change some of the properties of these resources.  Typical examples are physical resources or resources that are related to physical resources, like a physical computer or a physical network adapter.

Also, unlike data sources, the original attribute values will be restored when deleting the resource from the Terraform state, even if these values were set outside of Terraform.  Therefore, in order to undo any changes that happen outside of Terraform, it is important to always use a "data source" instead of a "resource" when you don't intend to change any properties of a persistent resource.  All of the persistent resources have a corresponding Terraform "data source" for this reason.

For these resources:
 
- Terraform's "Create" lifecycle-method imports the resource, saves the imported state so it can be reinstated at a later time, and updates the resource based on the attributes in the Terraform configuration.  
- Terraform's "Destroy" lifecycle-method reinstates the originally imported state. 

This corresponds to an implicit resource-`x-lifecycle` behaviour, where `import_if_exists = true` and `destroy_if_imported = false`.

Some of these persistent resources can also support an explicit data-source-`x-lifecycle`.

###### Example Usage

```terraform
resource "persistent_resource" "my_resource" {
    x_lifecycle {
        ignore_error_if_not_exists = true 
    }
}

output "my_resource_exists" {
    value = persistent_resource.my_resource.x_lifecycle[0].exists
}
```

###### Argument Attributes Reference

- `ignore_error_if_not_exists` - (boolean, Optional, defaults to `false`) -  If the resource doesn't exist, the Terraform state contains zeroed attributes for this resource.  No error is thrown. 

###### Exported Attributes Reference

- `exists` - (boolean) -  The resource exists, and the Terraform state contains the attributes of the resource.

<br/>
