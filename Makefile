REPO_URL := quay.io/jpacker

all:
	@echo "Commands:"
	@echo ""
	@echo "make build     # Build image ONLY"
	@echo "make push      # Build and push the image used by manual and cronjobs"
	@echo "make lint      # Validate the source code"
	@echo "make running   # Manually launch Running"
	@echo "make hibernate # Manually launch Hibernating"
	@echo "make setup     # Deploys the cronjobs"

build:
	docker build . -t ${REPO_URL}/hibernation-curator:latest

push: build
	docker push ${REPO_URL}/hibernation-curator:latest

clean:
	docker image rm ${REPO_URL}/hibernation-curator:latest

running:
	oc create -f deploy/running-job.yaml

hibernate:
	oc create -f deploy/hibernation-job.yaml

setup:
	oc apply -k deploy/
