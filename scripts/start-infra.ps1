param([string]$Profiles='app,rag',[switch]$Dev,[switch]$NoBuild)
$ErrorActionPreference='Stop'
$argumentsList=@((Join-Path $PSScriptRoot 'orbis.mjs'),'start',"--profiles=$Profiles")
if($Dev){ $argumentsList+='--dev' }
if($NoBuild){ $argumentsList+='--no-build' }
& node @argumentsList
exit $LASTEXITCODE
