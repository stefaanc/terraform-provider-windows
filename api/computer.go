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
    Name                 string
    NewName              string

    DNSClient            ComputerDNSClient

    RebootPending        bool
    RebootPendingDetails ComputerRebootPendingDetails
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

func (c *WindowsClient) ReadComputer() (mosProperties *Computer, err error) {
    return readComputer(c)
}

//------------------------------------------------------------------------------

func (c *WindowsClient) UpdateComputer(mosProperties *Computer) error {
    return updateComputer(c, mosProperties)
}

//------------------------------------------------------------------------------

func readComputer(c *WindowsClient) (mosProperties *Computer, err error) {
    // create buffer to capture stdout & stderr
    var stdout bytes.Buffer
    var stderr bytes.Buffer

    // run script
    err = runner.Run(c, readComputerScript, nil, &stdout, &stderr)
    if err != nil {
        var runnerErr runner.Error
        errors.As(err, &runnerErr)
        log.Printf("[ERROR][terraform-provider-windows/api/readComputer()] cannot read management_os\n")
        log.Printf("[ERROR][terraform-provider-windows/api/readComputer()] script exitcode: %d", runnerErr.ExitCode())
        log.Printf("[ERROR][terraform-provider-windows/api/readComputer()] script stdout: \n%s", stdout.String())
        log.Printf("[ERROR][terraform-provider-windows/api/readComputer()] script stderr: \n%s", stderr.String())

        // get to the cause of a "runner failed" error to display in terraform UI
        if strings.Contains(runnerErr.Error(), "runner failed") {
            err = fmt.Errorf("[terraform-provider-windows/api/readComputer()] runner: %s", stderr.String())
        }

        return nil, err
    }
    log.Printf("[INFO][terraform-provider-windows/api/readComputer()] read management_os \n%s", stdout.String())

    // convert stdout-JSON to mosProperties
    mosProperties = new(Computer)
    err = json.Unmarshal(stdout.Bytes(), mosProperties)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-windows/api/readComputer()] cannot convert json to 'mosProperties' for management_os\n")
        return nil, err
    }

    return mosProperties, nil
}

var readComputerScript = script.New("readComputer", "powershell", `
    $ErrorActionPreference = 'Stop'
    $ProgressPreference = 'SilentlyContinue'   # progress-bar fails when using ssh

    $mosProperties = @{
        Name                 = $env:ComputerName
        NewName              = ( Get-ItemProperty -Path 'HKLM:\SYSTEM\CurrentControlSet\Control\ComputerName\ComputerName' -Name 'ComputerName' -ErrorAction Ignore ).ComputerName
        DNSClient            = @{}
        RebootPending        = $false
        RebootPendingDetails = @{
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
    }

    $settings = Get-DnsClientGlobalSetting -ErrorAction Ignore
    if ( $settings ) {
        $mosProperties.DNSClient = @{
            SuffixSearchList = $settings.SuffixSearchList
            EnableDevolution = $settings.UseDevolution
            DevolutionLevel  = $settings.DevolutionLevel
        }
    }

    if ( Get-Item -Path 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\WindowsUpdate\Auto Update\RebootRequired' -ErrorAction Ignore ) {
        $mosProperties.RebootPendingDetails.RebootRequired = $true
        $mosProperties.RebootPending = $true
    }
    if ( Get-Item -Path 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\WindowsUpdate\Auto Update\PostRebootReporting' -ErrorAction Ignore ) {
        $mosProperties.RebootPendingDetails.PostRebootReporting = $true
        $mosProperties.RebootPending = $true
    }
    if ( Get-ItemProperty -Path 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\RunOnce' -Name 'DVDRebootSignal' -ErrorAction Ignore ) {
        $mosProperties.RebootPendingDetails.DVDRebootSignal = $true
        $mosProperties.RebootPending = $true
    }
    if ( Get-Item -Path 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Component Based Servicing\RebootPending' -ErrorAction Ignore ) {
        $mosProperties.RebootPendingDetails.RebootPending = $true
        $mosProperties.RebootPending = $true
    }
    if ( Get-Item -Path 'HKLM:\Software\Microsoft\Windows\CurrentVersion\Component Based Servicing\RebootInProgress' -ErrorAction Ignore ) {
        $mosProperties.RebootPendingDetails.RebootInProgress = $true
        $mosProperties.RebootPending = $true
    }
    if ( Get-Item -Path 'HKLM:\Software\Microsoft\Windows\CurrentVersion\Component Based Servicing\PackagesPending' -ErrorAction Ignore ) {
        $mosProperties.RebootPendingDetails.PackagesPending = $true
        $mosProperties.RebootPending = $true
    }
    if ( Get-ChildItem -Path 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\WindowsUpdate\Services\Pending' -ErrorAction Ignore ) {
        $mosProperties.RebootPendingDetails.ServicesPending = $true
        $mosProperties.RebootPending = $true
    }
    if ( ( $v = Get-ItemProperty -Path 'HKLM:\SOFTWARE\Microsoft\Updates' -Name 'UpdateExeVolatile' -ErrorAction Ignore | Select-Object -ExpandProperty 'UpdateExeVolatile' ) -and ( $v -ne 0 ) ) {
        $mosProperties.RebootPendingDetails.UpdateExeVolatile = $true
        $mosProperties.RebootPending = $true
    }
    if ( ( Get-ItemProperty -Path 'HKLM:\SYSTEM\CurrentControlSet\Control\ComputerName\ComputerName' -Name 'ComputerName' -ErrorAction Ignore ).ComputerName -ne $mosProperties.Name ) {
        $mosProperties.RebootPendingDetails.ComputerRenamePending = $true
        $mosProperties.RebootPending = $true
    }
    if (
        ( ( $v = Get-ItemProperty -Path 'HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager' -Name 'PendingFileRenameOperations'  -ErrorAction Ignore | Select-Object -ExpandProperty 'PendingFileRenameOperations'  ) -and $v ) -or
        ( ( $v = Get-ItemProperty -Path 'HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager' -Name 'PendingFileRenameOperations2' -ErrorAction Ignore | Select-Object -ExpandProperty 'PendingFileRenameOperations2' ) -and $v )
    ) {
        $mosProperties.RebootPendingDetails.FileRenamePending = $true
        $mosProperties.RebootPending = $true
    }
    if (
        ( Get-ItemProperty -Path 'HKLM:\SYSTEM\CurrentControlSet\Services\Netlogon' -Name 'JoinDomain'  -ErrorAction Ignore ) -or
        ( Get-ItemProperty -Path 'HKLM:\SYSTEM\CurrentControlSet\Services\Netlogon' -Name 'AvoidSpnSet' -ErrorAction Ignore )
    ) {
        $mosProperties.RebootPendingDetails.NetlogonPending = $true
        $mosProperties.RebootPending = $true
    }
    if ( Get-Item -Path 'HKLM:\SOFTWARE\Microsoft\ServerManager\CurrentRebootAttemps' -ErrorAction Ignore ) {
        $mosProperties.RebootPendingDetails.CurrentRebootAttemps = $true
        $mosProperties.RebootPending = $true
    }

    Write-Output $( ConvertTo-Json -InputObject $mosProperties -Depth 100 )
`)

//------------------------------------------------------------------------------

func updateComputer(c *WindowsClient, mosProperties *Computer) error {
    // convert properties to JSON
    mosPropertiesJSON, err := json.Marshal(mosProperties)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-windows/api/updateComputer(mosProperties)] cannot cannot convert 'mosProperties' to json for management_os\n")
        return err
    }

    // create buffer to capture stdout & stderr
    var stdout bytes.Buffer
    var stderr bytes.Buffer

    // run script
    err = runner.Run(c, updateComputerScript, updateComputerArguments{
        MOSPropertiesJSON: string(mosPropertiesJSON),
    }, &stdout, &stderr)
    if err != nil {
        var runnerErr runner.Error
        errors.As(err, &runnerErr)
        log.Printf("[ERROR][terraform-provider-windows/api/updateComputer()] cannot update management_os\n")
        log.Printf("[ERROR][terraform-provider-windows/api/updateComputer()] script exitcode: %d", runnerErr.ExitCode())
        log.Printf("[ERROR][terraform-provider-windows/api/updateComputer()] script stdout: \n%s", stdout.String())
        log.Printf("[ERROR][terraform-provider-windows/api/updateComputer()] script stderr: \n%s", stderr.String())

        // get to the cause of a "runner failed" error to display in terraform UI
        if strings.Contains(runnerErr.Error(), "runner failed") {
            err = fmt.Errorf("[terraform-provider-windows/api/updateComputer()] runner: %s", stderr.String())
        }

        return err
    }
    log.Printf("[INFO][terraform-provider-windows/api/updateComputer()] updated management_os \n%s", stdout.String())

    return nil
}

type updateComputerArguments struct{
    MOSPropertiesJSON string
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

    $mosProperties = ConvertFrom-Json -InputObject '{{.MOSPropertiesJSON}}'

    $pendingName = ( Get-ItemProperty -Path 'HKLM:\SYSTEM\CurrentControlSet\Control\ComputerName\ComputerName' -Name 'ComputerName' -ErrorAction Ignore ).ComputerName
    # remark that $pendingName is different from the current $env:ComputerName when there is a reboot pending because of a previous computer-name change
    if ( $pendingName -ne $mosProperties.NewName ) {
        # Rename-Computer doesn't allow you to change the computer-name back to the current $env:ComputerName after a previous computer-name change - using WMI does allow this
        $returnValue = ( Invoke-WmiMethod -Name 'Rename' -Path "Win32_ComputerSystem.Name='$env:ComputerName'" -ArgumentList "$( $mosProperties.NewName )" ).ReturnValue; catchExit $returnValue
    }

    $arguments = @{
        SuffixSearchList = $mosProperties.DNSClient.SuffixSearchList
        UseDevolution    = $mosProperties.DNSClient.EnableDevolution
        DevolutionLevel  = $mosProperties.DNSClient.DevolutionLevel
    }
    Set-DnsClientGlobalSetting @arguments -Confirm:$false | Out-Default
`)

//------------------------------------------------------------------------------
