## hello-cicd
    About simple project to show:
    1. Deploy gokit to GKE with CICD 

    - docker command - build golang app
    ```
    docker build -t spinliao/hello-cicd .
    docker run -d -p 80:3000 spinliao/hello-cicd
    ```
    
    - gke command - deploy go app on gke
    ```
    gcloud artifacts repositories create hello-cicd-repo --project=go-microservice-370513 --repository-format=docker --location=us-west1 --description="Hello CICD repository"
    gcloud builds submit --tag us-west1-docker.pkg.dev/go-microservice-370513/hello-cicd-repo/hello-cicd-gke .

    gcloud container clusters create-auto hello-cicd-gke --region us-west1

    kubectl apply -f gke/deployment.yaml
    kubectl apply -f gke/service.yaml

    gcloud container clusters delete hello-cicd-gke --region us-west1
    gcloud artifacts docker images delete us-west1-docker.pkg.dev/go-microservice-370513/hello-cicd-repo/hello-cicd-gke
    ```

    2. gRPC
    ```
    cmd to hello-cicd diractory
    protoc .\pb\hellosvc.proto --go_out=plugins=grpc:.
    ```

    3. http2
    ```
    go run 'C:\Program Files\Go/src/crypto/tls/generate_cert.go' --host localhost
    use http.ServeTLS
    ```