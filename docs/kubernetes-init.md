## Installing Traefik as an Ingress Controller in Kubernetes

### 1. Adding the Helm Repository
To install Traefik as an ingress controller, first add the official Traefik Helm repository:
```shell
helm repo add traefik https://helm.traefik.io/traefik  
helm repo update
```

### 2. Pulling the Traefik Chart
Next, pull the Traefik Helm chart and extract it:
```shell
helm pull  --untar traefik/traefik
cd traefik
```
### 3. Configuring Traefik
Create a new values file named `my.values.yaml` with the following custom configuration:
```yaml
ports:
  traefik:
    port: 9000
    expose: true
    exposedPort: 9000
    nodePort: 30300
    protocol: TCP
  web:
    port: 8000
    expose: true
    exposedPort: 80
    nodePort: 30400
    protocol: TCP
service:
  enabled: true
  single: true
  type: NodePort
```
### 4. Deploying Traefik with Custom Values
Apply the configuration using the following command:
```yaml
helm install traefik --namespace=[NAME_SPACE] --values=my.values.yaml .
```
### 5. Accessing Traefik
Once deployed, you should be able to access:
* Traefik Web UI: http://xx.xx.xx.xx:30400
* Traefik Dashboard: http://xx.xx.xx.xx:30300/dashboard/