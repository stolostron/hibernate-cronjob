NAMESPACE := open-cluster-management
all:
	@echo "Commands:"
	@echo ""
	@echo "make build     # Build image ONLY"
	@echo "make push      # Build and push the image used by manual and cronjobs"
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
	oc -n ${NAMESPACE} create -f deploy/running-job.yaml

hibernate:
	oc -n ${NAMESPACE} create -f deploy/hibernation-job.yaml

setup:
	oc -n ${NAMESPACE} apply -k deploy/
