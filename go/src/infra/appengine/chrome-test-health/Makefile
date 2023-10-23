.PHONY: help build clean
FINDIT_CLOUD_PROJECT = findit-for-me-staging
CHROME_TEST_HEALTH_CLOUD_PROJECT = chrome-test-health-staging

help:
	@echo "Available commands:"
	@sed -n '/^[a-zA-Z0-9_]*:/s/:.*//p' <Makefile

build: clean
	cd frontend && npm ci && npm run build && cp -R build ../static

emulator:
	set DATASTORE_DATASET=$(FINDIT_CLOUD_PROJECT)
	set DATASTORE_EMULATOR_HOST=localhost:8081
	set DATASTORE_EMULATOR_HOST_PATH=localhost:8081/datastore
	set DATASTORE_HOST=http://localhost:8081
	set DATASTORE_PROJECT_ID=$(FINDIT_CLOUD_PROJECT)
	gcloud beta emulators datastore start --consistency=1.0

clean:
	rm -rf static && rm -rf frontend/build

backend-dev:
	go run main.go \
	-cloud-project $(CHROME_TEST_HEALTH_CLOUD_PROJECT) \
	-findit-cloud-project ${FINDIT_CLOUD_PROJECT} \
	-auth-service-host chrome-infra-auth-dev.appspot.com \
	-encrypted-cookies-expose-state-endpoint

frontend-dev:
	cd frontend && npm start

runlocal:
	$(MAKE) -j3 emulator backend-dev
