# One-liner installer for Flux Relay CLI
# Usage: irm https://raw.githubusercontent.com/postacksol/flux-relay-cli/main/install.ps1 | iex
# Or:    iwr https://raw.githubusercontent.com/postacksol/flux-relay-cli/main/install.ps1 -OutFile install.ps1; .\install.ps1

& ([scriptblock]::Create((New-Object Net.WebClient).DownloadString('https://raw.githubusercontent.com/postacksol/flux-relay-cli/main/install.ps1')))
