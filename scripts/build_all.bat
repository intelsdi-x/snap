@echo off

echo Starting build

set _scripts_dir=%~dp0

call %_scripts_dir%deps.bat
call %_scripts_dir%build_snap.bat

echo Finished build
