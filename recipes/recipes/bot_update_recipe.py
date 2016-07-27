DEPS = [
  'depot_tools/bot_update',
  'depot_tools/gclient',
  'recipe_engine/properties',
]

def RunSteps(api):
  api.gclient.set_config('infra')
  api.bot_update.ensure_checkout()


def GenTests(api):
  yield api.test('basic') + api.properties.tryserver_gerrit(
      full_project_name='infra/infra',
      gerrit_host='chromium-review.googlesource.com',
      mastername='tryserver.infra',
      buildername='infra_tester')
