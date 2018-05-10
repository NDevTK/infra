vars = {
  # npm_modules.git is special: we can't check it out on Windows because paths
  # there are too long for Windows. Instead we use 'deps_os' gclient feature to
  # checkout it out only on Linux and Mac.
  "npm_modules_revision": "f83fafaa22f5ff396cf5306285ca3806d1b2cf1b",
}

deps = {
  "build":
    "https://chromium.googlesource.com/chromium/tools/build.git",

  # Used to initiate bootstrapping.
  #
  # This commit resolves to tag "15.1.0".
  "infra/bootstrap/virtualenv-ext":
   ("https://chromium.googlesource.com/external/github.com/pypa/virtualenv"
     "@2db288fe1f8b6755f8df53b8a8a0667a0f71b1cf"),

  "infra/luci":
   ("https://chromium.googlesource.com/infra/luci/luci-py"
     "@c7c92ca1eef54f04e8b0461de154f0a530cc1074"),

  # This unpinned dependency is present because it is used by the trybots for
  # the recipes-py repo; They check out infra with this at HEAD, and then apply
  # the patch to it and run verifications within that copy of the repo. They
  # piggyback on top of infra in order to take advantage of it's precompiled
  # version of python-coverage.
  "infra/recipes-py":
   ("https://chromium.googlesource.com/infra/luci/recipes-py"
     "@origin/master"),

  "infra/go/src/go.chromium.org/luci":
    ("https://chromium.googlesource.com/infra/luci/luci-go"
     "@f160ecfad9831737c3c270a0c44f7be68f6611c1"),

  "infra/go/src/go.chromium.org/gae":
    ("https://chromium.googlesource.com/infra/luci/gae"
     "@c93eea5a34cfd3efc3b65e72fa3383b59de7c6a3"),

  # Appengine third_party DEPS
  "infra/appengine/third_party/bootstrap":
    ("https://chromium.googlesource.com/external/github.com/twbs/bootstrap.git"
     "@b4895a0d6dc493f17fe9092db4debe44182d42ac"),

  "infra/appengine/third_party/cloudstorage":
    ("https://chromium.googlesource.com/external/github.com/"
     "GoogleCloudPlatform/appengine-gcs-client.git"
     "@76162a98044f2a481e2ef34d32b7e8196e534b78"),

  "infra/appengine/third_party/six":
    ("https://chromium.googlesource.com/external/bitbucket.org/gutworth/six.git"
     "@06daa9c7a4b43c49bf2a3d588c52672e1c5e3a2c"),

  "infra/appengine/third_party/oauth2client":
    ("https://chromium.googlesource.com/external/github.com/google/oauth2client.git"
     "@e8b1e794d28f2117dd3e2b8feeb506b4c199c533"),

  "infra/appengine/third_party/uritemplate":
    ("https://chromium.googlesource.com/external/github.com/uri-templates/"
     "uritemplate-py.git"
     "@1e780a49412cdbb273e9421974cb91845c124f3f"),

  "infra/appengine/third_party/httplib2":
    ("https://chromium.googlesource.com/external/github.com/jcgregorio/httplib2.git"
     "@058a1f9448d5c27c23772796f83a596caf9188e6"),

  "infra/appengine/third_party/endpoints-proto-datastore":
    ("https://chromium.googlesource.com/external/github.com/"
     "GoogleCloudPlatform/endpoints-proto-datastore.git"
     "@971bca8e31a4ab0ec78b823add5a47394d78965a"),

  "infra/appengine/third_party/difflibjs":
    ("https://chromium.googlesource.com/external/github.com/qiao/difflib.js.git"
     "@e11553ba3e303e2db206d04c95f8e51c5692ca28"),

  "infra/appengine/third_party/pipeline":
    ("https://chromium.googlesource.com/external/github.com/"
     "GoogleCloudPlatform/appengine-pipelines.git"
     "@58cf59907f67db359fe626ee06b6d3ac448c9e15"),

  "infra/appengine/third_party/google-api-python-client":
    ("https://chromium.googlesource.com/external/github.com/google/"
     "google-api-python-client.git"
     "@49d45a6c3318b75e551c3022020f46c78655f365"),

  "infra/appengine/third_party/catapult":
    ("https://chromium.googlesource.com/external/github.com/catapult-project/"
     "catapult.git"
     "@506457cbd726324f327b80ae11f46c1dfeb8710d"),

  "infra/appengine/third_party/gae-pytz":
    ("https://chromium.googlesource.com/external/code.google.com/p/gae-pytz/"
     "@4d72fd095c91f874aaafb892859acbe3f927b3cd"),

  "infra/appengine/third_party/dateutil":
    ("https://chromium.googlesource.com/external/code.launchpad.net/dateutil/"
     "@8c6026ba09716a4e164f5420120bfe2ebb2d9d82"),

  ## For ease of development. These are pulled in as wheels for run.py/test.py
  "expect_tests":
    "https://chromium.googlesource.com/infra/testing/expect_tests.git",
  "testing_support":
    "https://chromium.googlesource.com/infra/testing/testing_support.git",

  "infra/appengine/third_party/src/github.com/golang/oauth2":
  ("https://chromium.googlesource.com/external/github.com/golang/oauth2.git"
   "@cb029f4c1f58850787981eefaf9d9bf547c1a722"),
}


deps_os = {
  "unix": {
    "infra/appengine/third_party/npm_modules":
      ("https://chromium.googlesource.com/infra/third_party/npm_modules.git@" +
      Var("npm_modules_revision")),
  },
  "mac": {
    "infra/appengine/third_party/npm_modules":
      ("https://chromium.googlesource.com/infra/third_party/npm_modules.git@" +
      Var("npm_modules_revision")),
  }
}

hooks = [
  {
    "pattern": ".",
    "action": [
      "python", "-u", "./infra/bootstrap/remove_orphaned_pycs.py",
    ],
  },
  {
    "pattern": ".",
    "action": [
      "python", "-u", "./infra/bootstrap/bootstrap.py",
      "--deps_file", "infra/bootstrap/deps.pyl", "infra/ENV"
    ],
  },
  {
    "pattern": ".",
    "action": [
      "python", "-u", "./infra/bootstrap/install_cipd_packages.py", "-v",
    ],
  },
  {
    "pattern": ".",
    "action": [
      "python", "-u", "./infra/bootstrap/get_appengine.py", "--dest=.",
    ],
  },
]

recursedeps = ['build']
