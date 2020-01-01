//
// Copyright (c) 2019 Stefaan Coussement
// MIT License
//
// more info: https://github.com/stefaanc/terraform-provider-windows
//
package tfutil

import (
    "fmt"
    "regexp"
    "strings"

    "github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

//------------------------------------------------------------------------------

var DataSourceXLifecycleSchema schema.Schema =  schema.Schema{
    Type:     schema.TypeList,
    MaxItems: 1,
    Optional: true,
    Computed: true,
    Elem: &schema.Resource{
        Schema: map[string]*schema.Schema{
            // "ignore_error_if_not_exists" ignores the error when the data source doesn't exist, the resource is added to the terraform state with zeroed properties
            // this can be used in conditional resources
            "ignore_error_if_not_exists": &schema.Schema{
                Type:     schema.TypeBool,
                Optional: true,
                Default:  false,
            },
            // "exists" true if the resource exists
            "exists": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },
        },
    },
}

var ResourceXLifecycleSchema schema.Schema = schema.Schema{
    Type:     schema.TypeList,
    MaxItems: 1,
    Optional: true,
    Computed: true,
    Elem: &schema.Resource{
        Schema: map[string]*schema.Schema{
            // "import_if_exists" imports the resource into the terraform state when creating a resource that already exists
            // this will fail if any of the properties in the config are not the same as the properties of existing resource - this is to reduce the risk of accidental imports
            // this can be used to adapt resources that are shared with external systems
            "import_if_exists": &schema.Schema{
                Type:     schema.TypeBool,
                Optional: true,
                Default:  false,
            },
            // "imported" set if the resource was imported using the "import_if_exists" lifecycle customization
            "imported": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },
            // "destroy_if_imported" destroys the resource from the infrastructure when using 'terraform destroy' and when it was imported using 'import_if_exists = true'
            // by default, a resource that is imported using 'import_if_exists = true' is not destroyed from the infrastructure when using 'terraform destroy'
            "destroy_if_imported": &schema.Schema{
                Type:     schema.TypeBool,
                Optional: true,
                Default:  false,
            },
        },
    },
}

//------------------------------------------------------------------------------

func ValidateUUID() schema.SchemaValidateFunc {
    return func(i interface{}, k string) ([]string, []error) {
        v, ok := i.(string)
        if !ok {
            return nil, []error{fmt.Errorf("expected type of %s to be a string", k)}
        }

        v = strings.ToLower(v)
        r := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
        if !r.MatchString(v) {
            return nil, []error{fmt.Errorf("expected value of %s to be a valid UUID, using format \"xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx\" where \"x\" is a hex digit, got: %s", k, v)}
        }
        return nil, nil
    }
}

func ValidateSingleMAC() schema.SchemaValidateFunc {
    return func(i interface{}, k string) ([]string, []error) {
        v, ok := i.(string)
        if !ok {
            return nil, []error{fmt.Errorf("expected type of %s to be a string", k)}
        }

        v = strings.ToLower(v)
        r := regexp.MustCompile(`^[0-9a-f]{2}(-[0-9a-f]{2}){5}$`)
        if !r.MatchString(v) {
            return nil, []error{fmt.Errorf("expected value of %s to be a valid MAC, using format \"xx-xx-xx-xx-xx-xx\" where \"x\" is a hex digit, got: %s", k, v)}
        }
        return nil, nil
    }
}

//------------------------------------------------------------------------------

func StateAll(funcs ...schema.SchemaStateFunc) schema.SchemaStateFunc {
    return func(val interface{}) string {
        for _, f := range funcs {
            val = f(val)
        }
        return val.(string)
    }
}

func StateToUpper() schema.SchemaStateFunc {
    return func(val interface{}) string {
        return strings.ToUpper(val.(string))
    }
}

func StateToLower() schema.SchemaStateFunc {
    return func(val interface{}) string {
        return strings.ToLower(val.(string))
    }
}

func StateToCamel() schema.SchemaStateFunc {
    return func(val interface{}) string {
        s := strings.ToLower(val.(string))
        c := s[:1]
        return strings.Replace(s, c, strings.ToUpper(c), 1)
    }
}

func StateAcceptEmptyString() schema.SchemaStateFunc {
    return func(val interface{}) string {
        v := val.(string)
        if v == "" {
            return "<empty>"
        } else {
            return v
        }
    }
}

//------------------------------------------------------------------------------

// func DiffSuppressCase() schema.SchemaDiffSuppressFunc {
//     return func(k, old, new string, d *schema.ResourceData) bool {
//         if strings.ToLower(old) == strings.ToLower(new) {
//             return true
//         }
//         return false
//     }
// }

//------------------------------------------------------------------------------

func GetResource(d *schema.ResourceData, name string) (m map[string]interface{}) {
    if listOfInterfaces1, ok := d.GetOk(name); ok {
        listOfInterfaces2 := listOfInterfaces1.([]interface{})

        if len(listOfInterfaces2) > 0 {
            m = listOfInterfaces2[0].(map[string]interface{})
        }
    }

    if m == nil {
        m = make(map[string]interface{})
    }

    return m
}

func ExpandResource(d map[string]interface{}, name string) (m map[string]interface{}) {
    if listOfInterfaces1, ok := d[name]; ok {
        listOfInterfaces2 := listOfInterfaces1.([]interface{})

        if len(listOfInterfaces2) > 0 {
            m = listOfInterfaces2[0].(map[string]interface{})
        }
    }

    if m == nil {
        m = make(map[string]interface{})
    }

    return m
}

//------------------------------------------------------------------------------

func GetListOfResources(d *schema.ResourceData, name string) (l []map[string]interface{}) {
    if listOfInterfaces1, ok := d.GetOk(name); ok {
        listOfInterfaces2 := listOfInterfaces1.([]interface{})

        l = make([]map[string]interface{}, len(listOfInterfaces2))
        for i, m := range listOfInterfaces2 {
            l[i] = m.(map[string]interface{})
        }
    }

    return l
}

func ExpandListOfResources(d map[string]interface{}, name string) (l []map[string]interface{}) {
    if listOfInterfaces1, ok := d[name]; ok {
        listOfInterfaces2 := listOfInterfaces1.([]interface{})

        l = make([]map[string]interface{}, len(listOfInterfaces2))
        for i, m := range listOfInterfaces2 {
            l[i] = m.(map[string]interface{})
        }
    }

    return l
}

//------------------------------------------------------------------------------

func GetSetOfResources(d *schema.ResourceData, name string) (l []map[string]interface{}) {
    if listOfInterfaces1, ok := d.GetOk(name); ok {
        listOfInterfaces2 := listOfInterfaces1.(*schema.Set).List()

        l = make([]map[string]interface{}, len(listOfInterfaces2))
        for i, m := range listOfInterfaces2 {
            l[i] = m.(map[string]interface{})
        }
    }

    return l
}

func ExpandSetOfResources(d map[string]interface{}, name string) (l []map[string]interface{}) {
    if listOfInterfaces1, ok := d[name]; ok {
        listOfInterfaces2 := listOfInterfaces1.(*schema.Set).List()

        l = make([]map[string]interface{}, len(listOfInterfaces2))
        for i, m := range listOfInterfaces2 {
            l[i] = m.(map[string]interface{})
        }
    }

    return l
}

//------------------------------------------------------------------------------

func GetListOfStrings(d *schema.ResourceData, name string) (l []string) {
    if listOfInterfaces1, ok := d.GetOk(name); ok {
        listOfInterfaces2 := listOfInterfaces1.([]interface{})

        l = make([]string, len(listOfInterfaces2))
        for i, s := range listOfInterfaces2 {
            l[i] = s.(string)
        }
    }

    return l
}

func ExpandListOfStrings(d map[string]interface{}, name string) (l []string) {
    if listOfInterfaces1, ok := d[name]; ok {
        listOfInterfaces2 := listOfInterfaces1.([]interface{})

        l = make([]string, len(listOfInterfaces2))
        for i, s := range listOfInterfaces2 {
            l[i] = s.(string)
        }
    }

    return l
}

//------------------------------------------------------------------------------
