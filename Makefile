docker:
	docker build -t kube-net-prober . --target production
run: 
	docker run --publish 8080:8080 kube-net-prober
publish: 
	docker image tag kube-net-prober:latest bigmikes/kube-net-prober:test-version
	docker push bigmikes/kube-net-prober:test-version
