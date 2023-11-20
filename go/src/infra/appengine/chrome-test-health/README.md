## Backend

Every time a .proto file is changed in /api folder.

```sh
# Change to api folder where proto resides
cd api
# Compiles all the protos again.
go generate
# cd back to root.
cd ..
```


Running the server:

```sh
eval `../../../../env.py`
# Needs to be run the first time to set up BigQuery credentials
gcloud auth application-default login
# Emulator command needs to be run in a separate terminal window
make emulator
make backend-dev
```

This will set up the backend server running on port `8800` and using the
chrome-test-health-staging GCP project for its data.

## Frontend

Running the frontend without the backend:

```sh
cd frontend
npm install
npm start-stg
```

This will set up the frontend client running on port `3000` with a proxy to the
staging environment. For auth to work, you will need to complete the auth flow
and copy over the `LUCISID` cookie from the staging environment into localhost.

To do this, you will need to go to [http://localhost:3000/auth/openid/state](http://localhost:3000/auth/openid/state)
and go to developer tools. Go to Application -> Storage -> Cookes -> http://localhost:3000
and then add the Name:Value pair you obtained completing the auth flow prior.

You can also run the frontend against your local backend using:

```sh
npm start
```

This will proxy the frontend against `localhost:8800`, which should not require
an auth flow.

Before submitting code, make sure to run:

```sh
npm run fix # Reformat the code.
npm test # Run all tests to make sure that nothing breaks.
```

## Deployment

```sh
eval `../../../../env.py`
cd frontend
# Build static assets (html, js, css, etc) into frontend/build
npm run build
cd ..
./deploy.sh
```

See the latest version at [https://chrome-test-health.googleplex.com/](https://chrome-test-health.googleplex.com/)
