# Terraform Provider Windows

**A terraform provider to work with a Windows computer**

The Windows provider is used to interact with the resources supported by a Windows computer.  
The provider has been developed and tested on a Windows 10 machine, but may be used with other versions of Windows with some limitations.



<br/>

### !!! UNDER CONSTRUCTION !!!!!!!!



<br/>

## Prerequisites

To build:
- [GNU make](https://www.gnu.org/software/make/)
- [Golang](https://golang.org/) >= v1.13
- [Terraform plugin SDK](https://github.com/hashicorp/terraform-plugin-sdk) ~= v1.0.0

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

> :bulb:  
> The provider's API needs elevated credentials ("Run as Administrator") for most methods.
> When using `type = "local"`, you need to run terraform from an elevated shell.
> When using `type = "ssh"`, terraform will always use the most elevated credentials available to the configured user.

<br/>

### Data Sources

- [**windows_link_ip_interface**](docs/datasource.windows_link_ip_interface.md) -  Provides identifying attributes for resources that are related to an IP interface.

<br/>

### Resources

- [**windows_computer**](docs/resource.windows_computer.md) -  Provides access to the attributes of a windows computer.  This includes it's name, dns-client attributes, and reboot-pending status.

- [**windows_network_connection**](docs/resource.windows_network_connection.md) -  Provides access to the attributes of the network-connections of a windows computer.  This includes their IPv4 and IPv6 default gateways, dns-client attributes, and status.

- [**windows_network_adapter**](docs/resource.windows_network_adapter.md) -  Provides access to the attributes of the network-adapters of a windows computer.  This includes their MAC address, IP addresses, default gateways, DNS server addresses, and status.
