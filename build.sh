#!/bin/bash

#Чтение текущей версии из файла
version=$(cat version.txt)

#Разделение версии на основные, второстепенные и патч-номера
major=$(echo $version | cut -d. -f1)
minor=$(echo $version | cut -d. -f2)
patch=$(echo $version | cut -d. -f3)

#Увеличение патч-номера
patch=$((patch + 1))

#Формирование новой версии
new_version="$major.$minor.$patch"

#Запись новой версии в файл
echo $new_version > version.txt

echo "Новая версия: $new_version"
go build -ldflags "-X main.version=$(cat version.txt)" -o filestorage .