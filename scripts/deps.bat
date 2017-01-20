@echo off
echo Gathering Dependencies...
set _proj_dir="%~dp0.."
cd %_proj_dir%
go get github.com/Masterminds/glide
echo Retrieved Dependency: Glide
glide install
echo Finished Gathering Dependencies...
