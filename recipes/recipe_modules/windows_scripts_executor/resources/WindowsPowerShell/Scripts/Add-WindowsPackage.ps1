<#
  .SYNOPSIS
    Install a windows package to the given mount location of image

  .DESCRIPTION
    Runs Add-WindowsPackage command in Powershell, collects the ouput of the
    command and returns the output in json format.

  .PARAMETER PackagePath
    The absolute path to the location of the package. Valid values are
    1. A single .cab or .msu file.
    2. A folder that contains a single expanded .cab file.
    3. A folder that contains a single .msu file.
    4. A folder that contains multiple .cab or .msu files.

  .PARAMETER Path
    Specifies the full path to the root directory of the offline Windows
    image that you will service

  .PARAMETER LogPath
    The path to the log file. To be used to log the STDERR from the command exec

  .PARAMETER LogLevel
    Log level to be used for recording the output.

  .EXAMPLE
    Install winpe_wmi.cab to an image mounted at C:\b\mount

    PS> Add-WindowsPackage -PackagePath C:\b\cache\cipd\infra\tools\test -Path C:\b\mount -LogPath C:\b\sp\awp.log -LogLevel 2

#>

[cmdletbinding()]
param (
  [String]$PackagePath,

  [String]$Path,

  [String]$LogPath,

  [String]$LogLevel
)

# Return object. To be returned as json to STDOUT
$invoke_obj = @{
  'Success' = $true
  'Output' = ''
  'ErrorInfo' = @{
    'Message' = ''
    'Line' = ''
    'PositionMessage' = ''
  }
}

try {
  $params = @{
    'PackagePath' = $PackagePath
    'Path' = $Path
    'LogPath' = $LogPath
    'LogLevel' = $LogLevel
  }

  $invoke_obj.Output = Add-WindowsPackage @params

  $json = $invoke_obj | ConvertTo-Json -Compress -Depth 100 -ErrorAction Stop

} catch {
  $invoke_obj.Success = $false
  $invoke_obj.ErrorInfo.Message = $_.Exception.Message
  $invoke_obj.ErrorInfo.Line = $_.Exception.CommandInvocation.Line
  $invoke_obj.ErrorInfo.PositionMessage = $_.Exception.CommandInvocation.PositionMessage
  $json = $invoke_obj | ConvertTo-Json -Compress -Depth 100
  $json

} finally {
  $json
}
