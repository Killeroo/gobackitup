# gobackitup

[![](https://img.shields.io/badge/version-1.0-brightgreen.svg)]()

Small and simple cross platform backup program

# Usage

    [-s] [-source]       path   Path to backup
    [-d] [-destination]  path   Path to save backup to
    [-n] [-name]         name   Name of folder to save backup to (optional)
    [-z] [-zip]                 Compress the backup (optional)

# Example

    gobackitup --source C:\Some\Thing --destination H:\
    gobackitup --s /Volumes/USB --d /home/ --n MY_BACKUP
    gobackitup --s /Volumes/USB2 --d /home/documents --n MY_ZIP_BACKUP --zip
