//
// Copyright (c) 2019 Stefaan Coussement
// MIT License
//
// more info: https://github.com/stefaanc/terraform-provider-windows
//
package windows

import (
    "log"

    "github.com/stefaanc/terraform-provider-windows/api"
)

//------------------------------------------------------------------------------

type Config struct {
    Type     string

    // ssh
    Host     string
    Port     uint16
    User     string
    Password string
    Insecure bool
}

//------------------------------------------------------------------------------

func (c *Config) Client() (interface {}, error) {
    switch c.Type {
    case "local":
        log.Printf(`[INFO][terraform-provider-windows] configuring windows-provider
                    [INFO][terraform-provider-windows]     type: %q
`       , c.Type)
    case "ssh":
        log.Printf(`[INFO][terraform-provider-windows] configuring windows-provider
                    [INFO][terraform-provider-windows]     type: %q
                    [INFO][terraform-provider-windows]     host: %q
                    [INFO][terraform-provider-windows]     port: %d
                    [INFO][terraform-provider-windows]     user: %q
                    [INFO][terraform-provider-windows]     password: ********
                    [INFO][terraform-provider-windows]     insecure: %t
`       , c.Type, c.Host, c.Port, c.User, c.Insecure)
    }

    windowsClient := new(api.WindowsClient)
    switch c.Type {
    case "local":
        windowsClient.Type     = c.Type
    case "ssh":
        windowsClient.Type     = c.Type
        windowsClient.Host     = c.Host
        windowsClient.Port     = c.Port
        windowsClient.User     = c.User
        windowsClient.Password = c.Password
        windowsClient.Insecure = c.Insecure
    }

    log.Printf("[INFO][terraform-provider-windows] configured windows-provider\n")
    return windowsClient, nil
}

//------------------------------------------------------------------------------
