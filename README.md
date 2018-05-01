# gobackitup

[![](https://img.shields.io/badge/version-1.0-brightgreen.svg)]() [![Go Report Card](https://goreportcard.com/badge/github.com/Killeroo/gobackitup)](https://goreportcard.com/report/github.com/Killeroo/gobackitup)

Small and simple cross platform backup program. Download it [here.](https://github.com/Killeroo/gobackitup/releases)

![](https://user-images.githubusercontent.com/9999745/39470871-012e505c-4d38-11e8-9b97-02556598a7ed.png)

# Usage

    [-s] [-source]       path   Path to backup
    [-d] [-destination]  path   Path to save backup to
    [-n] [-name]         name   Name of folder to save backup to (optional)
    [-z] [-zip]                 Compress the backup (optional)

# Example

    gobackitup --source C:\Some\Thing --destination H:\
    gobackitup --s /Volumes/USB --d /home/ --n MY_BACKUP
    gobackitup --s /Volumes/USB2 --d /home/documents --n MY_ZIP_BACKUP --zip
