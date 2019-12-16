## Resource: "windows_computer"

> :bulb:  
> This resource is automatically created by the windows-computer, and cannot be destroyed. 
>  
> - Terraform's "Create" lifecycle-method imports the resource, saves the imported state so it can be reinstated at a later time, and updates the resource based on the attributes in the Terraform configuration.  
> - Terraform's "Destroy" lifecycle-method reinstates the originally imported state. 

### Example Usage

```terraform
resource "windows_computer" "my_computer" {
    new_name = "MY-COMPUTER"

    dns_client {
        suffix_search_list = [ "local" ]
        enable_devolution  = true
        devolution_level   = 0
    }
}

output "my_computer_name" {
    value = windows_computer.my_computer.name
}
output "my_computer_reboot_pending" {
    value = windows_computer.my_computer.reboot_pending
}
output "my_computer_rename_pending" {
    value = windows_computer.my_computer.reboot_pending_details[0].computer_rename_pending
}
```

<br/>

### Argument Attributes Reference

- `new_name` - (string, Optional) -  The new name of the windows-computer.  When this is different from the current name, then the exported `reboot_pending` attribute and `computer_rename_pending` attribute in `reboot_pending_details` will be set to `true`.  The new name will only become effective after a reboot of the windows computer.

- `dns_client` - (resource, Optional)

  - `suffix_search_list` - (list[string], Optional) -  Specifies a list of global suffixes that can be used in the specified order by the DNS client for resolving the IP address of the computer name. These suffixes are appended in the specified order to resolve the computer name that is specified. 
  This attribute cannot be set if the suffix search list setting is already deployed through Group Policy.

  - `enable_devolution` - (boolean, Optional) -  Indicates whether devolution is activated. With devolution, the DNS resolver creates new FQDNs by appending the single-label, unqualified domain name with the parent suffix of the primary DNS suffix name, and the parent of that suffix, and so on, stopping if the name is successfully resolved or at a level specified in the DevolutionLevel parameter. Devolution works by removing the left-most label and continuing to get to the parent suffix. 
  This attribute cannot be set if the devolution level setting is already deployed through Group Policy.

  - `devolution_level` - (integer, Optional) -  Specifies the number of labels up to which devolution should occur.  The devolution level is an integer between `0` and `4294967295`.  If this attribute is `0`, then the FRD algorithm is used. If this attribute is greather than `0`, then devolution occurs until the specified level. 
  This attribute cannot be set if the devolution level setting is already deployed through Group Policy.

<br/>

### Exported Attributes Reference

```terraform
{
    name     = "MY-COMPUTER"
    new_name = "MY-COMPUTER"

    dns_client = [{
        suffix_search_list = [ "local" ]
        enable_devolution  = true
        devolution_level   = 0
    }]

    reboot_pending = false
    reboot_pending_details = [{
        computer_rename_pending = false
        current_reboot_attemps  = false
        dvd_reboot_signal       = false
        file_rename_pending     = false
        netlogon_pending        = false
        packages_pending        = false
        post_reboot_reporting   = false
        reboot_in_progress      = false
        reboot_pending          = false
        reboot_required         = false
        services_pending        = false
        update_exe_volatile     = false
    }]
}
```

All argument attributes are set to the value of the existing resource.
In addition to the argument attributes:

- `name` - (string) -  The name of the windows-computer.

- `reboot_pending` - (boolean) -  The windows-computer is waiting for a reboot.

- `reboot_pending_details` - (resource) -  The reason for `reboot_pending`.

  - `computer_rename_pending` - (boolean) -  The windows-computer is waiting for a reboot because it was given a new name.

  - Other attributes are outside the scope of this documentation.  They refer to Windows registry items - see table below.  For more information, please refer to the Windows documentation

<br/>

> :information_source:  
> 
> Mapping of `reboot_pending_details` on the Windows registry
> 
> attribute                 | key, <br/> `true` condition
> :-------------------------|:-------------------------- 
> `computer_rename_pending` | `HKLM:\SYSTEM\CurrentControlSet\Control\ComputerName\ComputerName`, <br/> `ComputerName` value not equal to `$env:ComputerName`
> `current_reboot_attemps`  | `HKLM:\SOFTWARE\Microsoft\ServerManager\CurrentRebootAttempts`, <br/> key exist
> `dvd_reboot_signal`       | `HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\RunOnce`, <br/> `DVDRebootSignal` value exists
> `file_rename_pending`     | `HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager`, <br/> `PendingFileRenameOperations` value exists or `PendingFileRenameOperations2` value exists
> `netlogon_pending`        | `HKLM:\SYSTEM\CurrentControlSet\Services\Netlogon`, <br/> `JoinDomain` value exists or `AvoidSpnSet` value exists
> `packages_pending`        | `HKLM:\Software\Microsoft\Windows\CurrentVersion\Component Based Servicing\PackagesPending`, <br/> key exist
> `post_reboot_reporting`   | `HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\WindowsUpdate\Auto Update\PostRebootReporting`, <br/> key exists 
> `reboot_in_progress`      | `HKLM:\Software\Microsoft\Windows\CurrentVersion\Component Based Servicing\RebootInProgress`, <br/> key exist
> `reboot_pending`          | `HKLM:\Software\Microsoft\Windows\CurrentVersion\Component Based Servicing\RebootPending`, <br/> key exist
> `reboot_required`         | `HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\WindowsUpdate\Auto Update\RebootRequired`, <br/> key exists
> `services_pending`        | `HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\WindowsUpdate\Services\Pending`, <br/> any GUID subkeys exist
> `update_exe_volatile`     | `HKLM:\SOFTWARE\Microsoft\Updates`, <br/> `UpdateExeVolatile` value not equal to 0
