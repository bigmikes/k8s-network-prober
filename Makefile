IMG ?= kube-net-prober:latest

docker-build:
	docker build -t ${IMG} . --target production
docker-run: 
	docker run --publish 8080:8080 ${IMG}
docker-publish: 
	docker push ${IMG}
