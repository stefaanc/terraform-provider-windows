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
                MinItems: 1,
                MaxItems: 1,
                Computed: true,
                Elem: resourceWindowsComputerRebootPendingDetails(),
            },

            "original": &schema.Schema{   // used to reset values on terraform destroy
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

    newName   := d.Get("newName")
    dnsClient := tfutil.GetResource(d, "dns_client")

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

    // import computer
    log.Printf("[INFO][terraform-provider-windows] importing windows_computer %q into terraform state\n", id)

    managementOS, err := c.ReadComputer()
    if err != nil {
        log.Printf("[ERROR][terraform-provider-windows] cannot import windows_computer %q into terraform state\n", id)
        return err
    }

    // save existing config
    setOriginalComputerProperties(d, managementOS)

    // update computer
    if managementOS.NewName              != managementOS.Name ||
       managementOS.NewName              != newName ||
       !reflect.DeepEqual(managementOS.DNSClient.SuffixSearchList, dnsClient["suffix_search_list"]) ||
       managementOS.DNSClient.EnableDevolution != dnsClient["enable_devolution"] ||
       managementOS.DNSClient.DevolutionLevel  != dnsClient["devolution_level"] {

        log.Printf("[INFO][terraform-provider-windows] updating windows_computer %q\n", id)

        mosProperties := new(api.Computer)
        expandComputerProperties(mosProperties, d)

        err := c.UpdateComputer(mosProperties)
        if err != nil {
            log.Printf("[ERROR][terraform-provider-windows] cannot update windows_computer %q\n", id)
            log.Printf("[ERROR][terraform-provider-windows] cannot import windows_computer %q into terraform state\n", id)
            return err
        }
    }

    // set id
    d.SetId(id)

    log.Printf("[INFO][terraform-provider-windows] created windows_computer %q\n", id)
    return resourceWindowsComputerRead(d, m)
}

//------------------------------------------------------------------------------

func resourceWindowsComputerRead(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.WindowsClient)

    id       := d.Id()
    original := tfutil.GetResource(d, "original")   // make sure new terraform state includes 'original' from the old terraform state when doing a terraform refresh

    log.Printf("[INFO][terraform-provider-windows] reading windows_computer %q\n", id)

    // read computer
    managementOS, err := c.ReadComputer()
    if err != nil {
        // no lifecycle customizations
        log.Printf("[ERROR][terraform-provider-windows] cannot read windows_computer %q\n", id)

        // set id
        d.SetId("")

        log.Printf("[INFO][terraform-provider-windows] deleted windows_computer %q from terraform state\n", id)
        return nil   // don't return an error to allow terraform refresh to update state
    }

    // set properties
    setComputerProperties(d, managementOS)
    d.Set("original", []interface{}{ original })   // make sure new terraform state includes 'original' from the old terraform state when doing a terraform refresh

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

    // update computer
    mosProperties := new(api.Computer)
    expandComputerProperties(mosProperties, d)

    err := c.UpdateComputer(mosProperties)
    if err != nil {
        log.Printf("[WARNING][terraform-provider-windows] cannot update windows_computer %q\n", id)
        return err
    }

    log.Printf("[INFO][terraform-provider-windows] updated windows_computer %q\n", id)
    return resourceWindowsComputerRead(d, m)
}

//------------------------------------------------------------------------------

func resourceWindowsComputerDelete(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.WindowsClient)

    id       := d.Id()

    log.Printf("[INFO][terraform-provider-windows] deleting windows_computer %q\n", id)
    log.Printf("[INFO][terraform-provider-windows] restore original properties for windows_computer %q\n", id)

    // restore computer
    mosProperties := new(api.Computer)
    expandOriginalComputerProperties(mosProperties, d)

    err := c.UpdateComputer(mosProperties)
    if err != nil {
        log.Printf("[WARNING][terraform-provider-windows] cannot restore original properties for windows_computer %q\n", id)
    }

    // set id
    d.SetId("")

    log.Printf("[INFO][terraform-provider-windows] deleted windows_computer %q\n", id)
    return nil
}

//------------------------------------------------------------------------------

func setComputerProperties(d *schema.ResourceData, mosProperties *api.Computer) {
    d.Set("name", mosProperties.Name)
    d.Set("new_name", mosProperties.NewName)

    dnsClient := make(map[string]interface{})
    dnsClient["suffix_search_list"] = mosProperties.DNSClient.SuffixSearchList
    dnsClient["enable_devolution"]  = mosProperties.DNSClient.EnableDevolution
    dnsClient["devolution_level"]   = mosProperties.DNSClient.DevolutionLevel
    d.Set("dns_client", []interface{}{ dnsClient })

    d.Set("reboot_pending", mosProperties.RebootPending)

    rebootPendingDetails := make(map[string]interface{})
    rebootPendingDetails["reboot_required"]         = mosProperties.RebootPendingDetails.RebootRequired
    rebootPendingDetails["post_reboot_reporting"]   = mosProperties.RebootPendingDetails.PostRebootReporting
    rebootPendingDetails["dvd_reboot_signal"]       = mosProperties.RebootPendingDetails.DVDRebootSignal
    rebootPendingDetails["reboot_pending"]          = mosProperties.RebootPendingDetails.RebootPending
    rebootPendingDetails["reboot_in_progress"]      = mosProperties.RebootPendingDetails.RebootInProgress
    rebootPendingDetails["packages_pending"]        = mosProperties.RebootPendingDetails.PackagesPending
    rebootPendingDetails["services_pending"]        = mosProperties.RebootPendingDetails.ServicesPending
    rebootPendingDetails["update_exe_volatile"]     = mosProperties.RebootPendingDetails.UpdateExeVolatile
    rebootPendingDetails["computer_rename_pending"] = mosProperties.RebootPendingDetails.ComputerRenamePending
    rebootPendingDetails["file_rename_pending"]     = mosProperties.RebootPendingDetails.FileRenamePending
    rebootPendingDetails["netlogon_pending"]        = mosProperties.RebootPendingDetails.NetlogonPending
    rebootPendingDetails["current_reboot_attemps"]  = mosProperties.RebootPendingDetails.CurrentRebootAttemps
    d.Set("reboot_pending_details", []interface{}{ rebootPendingDetails })

}

func setOriginalComputerProperties(d *schema.ResourceData, mosProperties *api.Computer) {
    original := make(map[string]interface{})

    original["new_name"] = mosProperties.NewName

    original_dnsClient := make(map[string]interface{})
    original_dnsClient["suffix_search_list"] = mosProperties.DNSClient.SuffixSearchList
    original_dnsClient["enable_devolution"]  = mosProperties.DNSClient.EnableDevolution
    original_dnsClient["devolution_level"]   = mosProperties.DNSClient.DevolutionLevel
    original["dns_client"] = []interface{}{ original_dnsClient }

    d.Set("original", []interface{}{ original })
}

//------------------------------------------------------------------------------

func expandComputerProperties(mosProperties *api.Computer, d *schema.ResourceData) {
    if v, ok := d.GetOkExists("new_name"); ok {
        mosProperties.NewName = v.(string)
    }

    if _, ok := d.GetOk("dns_client"); ok {
        dnsClient := tfutil.GetResource(d, "dns_client")

        if _, ok := d.GetOkExists("dns_client.0.suffix_search_list"); ok {
            mosProperties.DNSClient.SuffixSearchList = tfutil.ExpandListOfStrings(dnsClient, "suffix_search_list")
        }
        if v, ok := d.GetOkExists("dns_client.0.enable_devolution"); ok {
            mosProperties.DNSClient.EnableDevolution = v.(bool)
        }
        if v, ok := d.GetOkExists("dns_client.0.devolution_level"); ok {
            mosProperties.DNSClient.DevolutionLevel  = uint32(v.(int))
        }
    }
}

func expandOriginalComputerProperties(mosProperties *api.Computer, d *schema.ResourceData) {
    original := tfutil.GetResource(d, "original")

    mosProperties.NewName = original["new_name"].(string)

    original_dnsClient := tfutil.ExpandResource(original, "dns_client")
    mosProperties.DNSClient.SuffixSearchList = tfutil.ExpandListOfStrings(original_dnsClient, "suffix_search_list")
    mosProperties.DNSClient.EnableDevolution = original_dnsClient["enable_devolution"].(bool)
    mosProperties.DNSClient.DevolutionLevel  = uint32(original_dnsClient["devolution_level"].(int))
}

//------------------------------------------------------------------------------
