.PHONY: run
run:
	go run ./main.go -debug

.PHONY: curl
curl:
	curl localhost:8080 -H "X-Cloud-Trace-Context: 0123456789abcdef0123456789abcdef/123;o=1"
	curl localhost:8080 -H "X-Cloud-Trace-Context: 0123456789abcdef0123456789/123;o=1"

.PHONY: docker-push
docker-push:
	docker build -t asia.gcr.io/$(GOOGLE_CLOUD_PROJECT)/cloud-run-sdk-http-example:latest .
	docker push asia.gcr.io/$(GOOGLE_CLOUD_PROJECT)/cloud-run-sdk-http-example:latest

.PHONY: deploy
deploy:
	gcloud run deploy cloud-run-sdk-http-example \
		--allow-unauthenticated \
		--platform=managed \
		--region=asia-east1 \
		--image=asia.gcr.io/${GOOGLE_CLOUD_PROJECT}/cloud-run-sdk-http-example:latest \
		--max-instances=1 \
		--memory=256Mi \
		--quiet