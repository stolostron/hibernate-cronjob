NAMESPACE := open-cluster-management

all: checks
	@echo "Development commands with Python:"
	@echo "  make build-py      # Build Python image ONLY"
	@echo "  make push-py       # Build and push the Pyhton image used by manual and cronjobs"
	@echo "  make tag-py-latest # Pushes the latest tag for the Python based image"
	@echo ""
	@echo "Manual launch commands with Python:"
	@echo "  make running-py    # Manually launch Running"
	@echo "  make hibernate-py  # Manually launch Hibernating"
	@echo "  make setup-py      # Deploys the Python cronjobs"
	@echo ""
	@echo "Development commands with Go"
	@echo "  make compile-go    # Compile Go code ONLY"
	@echo "  make build-go      # Build Go image ONLY"
	@echo "  make push-go       # Build and push the Go image used by manual and cronjobs"
	@echo "  make tag-go-latest # Pushes the latest tag for the Go based image"
	@echo ""
	@echo "Manual launch commands with Go:"
	@echo "  make running-go    # Manually launch Running"
	@echo "  make hibernate-go  # Manually launch Hibernating"
	@echo "  make setup-gp      # Deploys the Go cronjobs"
	@echo ""
	@echo "Clean up:"
	@echo "  make clean-py"
	@echo "  make clean-go"

checks:
ifeq (${REPO_URL},)
	err: ; "No REPO_URL environment variable"
endif

ifeq (${VERSION},)
	err: ; "Specify a VERSION environment variables"
endif

setup:
	@echo "You must choose setup-py or setup-go"

build-py: checks
	cp Dockerfile_PYTHON Dockerfile
	docker build . -t ${REPO_URL}/hibernation-curator:${VERSION}
	rm Dockerfile

push-py: checks build-py
	docker push ${REPO_URL}/hibernation-curator:${VERSION}

tag-py-latest: push-py
	docker tag ${REPO_URL}/hibernation-curator:${VERSION} ${REPO_URL}/hibernation-curator:latest
	docker push ${REPO_URL}/hibernation-curator:latest

clean-py:
	oc delete -k deploy-py
	docker image rm ${REPO_URL}/hibernation-curator:${VERSION}

clean-go:
	oc delete -k deploy-py
	docker image rm ${REPO_URL}/hibernation-curator-go:${VERSION}

running-py:
	oc -n ${NAMESPACE} create -f deploy-py/running-job.yaml

hibernate-py:
	oc -n ${NAMESPACE} create -f deploy-py/hibernating-job.yaml

setup-py:
	oc -n ${NAMESPACE} apply -k deploy-py/

# Go related routines

running-go:
	oc -n ${NAMESPACE} create -f deploy-go/running-job.yaml

hibernate-go:
	oc -n ${NAMESPACE} create -f deploy-go/hibernating-job.yaml

setup-go:
	oc -n ${NAMESPACE} apply -k deploy-go/

compile-go:
	go mod tidy
	go mod vendor
	go build -o action ./pkg

build-go: checks compile-go
	cp Dockerfile_GO Dockerfile
	docker build . -t ${REPO_URL}/hibernation-curator-go:${VERSION}
	rm Dockerfile

push-go: checks build-go
	docker push ${REPO_URL}/hibernation-curator-go:${VERSION}

tag-go-latest: push-go
	docker tag ${REPO_URL}/hibernation-curator-go:${VERSION} ${REPO_URL}/hibernation-curator-go:latest
	docker push ${REPO_URL}/hibernation-curator-go:latest