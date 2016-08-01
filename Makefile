IMAGE_VERSION=0.1.0

build-and-package: compile-linux build-image
build-deploy-dev: compile-linux build-image push-to-dev deploy-dev-image

compile-linux:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w' -o logpull

build-image:
	docker build -t thirtyx/logpull .

push-to-dev:
	docker tag -f thirtyx/logpull thirtyx/logpull:dev
	docker push thirtyx/logpull:dev

push-new-version:
	docker tag -f thirtyx/logpull thirtyx/logpull:$(IMAGE_VERSION)
	docker push thirtyx/logpull:$(IMAGE_VERSION)

deploy-dev-image:
	kubectl create -f logpull-dev.yaml  --namespace=apigee