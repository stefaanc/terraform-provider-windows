//
// Copyright (c) 2019 Stefaan Coussement
// MIT License
//
// more info: https://github.com/stefaanc/terraform-provider-windows
//
package api

import (
)

//------------------------------------------------------------------------------

type WindowsClient struct {
    Type       string   // "local" or "ssh"

    // local

    // ssh
    Host       string
    Port       uint16
    User       string
    Password   string
    Insecure   bool
}

//------------------------------------------------------------------------------
