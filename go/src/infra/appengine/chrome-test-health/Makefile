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
	-cloud-project "chrome-test-health-staging" \
	-findit-cloud-project "findit-for-me" \
	-auth-service-host chrome-infra-auth-dev.appspot.com \
	-encrypted-cookies-expose-state-endpoint \
	-primary-tink-aead-key sm://tink-aead-primary \
	-encrypted-cookies-client-id 961841358-3c87679tdm982jrrpebv04a31bl3ibgm.apps.googleusercontent.com \
	-encrypted-cookies-client-secret sm://oauth-client-secret \
	-encrypted-cookies-redirect-url http://localhost:8800/auth/openid/callback \
	-frontend-client-id 961841358-3c87679tdm982jrrpebv04a31bl3ibgm.apps.googleusercontent.com

frontend-dev:
	cd frontend && npm start

runlocal:
	$(MAKE) -j3 emulator backend-dev
