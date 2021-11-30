all:
	@echo "Manual launch commands:"
	@echo "  make params      # Display configuration options"
	@echo "  make options     # Display the override options.env"
	@echo "  make running     # Manually launch Running"
	@echo "  make hibernating # Manually launch Hibernating"
	@echo "  make cronjobs    # Deploys the CronJobs"
	@echo "  make roles       # Deploys the ClusterRole and ClusterRoleBinding"
	@echo ""
	@echo "Development commands"
	@echo "  make compile    # Compile code"
	@echo "  make build      # Build image"
	@echo "  make push       # Build and push the image used by manual and cronjobs"
	@echo "  make tag-latest # Pushes the latest tag for the image"
	@echo ""
	@echo "Clean up:"
	@echo "  make clean          # Deletes image from registry"
	@echo "  make clean-cronjobs # Deletes the CronJobs"
	@echo "  make clean-roles    # Deletes the ClusterRole and ClusterRoleBinding"

checks:
ifeq (${REPO_URL},)
	$(error "No REPO_URL environment variable")
endif

ifeq (${VERSION},)
	$(error "No VERSION environment variable")
endif

options.env:
	touch options.env

options:
	@cat ./options.env
	@echo ""

params:
	oc process -f templates/cronjobs.yaml --parameters

clean: checks
	podman image rm ${REPO_URL}/hibernation-curator:${VERSION}

running: options.env
	oc process -f templates/running-job.yaml --param-file options.env --ignore-unknown-parameters=true | oc apply -f -

hibernating: options.env
	oc process -f templates/hibernating-job.yaml --param-file options.env --ignore-unknown-parameters=true  | oc apply -f -

cronjobs: options.env
	oc process -f templates/cronjobs.yaml --param-file options.env --ignore-unknown-parameters=true  | oc apply -f -

clean-cronjobs: options.env
	oc process -f templates/cronjobs.yaml --param-file options.env --ignore-unknown-parameters=true  | oc delete -f -

roles: options.env
	oc process -f templates/roles.yaml --param-file options.env --ignore-unknown-parameters=true | oc apply -f -

clean-roles: options.env
	oc process -f templates/roles.yaml --param-file options.env --ignore-unknown-parameters=true | oc delete -f -

compile:
	go mod tidy
	go mod vendor
	go build -o action ./pkg

build: checks
	podman build -f Dockerfile.prow . -t ${REPO_URL}/hibernation-curator:${VERSION}

push: checks build
	podman push ${REPO_URL}/hibernation-curator:${VERSION}

tag-latest: push
	podman tag ${REPO_URL}/hibernation-curator:${VERSION} ${REPO_URL}/hibernation-curator:latest
	podman push ${REPO_URL}/hibernation-curator:latest