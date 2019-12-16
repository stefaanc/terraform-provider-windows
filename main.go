//
// Copyright (c) 2019 Stefaan Coussement
// MIT License
//
// more info: https://github.com/stefaanc/terraform-provider-windows
//
package main

import (
    "github.com/hashicorp/terraform-plugin-sdk/plugin"

    "github.com/stefaanc/terraform-provider-windows/windows"
)

func main() {
    plugin.Serve(&plugin.ServeOpts{
        ProviderFunc: windows.Provider,
    })
}