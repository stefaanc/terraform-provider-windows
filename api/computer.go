//
// Copyright (c) 2019 Stefaan Coussement
// MIT License
//
// more info: https://github.com/stefaanc/terraform-provider-windows
//
package api

import (
    "bytes"
    "errors"
    "fmt"
    "encoding/json"
    "log"
    "strings"

    "github.com/stefaanc/golang-exec/runner"
    "github.com/stefaanc/golang-exec/script"
)

//------------------------------------------------------------------------------

type Computer struct {
    Name                   string
    NewName                string

    DNSClient              ComputerDNSClient

    RebootPending          bool
    RebootPendingDetails   ComputerRebootPendingDetails

    NetworkAdapterNames    []string
    NetworkConnectionNames []string
}

type ComputerDNSClient struct {
    SuffixSearchList []string
    EnableDevolution bool
    DevolutionLevel  uint32
}

type ComputerRebootPendingDetails struct {
    RebootRequired        bool
    PostRebootReporting   bool
    DVDRebootSignal       bool
    RebootPending         bool
    RebootInProgress      bool
    PackagesPending       bool
    ServicesPending       bool
    UpdateExeVolatile     bool
    ComputerRenamePending bool
    FileRenamePending     bool
    NetlogonPending       bool
    CurrentRebootAttemps  bool
}

//------------------------------------------------------------------------------

func (c *WindowsClient) ReadComputer() (cProperties *Computer, err error) {
    return readComputer(c)
}

//------------------------------------------------------------------------------

func (c *WindowsClient) UpdateComputer(cProperties *Computer) error {
    return updateComputer(c, cProperties)
}

//------------------------------------------------------------------------------

func readComputer(c *WindowsClient) (cProperties *Computer, err error) {
    // create buffer to capture stdout & stderr
    var stdout bytes.Buffer
    var stderr bytes.Buffer

    // run script
    c.Lock.Lock()
    err = runner.Run(c, readComputerScript, nil, &stdout, &stderr)
    c.Lock.Unlock()
    if err != nil {
        var runnerErr runner.Error
        errors.As(err, &runnerErr)
        log.Printf("[ERROR][terraform-provider-windows/api/readComputer()] cannot read computer\n")
        log.Printf("[ERROR][terraform-provider-windows/api/readComputer()] script exitcode: %d", runnerErr.ExitCode())
        log.Printf("[ERROR][terraform-provider-windows/api/readComputer()] script stdout: \n%s", stdout.String())
        log.Printf("[ERROR][terraform-provider-windows/api/readComputer()] script stderr: \n%s", stderr.String())

        // get to the cause of a "runner failed" error to display in terraform UI
        if strings.Contains(runnerErr.Error(), "runner failed") {
            err = fmt.Errorf("[terraform-provider-windows/api/readComputer()] runner: %s", stderr.String())
        }

        return nil, err
    }
    log.Printf("[INFO][terraform-provider-windows/api/readComputer()] read computer \n%s", stdout.String())

    // convert stdout-JSON to properties
    cProperties = new(Computer)
    err = json.Unmarshal(stdout.Bytes(), cProperties)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-windows/api/readComputer()] cannot convert json to 'cProperties' for computer\n")
        return nil, err
    }

    return cProperties, nil
}

var readComputerScript = script.New("readComputer", "powershell", `
    $ErrorActionPreference = 'Stop'
    $ProgressPreference = 'SilentlyContinue'   # progress-bar fails when using ssh

    $cProperties = @{
        Name                   = $env:ComputerName
        NewName                = ( Get-ItemProperty -Path 'HKLM:\SYSTEM\CurrentControlSet\Control\ComputerName\ComputerName' -Name 'ComputerName' -ErrorAction Ignore ).ComputerName
        DNSClient              = @{}
        RebootPending          = $false
        RebootPendingDetails   = @{
            RebootRequired        = $false
            PostRebootReporting   = $false
            DVDRebootSignal       = $false
            RebootPending         = $false
            RebootInProgress      = $false
            PackagesPending       = $false
            ServicesPending       = $false
            UpdateExeVolatile     = $false
            ComputerRenamePending = $false
            FileRenamePending     = $false
            NetlogonPending       = $false
            CurrentRebootAttemps  = $false
        }
        NetworkAdapterNames    = @()
        NetworkConnectionNames = @()
    }

    $settings = Get-DnsClientGlobalSetting -ErrorAction Ignore
    if ( $settings ) {
        $cProperties.DNSClient = @{
            SuffixSearchList = $settings.SuffixSearchList
            EnableDevolution = $settings.UseDevolution
            DevolutionLevel  = $settings.DevolutionLevel
        }
    }

    if ( Get-Item -Path 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\WindowsUpdate\Auto Update\RebootRequired' -ErrorAction Ignore ) {
        $cProperties.RebootPendingDetails.RebootRequired = $true
        $cProperties.RebootPending = $true
    }
    if ( Get-Item -Path 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\WindowsUpdate\Auto Update\PostRebootReporting' -ErrorAction Ignore ) {
        $cProperties.RebootPendingDetails.PostRebootReporting = $true
        $cProperties.RebootPending = $true
    }
    if ( Get-ItemProperty -Path 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\RunOnce' -Name 'DVDRebootSignal' -ErrorAction Ignore ) {
        $cProperties.RebootPendingDetails.DVDRebootSignal = $true
        $cProperties.RebootPending = $true
    }
    if ( Get-Item -Path 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Component Based Servicing\RebootPending' -ErrorAction Ignore ) {
        $cProperties.RebootPendingDetails.RebootPending = $true
        $cProperties.RebootPending = $true
    }
    if ( Get-Item -Path 'HKLM:\Software\Microsoft\Windows\CurrentVersion\Component Based Servicing\RebootInProgress' -ErrorAction Ignore ) {
        $cProperties.RebootPendingDetails.RebootInProgress = $true
        $cProperties.RebootPending = $true
    }
    if ( Get-Item -Path 'HKLM:\Software\Microsoft\Windows\CurrentVersion\Component Based Servicing\PackagesPending' -ErrorAction Ignore ) {
        $cProperties.RebootPendingDetails.PackagesPending = $true
        $cProperties.RebootPending = $true
    }
    if ( Get-ChildItem -Path 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\WindowsUpdate\Services\Pending' -ErrorAction Ignore ) {
        $cProperties.RebootPendingDetails.ServicesPending = $true
        $cProperties.RebootPending = $true
    }
    if ( ( $v = Get-ItemProperty -Path 'HKLM:\SOFTWARE\Microsoft\Updates' -Name 'UpdateExeVolatile' -ErrorAction Ignore | Select-Object -ExpandProperty 'UpdateExeVolatile' ) -and ( $v -ne 0 ) ) {
        $cProperties.RebootPendingDetails.UpdateExeVolatile = $true
        $cProperties.RebootPending = $true
    }
    if ( ( Get-ItemProperty -Path 'HKLM:\SYSTEM\CurrentControlSet\Control\ComputerName\ComputerName' -Name 'ComputerName' -ErrorAction Ignore ).ComputerName -ne $cProperties.Name ) {
        $cProperties.RebootPendingDetails.ComputerRenamePending = $true
        $cProperties.RebootPending = $true
    }
    if (
        ( ( $v = Get-ItemProperty -Path 'HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager' -Name 'PendingFileRenameOperations'  -ErrorAction Ignore | Select-Object -ExpandProperty 'PendingFileRenameOperations'  ) -and $v ) -or
        ( ( $v = Get-ItemProperty -Path 'HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager' -Name 'PendingFileRenameOperations2' -ErrorAction Ignore | Select-Object -ExpandProperty 'PendingFileRenameOperations2' ) -and $v )
    ) {
        $cProperties.RebootPendingDetails.FileRenamePending = $true
        $cProperties.RebootPending = $true
    }
    if (
        ( Get-ItemProperty -Path 'HKLM:\SYSTEM\CurrentControlSet\Services\Netlogon' -Name 'JoinDomain'  -ErrorAction Ignore ) -or
        ( Get-ItemProperty -Path 'HKLM:\SYSTEM\CurrentControlSet\Services\Netlogon' -Name 'AvoidSpnSet' -ErrorAction Ignore )
    ) {
        $cProperties.RebootPendingDetails.NetlogonPending = $true
        $cProperties.RebootPending = $true
    }
    if ( Get-Item -Path 'HKLM:\SOFTWARE\Microsoft\ServerManager\CurrentRebootAttemps' -ErrorAction Ignore ) {
        $cProperties.RebootPendingDetails.CurrentRebootAttemps = $true
        $cProperties.RebootPending = $true
    }

    Get-NetAdapter -ErrorAction 'Ignore' | foreach {
        $cProperties.NetworkAdapterNames += $_.Name
    }
    Get-NetConnectionProfile -ErrorAction 'Ignore' | foreach {
        $cProperties.NetworkConnectionNames += $_.Name
    }

    Write-Output $( ConvertTo-Json -InputObject $cProperties -Depth 100 )
`)

//------------------------------------------------------------------------------

func updateComputer(c *WindowsClient, cProperties *Computer) error {
    // convert properties to JSON
    cPropertiesJSON, err := json.Marshal(cProperties)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-windows/api/updateComputer(cProperties)] cannot cannot convert 'cProperties' to json for computer\n")
        return err
    }

    // create buffer to capture stdout & stderr
    var stdout bytes.Buffer
    var stderr bytes.Buffer

    // run script
    c.Lock.Lock()
    err = runner.Run(c, updateComputerScript, updateComputerArguments{
        CPropertiesJSON: string(cPropertiesJSON),
    }, &stdout, &stderr)
    c.Lock.Unlock()
    if err != nil {
        var runnerErr runner.Error
        errors.As(err, &runnerErr)
        log.Printf("[ERROR][terraform-provider-windows/api/updateComputer()] cannot update computer\n")
        log.Printf("[ERROR][terraform-provider-windows/api/updateComputer()] script exitcode: %d", runnerErr.ExitCode())
        log.Printf("[ERROR][terraform-provider-windows/api/updateComputer()] script stdout: \n%s", stdout.String())
        log.Printf("[ERROR][terraform-provider-windows/api/updateComputer()] script stderr: \n%s", stderr.String())

        // get to the cause of a "runner failed" error to display in terraform UI
        if strings.Contains(runnerErr.Error(), "runner failed") {
            err = fmt.Errorf("[terraform-provider-windows/api/updateComputer()] runner: %s", stderr.String())
        }

        return err
    }
    log.Printf("[INFO][terraform-provider-windows/api/updateComputer()] updated computer \n%s", stdout.String())

    return nil
}

type updateComputerArguments struct{
    CPropertiesJSON string
}

var updateComputerScript = script.New("updateComputer", "powershell", `
    $ErrorActionPreference = 'Stop'
    $ProgressPreference = 'SilentlyContinue'   # progress-bar fails when using ssh

    function catchExit {
        param($returnValue)

        if ( $returnValue -ne 0 ) {
            $script  = "updateComputer"
            $command_start  = $MyInvocation.OffsetInLine
            $command_length = $MyInvocation.PositionMessage.Split("`+"`"+`n")[-1].Replace('+', '').Replace(' ', '').Length
            $command = $MyInvocation.Line.Substring($command_start - 1, $command_length)
            $lineno  = $MyInvocation.ScriptLineNumber
            $charno  = $MyInvocation.OffsetInLine
            if ( $returnValue -eq 87 ) {
                $message = "invalid new computer-name"
            }
            else {
                $message = "WMI execution failed with ReturnValue $returnValue"
            }

            $text = "ERROR: $exitcode, script: $script, line: $lineno, char: $charno, cmd: '$command' > `+"`"+`"$message`+"`"+`""
            Write-Error $text
        }
    }

    $cProperties = ConvertFrom-Json -InputObject '{{.CPropertiesJSON}}'

    $pendingName = ( Get-ItemProperty -Path 'HKLM:\SYSTEM\CurrentControlSet\Control\ComputerName\ComputerName' -Name 'ComputerName' -ErrorAction Ignore ).ComputerName
    # remark that $pendingName is different from the current $env:ComputerName when there is a reboot pending because of a previous computer-name change
    if ( ( $cProperties.NewName -ne "" ) -and ( $cProperties.NewName -ne $pendingName ) ) {
        # Rename-Computer doesn't allow you to change the pending computer-name back to the current $env:ComputerName after a previous computer-name change - using WMI does allow this
        $returnValue = ( Invoke-WmiMethod -Name 'Rename' -Path "Win32_ComputerSystem.Name='$env:ComputerName'" -ArgumentList "$( $cProperties.NewName )" ).ReturnValue; catchExit $returnValue
    }

    $arguments = @{
        SuffixSearchList = $cProperties.DNSClient.SuffixSearchList
        UseDevolution    = $cProperties.DNSClient.EnableDevolution
        DevolutionLevel  = $cProperties.DNSClient.DevolutionLevel
    }
    Set-DnsClientGlobalSetting @arguments -Confirm:$false | Out-Default
`)

//------------------------------------------------------------------------------
