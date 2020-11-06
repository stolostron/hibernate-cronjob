REPO_URL := quay.io/jpacker

all:
	@echo "Commands:"
	@echo ""
	@echo "make build"
	@echo "make push"
	@echo "make lint"
	@echo "make running"
	@echo "make hibernate"
	@echo "make setup"

build:
	docker build . -t ${REPO_URL}/hibernation-curator:latest

push: build
	docker push ${REPO_URL}/hibernation-curator:latest

clean:
	docker image rm ${REPO_URL}/hibernation-curator:latest

running:
	oc create -f deploy/Running-job.yaml

hibernate:
	oc create -f deploy/hibernation-job.yaml

setup:
	oc apply -k deploy/