## Backend

Running the server:

```sh
eval `../../../../env.py`
# Needs to be run the first time to set up BigQuery credentials
gcloud auth application-default login
go run main.go
```

This will set up the backend server running on port `8800`

## Frontend

Running the frontend:

```sh
# Run from a different shell from where you evaled env.py because the frontend
# uses a newer version of npm vs what's provided.
cd frontend
npm install
npm start
```

This will set up the frontend client running on port `3000` with an automatic
proxy to the backend server running on `8800`.  To view the UI, go to
[localhost:3000](http://localhost:3000)

Formatting:

```sh
npm run fix
```

## Deployment

```sh
# Use a clean shell so that you can run newer npm
cd frontend
npm run build
# That builds static assets (html, js, css, etc) into frontend/build
# frontend/build is symlinked to /static
cd ..
# Sets up the environment for gae.py to run
eval `../../../../env.py`
./deploy.sh
```

See the latest version at [https://chrome-infra-stats.googleplex.com/](https://chrome-infra-stats.googleplex.com/)
