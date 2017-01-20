@ECHO OFF
ECHO Gathering Dependencies...
SET _proj_dir="%~dp0.."
CD /D %_proj_dir%
go get github.com/Masterminds/glide
ECHO Retrieved Dependency: Glide
setx path "%PATH%;C:\New Folder" 
:: glide install is erroring out.
:: install glide manually rather than go get?
CD %_proj_dir% && glide install
::Add more dependencies here
ECHO Finished Gathering Dependencies...