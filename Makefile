REPO_URL := quay.io/jpacker

all:
	@echo "make commands:"
	@echo "make build"
	@echo "make push"
	@echo "make lint"

build:
	docker build . -t ${REPO_URL}/hibernation-curator:latest

push: build
	docker push ${REPO_URL}/hibernation-curator:latest

clean:
	docker image rm ${REPO_URL}/hibernation-curator:latest