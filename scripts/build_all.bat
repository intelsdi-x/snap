@ECHO OFF
ECHO Ready Go!
::Is throwing an error, so I'm moving on to build_snap.bat
::CALL "%~dp0\deps.bat"
CALL "%~dp0\build_snap.bat"
ECHO The end
PAUSE
EXIT /B 0