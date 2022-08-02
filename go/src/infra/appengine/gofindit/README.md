# GoFindit
GoFindit is the culprit finding service for compile and test failures for Chrome and ChromeOS.

This is the rewrite in Golang of the Python2 version of Findit (findit-for-me.appspot.com).

## Local Development
To run the server locally, firstly you need to authenticate
```
gcloud auth application-default login
```
and
```
luci-auth login -scopes "https://www.googleapis.com/auth/cloud-platform https://www.googleapis.com/auth/userinfo.email"
```

### Building the Frontend
In another terminal window, build the project with watch for development:
```
cd frontend/ui
npm run watch
```
This will build the React app. If left running, local changes to the React app
will trigger re-building automatically.

To run the frontend unit tests,
```
cd frontend/ui
npm test
```

### Running GoFindit
In the root gofindit directory, run
```
go run main.go
```

This will start a web server running at http://localhost:8800. Navigate to this URL using your preferred browser. Once you "log in", the GoFindit frontend
should load.
