@echo off

echo Gathering dependencies

set _proj_dir="%~dp0.."
cd /D %_proj_dir%
go get github.com/Masterminds/glide
glide install

echo Finished gathering dependencies
