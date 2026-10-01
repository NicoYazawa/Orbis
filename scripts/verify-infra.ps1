$ErrorActionPreference='Stop'
& node (Join-Path $PSScriptRoot 'orbis.mjs') verify '--profiles=app,rag'
exit $LASTEXITCODE
