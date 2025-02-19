# FileStorage Service

FileStorage Service is a microservice for file storage, developed in Go and deployed in Kubernetes. It includes functionality for uploading and downloading files and supports configuration through ConfigMap.

## Features

- File storage in MongoDB.
- Configuration management via Kubernetes ConfigMap.

## How to Run

### 1. Local Execution

#### Requirements:
- Go >= 1.20
- Configuration file `config.json`.
- Connection string to MongoDB configured for deployment in k3s.

#### Steps:
1. Compile the application:
   ```bash
   CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o filestorage .
   ```

2. Run the application:
   ```bash
   ./filestorage
   ```

3. Check availability:
   ```bash
   curl http://localhost:50205/version
   ```

### 2. Build a Docker Image

#### Steps:
1. Build the image:
   ```bash
   docker build -t filestorage:latest .
   ```

2. Run the container:
   ```bash
   docker run -p 50205:50205 filestorage:latest
   ```

3. Check availability:
   ```bash
   curl http://localhost:50205/version
   ```

### 3. Deploy in Kubernetes

#### Steps:
1. Create a ConfigMap for configuration:
   ```bash
   kubectl create configmap filestorage-config --from-file=config.json --dry-run=client -o yaml | kubectl apply -f -
   ```

2. Apply manifests:
   ```bash
   kubectl apply -f k8s/mongodb/mongo-pv.yaml
   kubectl apply -f k8s/mongodb/mongo-pvc.yaml
   kubectl apply -f k8s/mongodb/mongo-deployment.yaml
   kubectl apply -f k8s/mongodb/mongo-service.yaml   
   kubectl apply -f k8s/filestorage/deployment.yaml
   kubectl apply -f k8s/filestorage/service.yaml
   ```

3. Check service availability:
   ```bash
   kubectl get pods
   kubectl get services
   ```

## Configuration

Configuration file `config.json`:

```json
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
```

## Frequently Used Commands

1. Update ConfigMap:
   ```bash
   kubectl create configmap filestorage-config --from-file=config.json --dry-run=client -o yaml | kubectl apply -f -
   ```

2. Restart Pods:
   ```bash
   kubectl rollout restart deployment filestorage
   ```

## Notes

- By default, the application uses port 50205. It can be changed via `config.json`.
- Configuration is managed via ConfigMap.
- If running outside Kubernetes, ensure `config.json` is in the working directory.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

