#!/bin/bash
buildPath="./build"
linuxBuildPath="$buildPath/linux"
rm -rf $buildPath
mkdir -p $linuxBuildPath
fyne package -os linux -icon icon.png
mv appKill.tar.xz $linuxBuildPath