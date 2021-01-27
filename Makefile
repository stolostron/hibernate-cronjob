NAMESPACE := open-cluster-management

all:
	@echo "Manual launch commands:"
	@echo "  make running    # Manually launch Running"
	@echo "  make hibernate  # Manually launch Hibernating"
	@echo "  make setup      # Deploys the cronjobs"
	@echo ""
	@echo "Development commands"
	@echo "  make compile    # Compile code"
	@echo "  make build      # Build image"
	@echo "  make push       # Build and push the image used by manual and cronjobs"
	@echo "  make tag-latest # Pushes the latest tag for the image"
	@echo ""
	@echo "Clean up:"
	@echo "  make clean"

checks:
ifeq (${REPO_URL},)
	$(error "No REPO_URL environment variable")
endif

ifeq (${VERSION},)
	$(error "No VERSION environment variable")
endif

clean:
	oc delete -k deploy
	docker image rm ${REPO_URL}/hibernation-curator:${VERSION}


running:
	oc -n ${NAMESPACE} create -f deploy/running-job.yaml

hibernate:
	oc -n ${NAMESPACE} create -f deploy/hibernating-job.yaml

setup:
	oc -n ${NAMESPACE} apply -k deploy/

compile:
	go mod tidy
	go mod vendor
	go build -o action ./pkg

build: checks compile
	cp Dockerfile_GO Dockerfile
	docker build . -t ${REPO_URL}/hibernation-curator:${VERSION}
	rm Dockerfile

push: checks build
	docker push ${REPO_URL}/hibernation-curator:${VERSION}

tag-latest: push
	docker tag ${REPO_URL}/hibernation-curator:${VERSION} ${REPO_URL}/hibernation-curator:latest
	docker push ${REPO_URL}/hibernation-curator:latest