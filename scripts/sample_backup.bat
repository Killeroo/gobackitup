@echo off

echo.
echo ----------------------------
echo Avaliable Drives
echo ----------------------------
powershell -C "get-wmiobject win32_logicaldisk | Foreach {\"[\" + $_.DeviceID + \"] [\" + $_.volumename + \"] used/size [\" + [math]::truncate(($_.size - $_.freespace) / 1GB) + \"GB/\" + [math]::truncate($_.size / 1GB) + \"GB] \"}"
echo ----------------------------
set /p choice=Drive:
%cd%\gobackitup.exe --source %choice%:\ --destination C:\Backup --name "CETA_BACKUP" --zip
%cd%\gobackitup.exe --source %choice%:\ --destination F:\Backup --name "CETA_BACKUP" --zip
%cd%\gobackitup.exe --source %choice%:\ --destination E:\Backup --name "CETA_BACKUP" --zip
pause