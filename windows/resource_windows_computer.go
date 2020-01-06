//
// Copyright (c) 2019 Stefaan Coussement
// MIT License
//
// more info: https://github.com/stefaanc/terraform-provider-windows
//
package windows

import (
    "fmt"
    "log"
    "reflect"
    "strings"

    "github.com/hashicorp/terraform-plugin-sdk/helper/schema"
    "github.com/hashicorp/terraform-plugin-sdk/helper/validation"

    "github.com/stefaanc/terraform-provider-windows/api"
    "github.com/stefaanc/terraform-provider-windows/windows/tfutil"
)

//------------------------------------------------------------------------------

func resourceWindowsComputer() *schema.Resource {
    return &schema.Resource{
        Schema: map[string]*schema.Schema{
            "name": &schema.Schema{
                Type:     schema.TypeString,
                Computed: true,
            },
            "new_name": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,

                StateFunc: tfutil.StateToUpper(),
            },

            "dns_client": &schema.Schema{
                Type:     schema.TypeList,
                MaxItems: 1,
                Optional: true,
                Computed: true,
                Elem: resourceWindowsComputerDNSClient(),
            },

            "reboot_pending": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },
            "reboot_pending_details": &schema.Schema{
                Type:     schema.TypeList,
                MaxItems: 1,
                Computed: true,
                Elem: resourceWindowsComputerRebootPendingDetails(),
            },

            "network_adapter_names": &schema.Schema{
                Type:     schema.TypeList,
                Computed: true,
                Elem:     &schema.Schema{ Type: schema.TypeString },
            },
            "network_connection_names": &schema.Schema{
                Type:     schema.TypeList,
                Computed: true,
                Elem:     &schema.Schema{ Type: schema.TypeString },
            },

            // used to reset values on terraform destroy
            "original": &schema.Schema{
                Type:     schema.TypeList,
                MaxItems: 1,
                Computed: true,
                Elem: resourceWindowsComputerOriginal(),
            },
        },

        CustomizeDiff: resourceWindowsComputerCustomizeDiff,

        Create: resourceWindowsComputerCreate,
        Read:   resourceWindowsComputerRead,
        Update: resourceWindowsComputerUpdate,
        Delete: resourceWindowsComputerDelete,
    }
}

func resourceWindowsComputerDNSClient() *schema.Resource {
    return &schema.Resource{
        Schema: map[string]*schema.Schema{
            "suffix_search_list": &schema.Schema{
                Type:     schema.TypeList,
                Elem:     &schema.Schema{ Type: schema.TypeString, StateFunc: tfutil.StateToLower() },
                Optional: true,
                Computed: true,
            },
            "enable_devolution": &schema.Schema{
                Type:     schema.TypeBool,
                Optional: true,
                Computed: true,
            },
            "devolution_level": &schema.Schema{
                Type:     schema.TypeInt,   // uint32
                Optional: true,
                Computed: true,

                ValidateFunc: validation.IntBetween(0, 4294967295),
            },
        },
    }
}

func resourceWindowsComputerRebootPendingDetails() *schema.Resource {
    return &schema.Resource{
        Schema: map[string]*schema.Schema{
            "reboot_required": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },
            "post_reboot_reporting": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },
            "dvd_reboot_signal": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },
            "reboot_pending": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },
            "reboot_in_progress": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },
            "packages_pending": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },
            "services_pending": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },
            "update_exe_volatile": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },
            "computer_rename_pending": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },
            "file_rename_pending": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },
            "netlogon_pending": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },
            "current_reboot_attemps": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },
        },
    }
}

func resourceWindowsComputerOriginal() *schema.Resource {
    return &schema.Resource{
        Schema: map[string]*schema.Schema{
            "new_name": &schema.Schema{
                // save the pending new name instead of the current name, as this would overwrite the current name on reboot
                Type:     schema.TypeString,
                Computed: true,
            },

            "dns_client": &schema.Schema{
                Type:     schema.TypeList,
                MaxItems: 1,
                Computed: true,
                Elem: &schema.Resource{
                    Schema: map[string]*schema.Schema{
                        "suffix_search_list": &schema.Schema{
                            Type:     schema.TypeList,
                            Elem:     &schema.Schema{ Type: schema.TypeString },
                            Computed: true,
                        },
                        "enable_devolution": &schema.Schema{
                            Type:     schema.TypeBool,
                            Computed: true,
                        },
                        "devolution_level": &schema.Schema{
                            Type:     schema.TypeInt,   // uint32
                            Computed: true,
                        },
                    },
                },
            },
        },
    }
}

//------------------------------------------------------------------------------

func resourceWindowsComputerCustomizeDiff(d *schema.ResourceDiff, m interface{}) error {
    // set reboot_pending attributes when new_name changes
    if d.HasChange("new_name") {
        oldName := d.Get("name").(string)
        newName := strings.ToUpper(d.Get("new_name").(string))

        oldComputerNamePending := d.Get("reboot_pending_details.0.computer_rename_pending").(bool)
        newComputerNamePending := oldComputerNamePending
        if ( oldComputerNamePending  && ( newName == oldName ) ) ||
           ( !oldComputerNamePending && ( newName != oldName ) ) {
            newComputerNamePending = !oldComputerNamePending
        }

        if newComputerNamePending != oldComputerNamePending {
            if v, ok := d.GetOk("reboot_pending_details"); ok {
                details := v.([]interface{})[0].(map[string]interface{})
                details["computer_rename_pending"] = newComputerNamePending
                d.SetNew("reboot_pending_details", []interface{}{ details })

                d.SetNewComputed("reboot_pending")
            }
        }
    }

    return nil
}

//------------------------------------------------------------------------------

func resourceWindowsComputerCreate(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.WindowsClient)

    newName     := d.Get("newName")
    dnsClient   := tfutil.GetResource(d, "dns_client")

    host := "localhost"
    if c.Type != "local" {
        host = c.Host
    }

    id := fmt.Sprintf("//%s/computer", host)

    log.Printf(`[INFO][terraform-provider-windows] creating windows_computer %q
                    [INFO][terraform-provider-windows]     newName: %#v
                    [INFO][terraform-provider-windows]     dns_client {
                    [INFO][terraform-provider-windows]         suffix_search_list: %#v
                    [INFO][terraform-provider-windows]         enable_devolution:  %#v
                    [INFO][terraform-provider-windows]         devolution_level:   %#v
                    [INFO][terraform-provider-windows]     }
`       ,
        id,
        newName,
        dnsClient["suffix_search_list"],
        dnsClient["enable_devolution"],
        dnsClient["devolution_level"],
    )

    // import
    log.Printf("[INFO][terraform-provider-windows] importing windows_computer %q into terraform state\n", id)

    computer, err := c.ReadComputer()
    if err != nil {
        // no lifecycle customizations
        log.Printf("[ERROR][terraform-provider-windows] cannot import windows_computer %q into terraform state\n", id)
        return err
    }

    // save original config
    setOriginalComputerProperties(d, computer)

    if !diffComputerProperties(d, computer) {
        // no update required

        // set properties
        setComputerProperties(d, computer)

        // set id
        d.SetId(id)

        log.Printf("[INFO][terraform-provider-windows] created windows_computer %q\n", id)
        return nil

    } else {
        // update
        log.Printf("[INFO][terraform-provider-windows] updating windows_computer %q\n", id)

        cProperties := new(api.Computer)
        expandComputerProperties(cProperties, d)

        err := c.UpdateComputer(cProperties)
        if err != nil {
            log.Printf("[ERROR][terraform-provider-windows] cannot update windows_computer %q\n", id)
            return err
        }

        // set id
        d.SetId(id)

        log.Printf("[INFO][terraform-provider-windows] created windows_computer %q\n", id)
        return resourceWindowsComputerRead(d, m)
    }
}

//------------------------------------------------------------------------------

func resourceWindowsComputerRead(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.WindowsClient)

    id          := d.Id()

    log.Printf("[INFO][terraform-provider-windows] reading windows_computer %q\n", id)

    // read
    computer, err := c.ReadComputer()
    if err != nil {
        // no lifecycle customizations
        log.Printf("[ERROR][terraform-provider-windows] cannot read windows_computer %q\n", id)

        // set id
        d.SetId("")

        log.Printf("[INFO][terraform-provider-windows] deleted windows_computer %q from terraform state\n", id)
        return nil   // don't return an error to allow terraform refresh to update state
    }

    // set properties
    setComputerProperties(d, computer)

    log.Printf("[INFO][terraform-provider-windows] read windows_computer %q\n", id)
    return nil
}

//------------------------------------------------------------------------------

func resourceWindowsComputerUpdate(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.WindowsClient)

    id        := d.Id()
    newName   := d.Get("newName")
    dnsClient := tfutil.GetResource(d, "dns_client")

    log.Printf(`[INFO][terraform-provider-windows] updating windows_computer %q
                    [INFO][terraform-provider-windows]     newName: %#v
                    [INFO][terraform-provider-windows]     dns_client {
                    [INFO][terraform-provider-windows]         suffix_search_list: %#v
                    [INFO][terraform-provider-windows]         enable_devolution:  %#v
                    [INFO][terraform-provider-windows]         devolution_level:   %#v
                    [INFO][terraform-provider-windows]     }
`       ,
        id,
        newName,
        dnsClient["suffix_search_list"],
        dnsClient["enable_devolution"],
        dnsClient["devolution_level"],
    )

    // update
    cProperties := new(api.Computer)
    expandComputerProperties(cProperties, d)

    err := c.UpdateComputer(cProperties)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-windows] cannot update windows_computer %q\n", id)
        return err
    }

    log.Printf("[INFO][terraform-provider-windows] updated windows_computer %q\n", id)
    return resourceWindowsComputerRead(d, m)
}

//------------------------------------------------------------------------------

func resourceWindowsComputerDelete(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.WindowsClient)

    id       := d.Id()

    log.Printf("[INFO][terraform-provider-windows] deleting windows_computer %q from terraform state\n", id)
    log.Printf("[INFO][terraform-provider-windows] restoring original properties for windows_computer %q\n", id)

    // restore original config
    cProperties := new(api.Computer)
    expandOriginalComputerProperties(cProperties, d)

    err := c.UpdateComputer(cProperties)
    if err != nil {
        log.Printf("[WARNING][terraform-provider-windows] cannot restore original properties for windows_computer %q\n", id)
    }

    // set id
    d.SetId("")

    log.Printf("[INFO][terraform-provider-windows] deleted windows_computer %q from terraform state\n", id)
    return nil
}

//------------------------------------------------------------------------------

func setComputerProperties(d *schema.ResourceData, cProperties *api.Computer) {
    d.Set("name", cProperties.Name)
    d.Set("new_name", cProperties.NewName)

    dnsClient := make(map[string]interface{})
    dnsClient["suffix_search_list"] = cProperties.DNSClient.SuffixSearchList
    dnsClient["enable_devolution"]  = cProperties.DNSClient.EnableDevolution
    dnsClient["devolution_level"]   = cProperties.DNSClient.DevolutionLevel
    d.Set("dns_client", []interface{}{ dnsClient })

    d.Set("reboot_pending", cProperties.RebootPending)

    rebootPendingDetails := make(map[string]interface{})
    rebootPendingDetails["reboot_required"]         = cProperties.RebootPendingDetails.RebootRequired
    rebootPendingDetails["post_reboot_reporting"]   = cProperties.RebootPendingDetails.PostRebootReporting
    rebootPendingDetails["dvd_reboot_signal"]       = cProperties.RebootPendingDetails.DVDRebootSignal
    rebootPendingDetails["reboot_pending"]          = cProperties.RebootPendingDetails.RebootPending
    rebootPendingDetails["reboot_in_progress"]      = cProperties.RebootPendingDetails.RebootInProgress
    rebootPendingDetails["packages_pending"]        = cProperties.RebootPendingDetails.PackagesPending
    rebootPendingDetails["services_pending"]        = cProperties.RebootPendingDetails.ServicesPending
    rebootPendingDetails["update_exe_volatile"]     = cProperties.RebootPendingDetails.UpdateExeVolatile
    rebootPendingDetails["computer_rename_pending"] = cProperties.RebootPendingDetails.ComputerRenamePending
    rebootPendingDetails["file_rename_pending"]     = cProperties.RebootPendingDetails.FileRenamePending
    rebootPendingDetails["netlogon_pending"]        = cProperties.RebootPendingDetails.NetlogonPending
    rebootPendingDetails["current_reboot_attemps"]  = cProperties.RebootPendingDetails.CurrentRebootAttemps
    d.Set("reboot_pending_details", []interface{}{ rebootPendingDetails })

    d.Set("network_adapter_names", cProperties.NetworkAdapterNames)
    d.Set("network_connection_names", cProperties.NetworkConnectionNames)
}

func setOriginalComputerProperties(d *schema.ResourceData, cProperties *api.Computer) {
    original := make(map[string]interface{})

    original["new_name"] = cProperties.NewName

    original_dnsClient := make(map[string]interface{})
    original_dnsClient["suffix_search_list"] = cProperties.DNSClient.SuffixSearchList
    original_dnsClient["enable_devolution"]  = cProperties.DNSClient.EnableDevolution
    original_dnsClient["devolution_level"]   = cProperties.DNSClient.DevolutionLevel
    original["dns_client"] = []interface{}{ original_dnsClient }

    d.Set("original", []interface{}{ original })
}

//------------------------------------------------------------------------------

func diffComputerProperties(d *schema.ResourceData, cProperties *api.Computer) bool {
    if v, ok := d.GetOk("new_name"); ok && ( cProperties.NewName != v.(string) ) {
        return true
    }

    if v, ok := d.GetOk("dns_client"); ok && ( len(v.([]interface{})) > 0 ) {
        if v, ok := d.GetOkExists("dns_client.0.suffix_search_list"); ok && !reflect.DeepEqual(cProperties.DNSClient.SuffixSearchList, v) {
            return true
        }
        if v, ok := d.GetOkExists("dns_client.0.enable_devolution"); ok && ( cProperties.DNSClient.EnableDevolution != v.(bool) ) {
            return true
        }
        if v, ok := d.GetOkExists("dns_client.0.devolution_level"); ok && ( cProperties.DNSClient.DevolutionLevel != uint32(v.(int)) ) {
            return true
        }
    }

    return false
}

func expandComputerProperties(cProperties *api.Computer, d *schema.ResourceData) {
    cProperties.NewName = d.Get("new_name").(string)

    dnsClientList := tfutil.GetListOfResources(d, "dns_client")
    if len(dnsClientList) > 0 {
        if _, ok := d.GetOkExists("dns_client.0.suffix_search_list"); ok {
            dnsClient := dnsClientList[0]
            cProperties.DNSClient.SuffixSearchList = tfutil.ExpandListOfStrings(dnsClient, "suffix_search_list")
        } else {
            // there is no state yet (the resource is being created), and the attribute is not defined in config
            // since we don't have a state yet, we use the previously read original properties to get the current value (using the zero-value 'false', would overwrite the current value)
            original_dnsClient := tfutil.GetResource(d, "original.0.dns_client")
            cProperties.DNSClient.SuffixSearchList = tfutil.ExpandListOfStrings(original_dnsClient, "suffix_search_list")
        }

        if v, ok := d.GetOkExists("dns_client.0.enable_devolution"); ok {
            cProperties.DNSClient.EnableDevolution = v.(bool)
        } else {
            // there is no state yet (the resource is being created), and the attribute is not defined in config
            // since we don't have a state yet, we use the previously read original properties to get the current value (using the zero-value 'false', would overwrite the current value)
            original_dnsClient := tfutil.GetResource(d, "original.0.dns_client")
            cProperties.DNSClient.EnableDevolution = original_dnsClient["enable_devolution"].(bool)
        }

        if v, ok := d.GetOkExists("dns_client.0.devolution_level"); ok {
            cProperties.DNSClient.DevolutionLevel  = uint32(v.(int))
        } else {
            // there is no state yet (the resource is being created), and the attribute is not defined in config
            // since we don't have a state yet, we use the previously read original properties to get the current value (using the zero-value 'false', would overwrite the current value)
            original_dnsClient := tfutil.GetResource(d, "original.0.dns_client")
            cProperties.DNSClient.DevolutionLevel = uint32(original_dnsClient["devolution_level"].(int))
        }
    }
}

func expandOriginalComputerProperties(cProperties *api.Computer, d *schema.ResourceData) {
    original := tfutil.GetResource(d, "original")

    cProperties.NewName = original["new_name"].(string)

    original_dnsClient := tfutil.ExpandResource(original, "dns_client")
    cProperties.DNSClient.SuffixSearchList = tfutil.ExpandListOfStrings(original_dnsClient, "suffix_search_list")
    cProperties.DNSClient.EnableDevolution = original_dnsClient["enable_devolution"].(bool)
    cProperties.DNSClient.DevolutionLevel  = uint32(original_dnsClient["devolution_level"].(int))
}

//------------------------------------------------------------------------------
