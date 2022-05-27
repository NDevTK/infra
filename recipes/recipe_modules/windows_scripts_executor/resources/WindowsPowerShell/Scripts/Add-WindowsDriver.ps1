<#
  .SYNOPSIS
    Install a windows driver to the given mount location of an image

  .DESCRIPTION
    Runs Add-WindowsDriver command in Powershell, collects the output of the
    command and returns the output in json format.

  .PARAMETER Driver
    The absolute path to the location of the driver. Valid values are
    1. A single .inf file
    2. A folder that containing one or more .inf file in its structure

  .PARAMETER Path
    Specifies the full path to the root directory of the offline Windows
    image that you will service

  .PARAMETER LogPath
    The path to the log file. To be used to log the STDERR from the command exec

  .PARAMETER LogLevel
    Log level to be used for recording the output.

  .PARAMETER Recurse
    Search for Drivers to load recursively starting a the path provided to the Driver param

  .EXAMPLE
    Install vmware.inf to an image mounted at C:\b\mount

    PS> $driver_path = C:\b\cache\cipd\infra\tools\test\vmware.inf
    PS> $path = C:\b\mount
    PS> $log_path = C:\b\sp\awp.log
    PS> $log_level = 2

    PS> Add-WindowsDriver -Driver $driver_path -Path $path -LogPath $log_path -LogLevel $log_level

    .EXAMPLE
    Install multiple driver in a folder to an image mounted at C:\b\mount

    PS> $driver_path = C:\b\cache\cipd\infra\tools\test
    PS> $path = C:\b\mount
    PS> $log_path = C:\b\sp\awp.log
    PS> $log_level = 2

    PS> Add-WindowsDriver -Driver $driver_path -Path $path -LogPath $log_path -LogLevel $log_level -Recurse

#>

[cmdletbinding()]
param (
  $Driver,

  $Path,

  $LogPath,

  $LogLevel,

  [Switch]$Recurse
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

  if ('Driver' -notin $PSBoundParameters.keys) {
    throw 'Driver was not provided and is a REQUIRED parameter'
  }

  if ('Path' -notin $PSBoundParameters.keys) {
    throw 'Path was not provided and is a REQUIRED parameter'
  }

  if ('LogPath' -notin $PSBoundParameters.keys) {
    throw 'LogPath was not provided and is a REQUIRED parameter'
  }

  if ('LogLevel' -notin $PSBoundParameters.keys) {
    # use log level as 2 by default
    $LogLevel = '2'
  }

  $params = @{
    'Driver' = $Driver
    'Path' = $Path
    'LogPath' = $LogPath
    'LogLevel' = $LogLevel
  }

  if ($Recurse) {
    $params.add('Recurse', $Recurse)
  }

  $invoke_obj.Output = Add-WindowsDriver @params

  if ($invoke_obj.Output.gettype().Name -eq 'ImageObject') {
    throw "Failed to add drivers from $Driver. Returned img obj."
  }

  $cat_files = (Get-ChildItem -Path $Driver -Filter *.cat -Recurse).Name

  $compare_object_params = @{
    'ReferenceObject'  = $cat_files
    'DifferenceObject' = $invoke_obj.Output.CatalogFile
    'IncludeEqual'     = $true
  }
  $comparison = Compare-Object @compare_object_params

  if ($comparison.SideIndicator.contains('<=')) {
    $unloaded_drivers = ($comparison | Where-Object {$_.SideIndicator -eq '<='}).InputObject
    throw "Failed to add all drivers. Missing drivers:`n$($unloaded_drivers -join '`n')"
  }

  $json = $invoke_obj | ConvertTo-Json -Compress -Depth 100 -ErrorAction Stop

} catch {
  $invoke_obj.Success = $false
  $invoke_obj.ErrorInfo.Message = $_.Exception.Message
  $invoke_obj.ErrorInfo.Line = $_.Exception.CommandInvocation.Line
  $invoke_obj.ErrorInfo.PositionMessage = $_.Exception.CommandInvocation.PositionMessage
  $json = $invoke_obj | ConvertTo-Json -Compress -Depth 100
} finally {
  $json
}