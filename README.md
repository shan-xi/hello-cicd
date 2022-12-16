# hello-cicd

## Demo

    1. Using go-kit to build a service with HTTP and gRPC transport.
    2. Integrated with gihub, CircleCI and GKE to show a basic CICD flow
    3. Deploy a loadBalancer type go service on GKE.

## Below is command records

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

    gcloud container clusters list

    gcloud container clusters get-credentials hello-cicd-gke --region us-west1

    kubectl config current-context

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

    4. circleci
    ```
    gcloud auth configure-docker

    docker build -t hello-cicd-gke:v1 .

    docker tag hello-cicd-gke:v1 gcr.io/go-microservice-370513/hello-cicd-gke:v1

    docker push gcr.io/go-microservice-370513/hello-cicd-gke:v1    

    gcloud container clusters create-auto hello-cicd-gke --region us-west1

    kubectl apply --validate=true --dry-run=true -f .\gke\

    kubectl apply --validate=true -f .\gke\
    ```

## Reference

[Simplifying your CI/CD build pipeline to GKE with CircleCI orbs](https://circleci.com/blog/simplifying-your-ci-cd-build-pipeline-to-gke-with-circleci-orbs/)

[Go-kit example with all-features](https://github.com/go-kit/examples/tree/master/addsvc)

