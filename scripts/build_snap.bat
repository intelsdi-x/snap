@ECHO OFF
ECHO Building SNAP...
SET _proj_dir="%~dp0.."
for /f "tokens=1-3" %%i in ('git --version') do SET git_version=%%k
git --version
ECHO %git_version%
go build -ldflags "-w -X main.gitversion=%git_version%"