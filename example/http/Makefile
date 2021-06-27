.PHONY: run
run:
	go run ./main.go -debug

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