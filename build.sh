#!/bin/bash

# Проверка наличия прав суперпользователя
if [ "$(id -u)" -ne 0 ]; then
  echo "Пожалуйста, запустите этот скрипт с правами суперпользователя (sudo)."
  exit 1
fi

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
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(cat version.txt)" -o filestorage .

# Сборка Docker образа
docker build -t filestorage:latest .

# Выгрузка Docker образа
docker save filestorage:latest -o filestorage.tar

# Очистка образов
docker image prune -f

# Создание ConfigMap
kubectl create configmap filestorage-config --from-file=config.json --dry-run=client -o yaml | kubectl apply -f -
echo "ConfigMap успешно создан или обновлен."

# Загрузка Docker образа в k3s
k3s ctr image import filestorage.tar

# Обновление подов k3s
kubectl set image deployment/filestorage filestorage=filestorage:latest

# Перезапуск развертывания для применения обновлений
kubectl rollout restart deployment filestorage

echo "Docker образ и Kubernetes поды успешно обновлены."