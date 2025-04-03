## Kafka on Kubernetes
This guide covers how to deploy Kafka on Kubernetes using Helm and test it with kcat (a non-JVM Kafka producer and consumer).

### Installing Kafka

#### Deploy Kafka using Helm:
```shell
helm install kafka oci://registry-1.docker.io/bitnamicharts/kafka \
  --set auth.enabled=false \
  --set listeners.interbroker.protocol=PLAINTEXT \
  --set listeners.controller.protocol=PLAINTEXT \
  --set listeners.client.protocol=PLAINTEXT
```
This installs Kafka with plaintext communication between brokers, controllers, and clients.

#### Deploying kcat for Testing
kcat is a lightweight Kafka client that can produce and consume messages. Deploy kcat as a Kubernetes pod:

##### 1. Create the deployment file:
```shell
nano kcat-deployment.yaml
```

##### 2. Add the following YAML configuration:
```yaml
kind: Deployment
apiVersion: apps/v1
metadata:
  name: kcat
  labels:
    app: kcat
spec:
  selector:
    matchLabels:
      app: kcat
  template:
    metadata:
      labels:
        app: kcat
    spec:
      containers:
        - name: kcat
          image: edenhill/kcat:1.7.0
          command: ["/bin/sh"]
          args: ["-c", "trap : TERM INT; sleep 1000 & wait"]
```

##### 3. Apply the deployment:
```shell
kubectl apply -f kcat-deployment.yaml
```

##### 4. Start an interactive shell session inside the kcat pod:
```shell
kubectl exec --stdin --tty [pod-name] -- /bin/sh
```
#### 5. Sending a Test Message
Enter the command below to send Kafka a test message to ingest:
```shell
echo "Test Message" | kcat -P -b kafka:9092 -t testtopic -p -1
```
If successful, the command prints no output.

#### 6. Consuming Messages
Switch to the consumer role and query Kafka for messages by typing:
```shell
kcat -C -b kafka:9092 -t testtopic -p -1
```
You should see `Test Message` printed in the output.

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