# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import logging

from google.appengine.api.modules import modules

from components import auth
from components import config as config_api
from components import decorators
from components import endpoints_webapp2
from components import prpc

import webapp2

from legacy import api as legacy_api
from legacy import api_common
from legacy import swarmbucket_api
import bq
import bulkproc
import config
import model
import notifications
import resultdb
import service
import swarming
import user

README_MD = (
    'https://chromium.googlesource.com/infra/infra/+/HEAD/'
    'appengine/cr-buildbucket/README.md'
)


class DummyStartHandler(auth.AuthenticatingHandler):  # pragma: no cover
  """Dummy handler for /_ah/start.

  Derived from AuthenticatingHandler to initialize auth db before the first
  request.
  """

  @auth.public
  def get(self):
    pass


class MainHandler(webapp2.RequestHandler):  # pragma: no cover
  """Redirects to README.md."""

  def get(self):
    # `led` fetches the index page to see if Buildbucket is available. It
    # follows HTTP-level redirects. We don't want it to test Gitiles available.
    # Use Javascript-level redirect instead.
    self.response.headers['Content-Type'] = 'text/html; charset=utf-8'
    self.response.write(
        """
      <p>Redirecting to the Buildbucket documentation...</p>
      <script>window.location.replace(%r);</script>
    """ % README_MD
    )


class CronUpdateBuckets(webapp2.RequestHandler):  # pragma: no cover
  """Updates buckets from configs."""

  @decorators.require_cronjob
  def get(self):
    config.cron_update_buckets()


class BuildRPCHandler(webapp2.RequestHandler):  # pragma: no cover
  """Redirects to API explorer to see the build."""

  def get(self, build_id):
    api_path = '/_ah/api/buildbucket/v1/builds/%s' % build_id
    return self.redirect(api_path)


class RedirectHandlerBase(auth.AuthenticatingHandler):  # pragma: no cover

  def _get_build(self, build_id):
    try:
      build_id = int(build_id)
    except ValueError:
      self.response.write('invalid build id')
      self.abort(400)

    build = model.Build.get_by_id(build_id)
    can_view = build and user.has_perm(user.PERM_BUILDS_GET, build.bucket_id)

    if not can_view:
      if auth.get_current_identity().is_anonymous:
        self.redirect(self.create_login_url(self.request.url), abort=True)
      self.response.write('build %d not found' % build_id)
      self.abort(404)

    return build


class ViewBuildHandler(RedirectHandlerBase):  # pragma: no cover
  """Redirects to Milo build page."""

  @auth.public
  def get(self, build_id):
    build = self._get_build(build_id)
    return self.redirect(str(api_common.get_build_url(build)))


class ViewLogHandler(RedirectHandlerBase):  # pragma: no cover
  """Redirects to LogDog log content page."""

  @staticmethod
  def _find_log(build_proto, step_name, log_name):
    for step in build_proto.steps:
      if step.name == step_name:
        for log in step.logs:
          if log.name == log_name:
            return log
        break
    return None

  @auth.public
  def get(self, build_id, step_name):
    log_name = self.request.params.get('log') or 'stdout'
    build = self._get_build(build_id)
    bundle = model.BuildBundle.get(build, steps=True)
    bundle.to_proto(build.proto, load_tags=False)
    log = self._find_log(build.proto, step_name, log_name)
    if not log or not log.view_url:
      self.abort(
          404, 'view url for log %r in step %r in build %r not found' %
          (log_name, step_name, build_id)
      )
    return self.redirect(str(log.view_url))


class UnregisterBuilders(webapp2.RequestHandler):  # pragma: no cover
  """Unregisters builders that didn't have builds for a long time."""

  @decorators.require_cronjob
  def get(self):
    service.unregister_builders()

def get_frontend_routes():  # pragma: no cover
  endpoints_services = [
      legacy_api.BuildBucketApi,
      config_api.ConfigApi,
      swarmbucket_api.SwarmbucketApi,
  ]
  routes = [
      webapp2.Route(r'/_ah/start', DummyStartHandler),
      webapp2.Route(r'/', MainHandler),
      webapp2.Route(r'/b/<build_id:\d+>', BuildRPCHandler),
      webapp2.Route(r'/build/<build_id:\d+>', ViewBuildHandler),
      webapp2.Route(r'/builds/<build_id:\d+>', ViewBuildHandler),
      webapp2.Route(r'/log/<build_id:\d+>/<step_name:.+>', ViewLogHandler),
  ]
  routes.extend(endpoints_webapp2.api_routes(endpoints_services))
  # /api routes should be removed once clients are hitting /_ah/api.
  routes.extend(
      endpoints_webapp2.api_routes(endpoints_services, base_path='/api')
  )

  prpc_server = prpc.Server()
  prpc_server.add_interceptor(auth.prpc_interceptor)
  routes += prpc_server.get_routes()

  return routes


def get_backend_routes():  # pragma: no cover
  prpc_server = prpc.Server()
  prpc_server.add_interceptor(auth.prpc_interceptor)

  return [  # pragma: no branch
      webapp2.Route(r'/internal/cron/buildbucket/update_buckets',
                    CronUpdateBuckets),
      webapp2.Route(r'/internal/cron/buildbucket/bq-export',
                    bq.CronExportBuilds),
      webapp2.Route(r'/internal/cron/buildbucket/unregister-builders',
                    UnregisterBuilders),
      webapp2.Route(r'/internal/task/buildbucket/notify/<build_id:\d+>',
                    notifications.TaskPublishNotification),
      webapp2.Route(r'/internal/task/bq/export/<build_id:\d+>',
                    bq.TaskExport),
      webapp2.Route(r'/internal/task/resultdb/finalize/<build_id:\d+>',
                    resultdb.FinalizeInvocation),
  ] + (bulkproc.get_routes() + prpc_server.get_routes())
