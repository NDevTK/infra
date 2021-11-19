# Copyright 2020 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from collections import namedtuple
from contextlib import contextmanager
from recipe_engine import recipe_api

from PB.go.chromium.org.luci.buildbucket.proto.common import FAILURE, SUCCESS
from PB.recipe_engine.result import RawResult

from PB.recipes.infra import gae_tarball_uploader as pb

DEPS = [
    'recipe_engine/buildbucket',
    'recipe_engine/futures',
    'recipe_engine/json',
    'recipe_engine/path',
    'recipe_engine/properties',
    'recipe_engine/step',
    'recipe_engine/time',

    'cloudbuildhelper',
    'infra_checkout',
]


PROPERTIES = pb.Inputs


# Metadata is returned by _checkout.
Metadata = namedtuple('Metadata', [
    'repo_url',       # "https://..."
    'revision',       # "abcdefacbdf..."
    'canonical_tag',  # derived from the git revision and commit position
    'checkout',       # cloudbuildhelper.CheckoutMetadata
])


def RunSteps(api, properties):
  try:
    _validate_props(properties)
  except ValueError as exc:
    raise recipe_api.InfraFailure('Bad input properties: %s' % exc)

  # Checkout the code.
  meta, build_env = _checkout(api, properties.project)

  # Discover what *.yaml manifests (full paths to them) we need to build.
  manifests = api.cloudbuildhelper.discover_manifests(
      meta.checkout.root, properties.manifests)
  if not manifests:  # pragma: no cover
    raise recipe_api.InfraFailure('Found no manifests to build')

  with build_env(api):
    # Report the exact version we going to use.
    api.cloudbuildhelper.report_version()

    # Build and upload corresponding tarballs (in parallel).
    futures = {}
    for m in manifests:
      fut = api.futures.spawn(
          api.cloudbuildhelper.upload,
          manifest=m,
          canonical_tag=meta.canonical_tag,
          build_id=api.buildbucket.build_url(),
          infra=properties.infra,
          checkout_metadata=meta.checkout)
      futures[fut] = m

  # Wait until all uploads complete.
  built = []
  fails = []
  for fut in api.futures.iwait(futures.keys()):
    try:
      built.append(fut.result())
    except api.step.StepFailure:
      fails.append(api.path.basename(futures[fut]))

  summary_lines = []
  # Try to roll even if something failed. One broken tarball should not block
  # the rest of them.
  if built and properties.HasField('roll_into'):
    with api.step.nest('upload roll CL') as roll:
      num, url = _roll_built_tarballs(api, properties.roll_into, built, meta)
      if num is not None:
        roll.presentation.links['Issue %s' % num] = url
        summary_lines.extend([
          'Created roll CL ' + url,
          ''
        ])

  status = SUCCESS
  if fails:
    status = FAILURE
    summary_lines.append('Failed to build:')
    summary_lines.extend('  * %s' % f for f in fails)
  return RawResult(status=status, summary_markdown='\n'.join(summary_lines))


def _validate_props(p):  # pragma: no cover
  if p.project == PROPERTIES.PROJECT_UNDEFINED:
    raise ValueError('"project" is required')
  if not p.infra:
    raise ValueError('"infra" is required')
  if not p.manifests:
    raise ValueError('"manifests" is required')


def _checkout(api, project):
  """Checks out some committed revision (based on Buildbucket properties).

  Args:
    api: recipes API.
    project: PROPERTIES.Project enum.

  Returns:
    (Metadata, build environment context manager).
  """
  conf, internal, repo_url = {
    PROPERTIES.PROJECT_INFRA: (
        'infra',
        False,
        'https://chromium.googlesource.com/infra/infra',
    ),
    PROPERTIES.PROJECT_INFRA_INTERNAL: (
        'infra_internal',
        True,
        'https://chrome-internal.googlesource.com/infra/infra_internal',
    ),
  }[project]

  co = api.infra_checkout.checkout(
      gclient_config_name=conf,
      internal=internal,
      go_version_variant='bleeding_edge')
  co.gclient_runhooks()

  props = co.bot_update_step.presentation.properties

  @contextmanager
  def build_environ(api):
    with co.go_env():
      # Use 'cloudbuildhelper' that comes with the infra checkout (it's in
      # PATH), to make sure builders use the same version as developers.
      api.cloudbuildhelper.command = 'cloudbuildhelper'
      yield

  return Metadata(
      repo_url=repo_url,
      revision=props['got_revision'],
      canonical_tag=api.cloudbuildhelper.get_commit_label(
          path=co.path.join('infra_internal' if internal else 'infra'),
          revision=props['got_revision'],
          commit_position=props.get('got_revision_cp'),
      ),
      checkout=api.cloudbuildhelper.CheckoutMetadata(
          root=co.path,
          repos=co.bot_update_step.json.output['manifest'],
      )), build_environ


def _roll_built_tarballs(api, spec, tarballs, meta):
  """Uploads a CL with info about tarballs into a repo with pinned tarballs.

  See comments in gae_tarball_uploader.proto for more details.

  Args:
    api: recipes API.
    spec: instance of pb.Inputs.RollInto proto with the config.
    tarballs: a list of CloudBuildHelperApi.Tarball with info about tarballs.
    meta: Metadata struct, as returned by _checkout.

  Returns:
    (None, None) if didn't create a CL (because nothing has changed).
    (Issue number, Issue URL) if created a CL.
  """
  return api.cloudbuildhelper.do_roll(
      repo_url=spec.repo_url,
      root=api.path['cache'].join('builder', 'roll'),
      callback=lambda root: _mutate_pins_repo(api, root, spec, tarballs, meta))


def _mutate_pins_repo(api, root, spec, tarballs, meta):
  """Modifies the checked out repo with tarball pins.

  Args:
    api: recipes API.
    root: the directory where the repo is checked out.
    spec: instance of images_builder.Inputs.RollInto proto with the config.
    tarballs: a list of CloudBuildHelperApi.Tarball with info about tarballs.
    meta: Metadata struct, as returned by _checkout.

  Returns:
    cloudbuildhelper.RollCL to proceed with the roll or None to skip it.
  """
  # RFC3389 timstamp in UTC zone.
  date = api.time.utcnow().isoformat('T') + 'Z'

  # Prepare version JSON specs for all tarballs.
  # See //scripts/roll_tarballs.py in infradata/gae repo.
  versions = []
  for tb in tarballs:
    versions.append({
        'tarball': tb.name,
        'version': {
            'version': tb.version,
            'location': 'gs://%s/%s' % (tb.bucket, tb.path),
            'sha256': tb.sha256,
            'metadata': {
                'date': date,
                'source': {
                    'repo': meta.repo_url,
                    'revision': meta.revision,
                },
                'sources': tb.sources,
                'links': {
                    'buildbucket': api.buildbucket.build_url(),
                },
            },
        },
    })

  # Add all new tags (if any).
  res = api.step(
      name='roll_tarballs.py',
      cmd=[root.join('scripts', 'roll_tarballs.py')],
      stdin=api.json.input({'tarballs': versions}),
      stdout=api.json.output(),
      step_test_data=lambda: api.json.test_api.output_stream(
          _roll_tarballs_test_data(versions)))
  rolled = res.stdout.get('tarballs') or []
  deployments = res.stdout.get('deployments') or []
  diff = res.stdout.get('diff') or ''

  # If added new pins, delete old unused pins (if any). Note that if we are
  # building a rollback CL, we often do not add new pins (since we actually
  # rebuild a previously built tarball). We still need to land a CL to do the
  # rollback. If it turns out nothing has changed, api.cloudbuildhelper.do_roll
  # will just skip uploading the change.
  if rolled:
    api.step(
        name='prune_tarballs.py',
        cmd=[root.join('scripts', 'prune_tarballs.py'), '--verbose'])

  # Generate the commit message.
  message = str('\n'.join([
      'Rolling in tarballs.',
      '',
      'Produced by %s' % api.buildbucket.build_url(),
      '',
      'Updated staging deployments:',
  ] + [
      '  * %s: %s -> %s' % (d['tarball'], d['from'], d['to'])
      for d in deployments
  ] + [''] + ([diff, ''] if diff else [])))

  # List of people to CC based on what staging deployments were updated.
  extra_cc = set()
  for dep in deployments:
    extra_cc.update(dep.get('cc') or [])

  return api.cloudbuildhelper.RollCL(
      message=message,
      cc=extra_cc,
      tbr=spec.tbr,
      commit=spec.commit)


def _roll_tarballs_test_data(versions):
  return {
      'tarballs': versions,
      'deployments': [
          {
              'cc': ['n1@example.com', 'n2@example.com'],
              'channel': 'staging',
              'from': 'prev-version',
              'spec': 'apps/something/channels.json',
              'tarball': v['tarball'],
              'to': v['version']['version'],
          }
          for v in versions
      ],
      'diff': 'Diff line1\nDiff line2',
  }


def GenTests(api):
  yield (
      api.test('ci-infra') +
      api.properties(
          project=PROPERTIES.PROJECT_INFRA,
          infra='prod',
          manifests=['infra/build/gae'],
      )
  )

  yield (
      api.test('ci-infra-internal') +
      api.properties(
          project=PROPERTIES.PROJECT_INFRA_INTERNAL,
          infra='prod',
          manifests=['infra_internal/build/gae'],
      )
  )

  yield (
      api.test('ci-infra-with-roll') +
      api.properties(
          project=PROPERTIES.PROJECT_INFRA,
          infra='prod',
          manifests=['infra/build/gae'],
          roll_into={
              'repo_url': 'https://tarballs.repo.example.com',
              'tbr': ['someone@example.com'],
              'commit': True,
          },
      ) +
      api.step_data('upload roll CL.git diff', retcode=1)
  )

  yield (
      api.test('build-failure') +
      api.properties(
          project=PROPERTIES.PROJECT_INFRA,
          infra='prod',
          manifests=['infra/build/images/deterministic'],
      ) +
      api.step_data(
          'cloudbuildhelper upload target',
          api.cloudbuildhelper.upload_error_output('Boom'),
          retcode=1)
  )

  yield (
      api.test('bad-props') +
      api.properties(project=0)
  )
