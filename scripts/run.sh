#!/bin/bash
set -e
minikube start
eval $(minikube docker-env)

docker build -t menloltd/indigo-server:latest ./apps/indigo-api-gateway

helm dependency update ./charts/indigo-server
helm install indigo-server ./charts/indigo-server --set gateway.image.tag=latest

kubectl port-forward svc/indigo-server-indigo-api-gateway 8080:8080
# helm uninstall indigo-server
# check http://localhost:8080/api/swagger/index.html#/