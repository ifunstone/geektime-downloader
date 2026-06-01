$ErrorActionPreference = "Stop"

$projectRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$outputDir = Join-Path $projectRoot "bin"
$outputFile = Join-Path $outputDir "geektime-downloader.exe"
$defaultGccPath = "C:\Users\skyma\AppData\Local\Microsoft\WinGet\Packages\BrechtSanders.WinLibs.POSIX.UCRT_Microsoft.Winget.Source_8wekyb3d8bbwe\mingw64\bin\gcc.exe"

Write-Host "==> Checking build environment"

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    throw "Go is not installed or not in PATH."
}

$gccCommand = Get-Command gcc -ErrorAction SilentlyContinue
$compiler = $null
if ($gccCommand) {
    $compiler = $gccCommand.Source
} elseif (Test-Path $defaultGccPath) {
    $compiler = $defaultGccPath
}

if (-not $compiler) {
    throw "gcc.exe is not installed or not in PATH. Fyne desktop builds require a C compiler."
}

if (-not (Test-Path $outputDir)) {
    New-Item -ItemType Directory -Path $outputDir | Out-Null
}

Push-Location $projectRoot
try {
    Write-Host "==> Building $outputFile"
    $env:CGO_ENABLED = "1"
    $env:CC = $compiler
    go build -trimpath -ldflags "-s -w" -o $outputFile .
    Write-Host "==> Build complete: $outputFile"
}
finally {
    Pop-Location
}
