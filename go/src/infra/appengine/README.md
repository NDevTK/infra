# Creating a new Infra Go GAE App

If you've never created a GAE app before, it can be helpful to take a look at
the cloud docs: [Quickstart for Go Appengine Standard Environment](https://cloud.google.com/appengine/docs/standard/go/quickstart)

In luci-go, we have an [example app](https://github.com/luci/luci-go/tree/master/examples/appengine/helloworld_standard)
that contains a lot of boilerplate code. You should probably use this unless
your app is extremely simple.

If you don't use the standard boilerplate, it is still worth [linking
gae.py](https://github.com/luci/luci-go/tree/master/examples/appengine/helloworld_standard#running-on-devserver)
to have a simple way to run the app locally. Once linked, you can use `./gae.py
devserver -A <project>` to run your app locally.
