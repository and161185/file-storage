# FileStorage Service

FileStorage Service — это микросервис для хранения файлов, разработанный на Go и развёрнутый в Kubernetes. Он включает функционал загрузки и скачивания файлов, а также поддерживает конфигурацию через ConfigMap.

## Возможности

- Хранение файлов в MongoDB.
- Управление конфигурацией через Kubernetes ConfigMap.

---

## Как запустить

### 1. Локальный запуск

#### Требования:
- Go >= 1.20
- Конфигурационный файл `config.json`.
- строка подключения к монго указана с учетом развертывания монго в k3s

#### Шаги:
1. Скомпилировать приложение:
   CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o filestorage .

2. Запустить приложение:
   ./filestorage

3. Проверить доступность:
   curl http://localhost:50205/version

### 2. Сборка Docker-образа

#### Шаги:
1. Собрать образ:
   docker build -t filestorage:latest .

2. Запустить:
   docker run -p 50205:50205 filestorage:latest

3. Проверить доступность:
   curl http://localhost:50205/version

### 3. Деплой в Kubernetes

#### Шаги:
1. Создать ConfigMap для конфигурации:
   kubectl create configmap filestorage-config --from-file=config.json --dry-run=client -o yaml | kubectl apply -f -

2. Применить манифесты:
   kubectl apply -f k8s/filestorage/deployment.yaml
   kubectl apply -f k8s/filestorage/service.yaml
   kubectl apply -f k8s/mongodb/mongo-pv.yaml
   kubectl apply -f k8s/mongodb/mongo-pvc.yaml
   kubectl apply -f k8s/mongodb/mongo-deployment.yaml

3. Проверить доступность сервиса:
   kubectl get pods
   kubectl get services

## Конфигурация

Файл config.json:

{
    "database": {
        "uri": "mongodb://<host>:<port>",
        "name": "file_storage",
        "collection": "files",
        "max_pool_size": 64,
        "min_pool_size": 8,
        "max_conn_idle_time_sec": 60
    },
    "server": {
        "port": 50205
    },
    "features": {
        "test": false
    },
    "tokens": {
        "general_token": "your_general_token",
        "download_token": "your_download_token"
    }
}

## Часто используемые команды

1. Обновление ConfigMap
   kubectl create configmap filestorage-config --from-file=config.json --dry-run=client -o yaml | kubectl apply -f -

2. Перезапуск Pod'ов:
   kubectl rollout restart deployment filestorage

## Замечания

- По умолчанию приложение использует порт 50205. При необходимости его можно изменить через config.json.
- Конфигурация управляется через ConfigMap.
- Если приложение запускается вне Kubernetes, убедитесь, что config.json лежит в рабочей директории.