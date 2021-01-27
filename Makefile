all:
	@echo "Manual launch commands:"
	@echo "  make params     # Display configuration options"
	@echo "  make running    # Manually launch Running"
	@echo "  make hibernate  # Manually launch Hibernating"
	@echo "  make cronjobs   # Deploys the CronJobs"
	@echo "  make roles      # Deploys the ClusterRole and ClusterRoleBinding"
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

params:
	oc process -f templates/cronjobs.yaml --parameters

clean: checks
	docker image rm ${REPO_URL}/hibernation-curator:${VERSION}

running: options.env
	oc process -f templates/running-job.yaml --param-file options.env

hibernating: options.env
	oc process -f templates/hibernating-job.yaml --param-file options.env

cronjobs: options.env
	oc process -f templates/cronjobs.yaml --param-file options.env

clean-cronjobs: options.env
	oc process -f templates/cronjobs.yaml --param-file options.env

roles: options.env
	oc process -f templates/roles.yaml -p NAMESPACE=`oc project -q`

clean-roles: options.env
	oc process -f templates/roles.yaml -p NAMESPACE=`oc project -q`

compile:
	go mod tidy
	go mod vendor
	go build -o action ./pkg

build: checks compile
	docker build . -t ${REPO_URL}/hibernation-curator:${VERSION}

push: checks build
	docker push ${REPO_URL}/hibernation-curator:${VERSION}

tag-latest: push
	docker tag ${REPO_URL}/hibernation-curator:${VERSION} ${REPO_URL}/hibernation-curator:latest
	docker push ${REPO_URL}/hibernation-curator:latest