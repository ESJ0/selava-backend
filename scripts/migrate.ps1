$ErrorActionPreference = 'Stop'
$container = if ($env:POSTGRES_CONTAINER) { $env:POSTGRES_CONTAINER } else { 'lavanderia-db' }
$user = if ($env:DB_USER) { $env:DB_USER } else { 'postgres' }
$database = if ($env:DB_NAME) { $env:DB_NAME } else { 'lavanderia' }

Get-ChildItem "$PSScriptRoot\..\migrations\*.up.sql" |
    Sort-Object Name |
    ForEach-Object {
        Write-Host "Aplicando $($_.Name)"
        Get-Content -Raw $_.FullName | docker exec -i $container psql -v ON_ERROR_STOP=1 -U $user -d $database
        if ($LASTEXITCODE -ne 0) { throw "Falló $($_.Name)" }
    }
