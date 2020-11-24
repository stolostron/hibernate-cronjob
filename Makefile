NAMESPACE := open-cluster-management

checks:
ifeq (${REPO_URL},)
	err: ; "No REPO_URL environment variable"
endif

ifeq (${VERSION},)
	err: ; "Specify a VERSION environment variables"
endif

all:
	@echo "Commands:"
	@echo ""
	@echo "make build-py  # Build Python image ONLY"
	@echo "make push-py   # Build and push the Pyhton image used by manual and cronjobs"
	@echo "make build-go  # Build Go image ONLY"
	@echo "make push-go   # Build and push the Go image used by manual and cronjobs"
	@echo "make running   # Manually launch Running"
	@echo "make hibernate # Manually launch Hibernating"
	@echo "make setup     # Deploys the cronjobs"

build-py: checks
	cp Dockerfile_PYTHON Dockerfile
	docker build . -t ${REPO_URL}/hibernation-curator:${VERSION}
	rm Dockerfile

push-py: checks build-py
	docker push ${REPO_URL}/hibernation-curator:${VERSION}

tag-py-latest: push-py
	docker tag ${REPO_URL}/hibernation-curator:${VERSION} ${REPO_URL}/hibernation-curator:latest
	docker push ${REPO_URL}/hibernation-curator:latest

clean:
	docker image rm ${REPO_URL}/hibernation-curator:${VERSION}
	docker image rm ${REPO_URL}/hibernation-curator-go:${VERSION}

running:
	oc -n ${NAMESPACE} create -f deploy/running-job.yaml

hibernate:
	oc -n ${NAMESPACE} create -f deploy/hibernating-job.yaml

setup:
	oc -n ${NAMESPACE} apply -k deploy/

build-go: checks
	go mod tidy
	go mod vendor
	go build -o action ./pkg
	cp Dockerfile_GO Dockerfile
	docker build . -t ${REPO_URL}/hibernation-curator-go:${VERSION}
	rm Dockerfile

push-go: checks build-go
	docker push ${REPO_URL}/hibernation-curator-go:${VERSION}

tag-go-latest: push-go
	docker tag ${REPO_URL}/hibernation-curator-go:${VERSION} ${REPO_URL}/hibernation-curator-go:latest
	docker push ${REPO_URL}/hibernation-curator-go:latest