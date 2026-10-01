param([switch]$Dev)
$argumentsList=@((Join-Path $PSScriptRoot 'orbis.mjs'),'stop','--profiles=app,rag,obs,ollama,search')
if($Dev){$argumentsList+='--dev'}
& node @argumentsList
exit $LASTEXITCODE
