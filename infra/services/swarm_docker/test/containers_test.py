# Copyright (c) 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from datetime import datetime
import collections
import docker
import mock
import requests
import sys
import unittest

from infra.services.swarm_docker import containers


class FakeClient(object):
  """Mocks the client object returned from docker's client API.

  containers.DockerClient wraps it. Mocked here to verify wrapper class
  bheaves correctly.
  """
  def __init__(self):
    self.containers = None
    self.images = FakeImageList()
    self.creds = None
    self.responsive = True

  def ping(self):
    if self.responsive:
      return True
    else:
      raise docker.errors.APIError('omg engine not running')


class FakeImage(object):

  def __init__(self, image_id, image_url):
    self.id = image_id
    self.tags = [image_url]


class FakeImageList(object):
  def __init__(self):
    self.images = []

  def list(self):
    return self.images

  def get(self, image_url):
    for i in self.images:
      if image_url in i.tags:
        return True
    raise docker.errors.ImageNotFound('omg no image')

  def remove(self, image_id):
    for i in self.images[:]:
      if i.id == image_id:
        self.images.remove(i)

  def pull(self, image):
    self.images.append(image)


class FakeContainer(object):
  """Used to mock containers.Container"""
  def __init__(self, name, uptime=None):
    self._container = FakeContainerBackend(name)
    self.name = name
    self.uptime = uptime
    self.swarming_bot_killed = False

  def get_container_uptime(self, _):
    return self.uptime

  def kill_swarming_bot(self, *_, **__):
    self.swarming_bot_killed = True


class FakeContainerBackend(object):
  """Mocks the container objects returned from docker's client API.

  containers.Container wraps each one. Mocked here to verify the wrapper class
  behaves correctly.
  """

  ExecResult = collections.namedtuple('ExecResult', 'exit_code,output')

  def __init__(self, name, devices='not set'):
    self.name = name
    self.was_deleted = False
    self.was_started = False
    self.was_stopped = False
    self.is_paused = False
    self.exec_outputs = []
    self.exec_inputs = []
    self.attrs = {}
    self.devices = devices
    self.image = ''

  def remove(self, **_kwargs):
    self.was_deleted = True

  def start(self):
    self.was_started = True

  def pause(self):
    assert not self.is_paused
    self.is_paused = True

  def unpause(self):
    assert self.is_paused
    self.is_paused = False

  def stop(self, **_kwargs):
    self.was_stopped = True

  def exec_run(self, cmd, **_kwargs):
    self.exec_inputs.append(cmd)
    return self.ExecResult(0, self.exec_outputs.pop(0))


class FakeContainerList(object):
  """Mocks the container list objects returned from docker's client API."""
  def __init__(self, containers_list):
    self._list = containers_list

  def create(self, **kwargs):
    return FakeContainerBackend(kwargs['name'], kwargs['devices'])

  def list(self, filters=None, **_kwargs):  # pylint: disable=unused-argument
    if filters is None:
      filters = {}
    status = filters.get('status')
    if status == 'paused':
      return [c for c in self._list if c.is_paused]
    elif status == 'running':
      return [c for c in self._list if not c.is_paused]
    elif status == 'created':
      return [
          c for c in self._list if c.attrs.get('State', {}).get(
              'Status') == 'created']
    else:
      return self._list

  def get(self, name):
    for c in self._list:
      if c.name == name:
        return c
    raise docker.errors.NotFound('omg container missing')


class TestContainerDescriptor(unittest.TestCase):
  def setUp(self):
    self.desc = containers.ContainerDescriptor('7')

  def test_name(self):
    self.assertEquals(self.desc.name, '7')

  @mock.patch('socket.gethostname')
  def test_hostname(self, mock_gethostname):
    mock_gethostname.return_value = 'build123-a4'
    self.assertEquals(self.desc.hostname, 'build123-a4--7')

  def test_log_started_smoke(self):
    self.desc.log_started()

  def test_shutdown_file(self):
    self.assertEqual(self.desc.shutdown_file, '/b/7.shutdown.stamp')

  def test_lock_file(self):
    self.assertEqual(self.desc.lock_file, '/var/lock/swarm_docker.7.lock')

  def test_should_create_container(self):
    self.assertTrue(self.desc.should_create_container())


class TestDockerClient(unittest.TestCase):
  def setUp(self):
    self.fake_client = FakeClient()
    mock.patch('docker.from_env', return_value=self.fake_client).start()
    self.container_names = ['5', '6']
    self.fake_client.containers = FakeContainerList(
        [FakeContainerBackend(name) for name in self.container_names])

  @mock.patch('time.sleep')
  def test_ping_success(self, mock_sleep):
    self.fake_client.responsive = True
    mock_sleep.return_value = None

    client = containers.DockerClient()
    self.assertTrue(client.ping())

  @mock.patch('time.sleep')
  def test_ping_fail(self, mock_sleep):
    self.fake_client.responsive = False
    mock_sleep.return_value = None

    client = containers.DockerClient()
    self.assertFalse(client.ping(retries=5))
    mock_sleep.assert_has_calls(
        [mock.call(1), mock.call(2), mock.call(4), mock.call(8)])

  def test_images(self):
    img1 = FakeImage('image1-id', 'image1-url')
    self.fake_client.images.images.append(img1)

    client = containers.DockerClient()
    self.assertEqual(client.images(), [img1])

  def test_has_image(self):
    self.fake_client.images.images.append(FakeImage('image1-id', 'image1-url'))

    client = containers.DockerClient()
    self.assertTrue(client.has_image('image1-url'))
    self.assertFalse(client.has_image('image99-url'))

  def test_remove_image(self):
    self.fake_client.images.images.append(FakeImage('image1-id', 'image1-url'))

    client = containers.DockerClient()
    client.remove_image('image1-id')
    self.assertEqual(client.images(), [])

  def test_remove_outdated_images(self):
    old_img = FakeImage('old-image-id', 'old-image-url')
    new_img = FakeImage('new-image-id', 'new-image-url')
    self.fake_client.images.images = [old_img, new_img]

    client = containers.DockerClient()
    client.remove_outdated_images('new-image-url')
    self.assertEqual(client.images(), [new_img])

  def test_remove_outdated_images_no_op(self):
    """remove_outdated_images() is a no-op. Needed for 100% coverage."""
    new_img = FakeImage('new-image-id', 'new-image-url')
    self.fake_client.images.images = [new_img]

    client = containers.DockerClient()
    client.remove_outdated_images('new-image-url')
    self.assertEqual(client.images(), [new_img])

  def test_pull(self):
    client = containers.DockerClient()
    client.logged_in = True
    client.pull('image1')
    self.assertTrue('image1' in self.fake_client.images.images)

  def test_get_running_containers(self):
    running_containers = containers.DockerClient().get_running_containers()
    self.assertEqual(
        set(c.name for c in running_containers), set(self.container_names))

  def test_get_paused_containers(self):
    self.fake_client.containers.get('5').pause()

    paused_containers = containers.DockerClient().get_paused_containers()
    self.assertEqual(len(paused_containers), 1)
    self.assertEqual(paused_containers[0].name, '5')

  def test_get_created_containers(self):
    self.fake_client.containers.get('5').attrs['State'] = {'Status': 'created'}

    created_containers = containers.DockerClient().get_created_containers()
    self.assertEqual(len(created_containers), 1)
    self.assertEqual(created_containers[0].name, '5')

  def test_get_container(self):
    container = containers.DockerClient().get_container(
        containers.ContainerDescriptor('5'))
    self.assertEqual(container.name, '5')

  def test_get_missing_container(self):
    container = containers.DockerClient().get_container(
        containers.ContainerDescriptor('1'))
    self.assertEqual(container, None)

  def test_stop_old_containers(self):
    young_container = FakeContainer('young_container', uptime=10)
    old_container = FakeContainer('old_container', uptime=999)

    containers.DockerClient().stop_old_containers(
        [young_container, old_container], 100)
    self.assertFalse(young_container.swarming_bot_killed)
    self.assertTrue(old_container.swarming_bot_killed)

  def test_stop_frozen_containers(self):

    def _raise_frozen_container(*_args, **_kwargs):
      raise containers.FrozenContainerError()
    frozen_container1 = FakeContainer('frozen_container1', uptime=999)
    frozen_container1.kill_swarming_bot = _raise_frozen_container
    frozen_container2 = FakeContainer('frozen_container2', uptime=999)
    frozen_container2.kill_swarming_bot = _raise_frozen_container

    with self.assertRaises(containers.FrozenEngineError):
      containers.DockerClient().stop_old_containers(
          [frozen_container1, frozen_container2], 100)

  def test_delete_stopped_containers(self):
    created_c = FakeContainerBackend('11')
    created_c.attrs['State'] = {'Status': 'created'}
    self.fake_client.containers._list.append(created_c)

    containers.DockerClient().delete_stopped_containers()
    self.assertTrue(
        all(c.was_deleted for c in self.fake_client.containers.list()))

  @mock.patch('os.chown')
  @mock.patch('os.mkdir')
  @mock.patch('os.path.exists')
  @mock.patch('pwd.getpwnam')
  def test_create_container(self, mock_getpwnam, mock_exists, mock_mkdir,
                            mock_chown):
    mock_getpwnam.return_value = collections.namedtuple(
        'pwnam', 'pw_uid, pw_gid')(1,2)
    mock_exists.return_value = False
    running_containers = [FakeContainer('1'), FakeContainer('2')]
    self.fake_client.containers = FakeContainerList(running_containers)

    container = containers.DockerClient().create_container(
        containers.ContainerDescriptor('1'), 'image', 'swarm-url.com', {})
    self.assertEquals(container.name, '1')
    mock_chown.assert_called_with(mock_mkdir.call_args[0][0], 1, 2)

  @mock.patch('os.chown')
  @mock.patch('os.mkdir')
  @mock.patch('os.path.exists')
  @mock.patch('pwd.getpwnam')
  def test_create_container_with_env(self, mock_getpwnam, mock_exists,
                                     mock_mkdir, mock_chown):
    mock_getpwnam.return_value = collections.namedtuple(
        'pwnam', 'pw_uid, pw_gid')(1,2)
    mock_exists.return_value = False
    running_containers = [FakeContainer('1'), FakeContainer('2')]
    self.fake_client.containers = FakeContainerList(running_containers)
    additional_env = {'SOME_ENV': 'SOME_VAL'}

    container = containers.DockerClient().create_container(
        containers.ContainerDescriptor('1'), 'image', 'swarm-url.com', {},
        additional_env)
    self.assertEquals(container.name, '1')
    mock_chown.assert_called_with(mock_mkdir.call_args[0][0], 1, 2)

  @mock.patch('os.chown')
  @mock.patch('os.mkdir')
  @mock.patch('os.path.exists')
  @mock.patch('pwd.getpwnam')
  @mock.patch('sys.platform', 'darwin')
  def test_create_container_darwin(self, mock_getpwnam, mock_exists, mock_mkdir,
                                   mock_chown):
    mock_getpwnam.return_value = collections.namedtuple(
        'pwnam', 'pw_uid, pw_gid')(1,2)
    mock_exists.side_effect = lambda d: d == containers._KVM_DEVICE

    container = containers.DockerClient().create_container(
        containers.ContainerDescriptor('1'), 'image', 'swarm-url.com', {})
    self.assertEquals(container.name, '1')
    mock_chown.assert_called_with(mock_mkdir.call_args[0][0], 1, 2)
    self.assertEquals(container.devices, None)

  @mock.patch('os.chown')
  @mock.patch('os.mkdir')
  @mock.patch('os.path.exists')
  @mock.patch('pwd.getpwnam')
  @mock.patch('sys.platform', 'linux2')
  def test_create_container_linux_no_kvm(self, mock_getpwnam, mock_exists,
                                         mock_mkdir, mock_chown):
    mock_getpwnam.return_value = collections.namedtuple(
        'pwnam', 'pw_uid, pw_gid')(1,2)
    mock_exists.return_value = False

    container = containers.DockerClient().create_container(
        containers.ContainerDescriptor('1'), 'image', 'swarm-url.com', {})
    self.assertEquals(container.name, '1')
    mock_chown.assert_called_with(mock_mkdir.call_args[0][0], 1, 2)
    self.assertEquals(container.devices, None)

  @mock.patch('os.chown')
  @mock.patch('os.mkdir')
  @mock.patch('os.path.exists')
  @mock.patch('pwd.getpwnam')
  @mock.patch('sys.platform', 'linux2')
  def test_create_container_linux_kvm(self, mock_getpwnam, mock_exists,
                                      mock_mkdir, mock_chown):
    mock_getpwnam.return_value = collections.namedtuple(
        'pwnam', 'pw_uid, pw_gid')(1,2)
    mock_exists.side_effect = lambda d: d == containers._KVM_DEVICE

    container = containers.DockerClient().create_container(
        containers.ContainerDescriptor('1'), 'image', 'swarm-url.com', {})
    self.assertEquals(container.name, '1')
    mock_chown.assert_called_with(mock_mkdir.call_args[0][0], 1, 2)
    self.assertEquals(container.devices,
                      ['{0}:{0}'.format(containers._KVM_DEVICE)])

  @mock.patch('os.chown')
  @mock.patch('os.mkdir')
  @mock.patch('os.path.exists')
  @mock.patch('pwd.getpwnam')
  @mock.patch('sys.platform', 'linux2')
  def test_create_container_linux_tun(self, mock_getpwnam, mock_exists,
                                      mock_mkdir, mock_chown):
    mock_getpwnam.return_value = collections.namedtuple(
        'pwnam', 'pw_uid, pw_gid')(1,2)
    mock_exists.side_effect = lambda d: d in (
        containers._KVM_DEVICE, containers._TUN_DEVICE)

    container = containers.DockerClient().create_container(
        containers.ContainerDescriptor('1'), 'image', 'swarm-url.com', {})
    self.assertEquals(container.name, '1')
    mock_chown.assert_called_with(mock_mkdir.call_args[0][0], 1, 2)
    self.assertEquals(container.devices,
                      ['{0}:{0}'.format(containers._KVM_DEVICE),
                       '{0}:{0}'.format(containers._TUN_DEVICE)])

  def test_num_containers_is_set(self):
    client = containers.DockerClient()
    self.assertIsNone(client._get_env('').get('NUM_CONFIGURED_CONTAINERS'))
    client.set_num_configured_containers(42)
    self.assertEquals(client._get_env('').get('NUM_CONFIGURED_CONTAINERS'), 42)

  @mock.patch('socket.getfqdn')
  def test_host_hostname_is_set(self, mock_getfqdn):
    mock_getfqdn.return_value = 'hostofa_hostofa_hostofa_host'
    client = containers.DockerClient()
    self.assertEquals(
        client._get_env('').get('DOCKER_HOST_HOSTNAME'),
        'hostofa_hostofa_hostofa_host')


class TestContainer(unittest.TestCase):

  def setUp(self):
    self.container_backend = FakeContainerBackend('container1')
    self.container = containers.Container(self.container_backend)

  def test_get_labels(self):
    self.container_backend.attrs = {'Config': {'Labels': {'label1': 'val1'}}}
    self.assertEquals(self.container.labels, {'label1': 'val1'})

  def test_get_exit_code(self):
    self.container_backend.attrs['State'] = {'ExitCode': 111}
    self.assertEqual(self.container.exit_code, 111)

  def test_get_state(self):
    self.container_backend.attrs = {'State': {'Status': 'running'}}
    status = self.container.state
    self.assertEquals(status, 'running')

  def test_get_image(self):
    self.container_backend.image = 'test-image'
    self.assertEquals(self.container.image, 'test-image')

  def test_get_container_uptime(self):
    now = datetime.strptime(
        '2000-01-01T01:30:00.000000', '%Y-%m-%dT%H:%M:%S.%f')
    self.container_backend.attrs = {
        'State': {'StartedAt': '2000-01-01T00:00:00.0000000000'}
    }
    uptime = self.container.get_container_uptime(now)
    self.assertEquals(uptime, 90)

  def test_get_swarming_bot_pid(self):
    self.container_backend.exec_outputs = ['123']
    pid = self.container.get_swarming_bot_pid()
    self.assertEquals(pid, 123)

  def test_get_swarming_bot_pid_backend_error(self):
    self.container_backend.exec_outputs = ['rpc error: omg failure']
    pid = self.container.get_swarming_bot_pid()
    self.assertEquals(pid, None)

  def test_get_swarming_bot_pid_lsof_error(self):
    self.container_backend.exec_outputs = ['omg lsof failure']
    pid = self.container.get_swarming_bot_pid()
    self.assertEquals(pid, None)

  def test_get_swarming_bot_pid_404_error(self):
    def _raises_docker_not_found(*_args, **_kwargs):
      raise docker.errors.NotFound('404')
    self.container_backend.exec_run = _raises_docker_not_found
    pid = self.container.get_swarming_bot_pid()
    self.assertEquals(pid, None)

  def test_kill_swarming_bot(self):
    self.container_backend.exec_outputs = ['123', '']
    self.container.kill_swarming_bot()
    self.assertEquals(self.container_backend.exec_inputs[-1], 'kill -15 123')

  def test_kill_swarming_bot_error_no_shutdown(self):
    self.container_backend.attrs = {
        'State': {
            'StartedAt': '2000-01-01T00:00:00.0000000000'
        }
    }
    # 1 hour uptime.
    now = datetime.strptime('2000-01-01T01:00:00.000000',
                            '%Y-%m-%dT%H:%M:%S.%f')
    self.container_backend.exec_outputs = ['omg failure']
    self.container.kill_swarming_bot(now=now, max_uptime=60)
    # Ensure nothing was killed when the bot's pid couldn't be found and its
    # uptime is much less than max_uptime.
    self.assertFalse(
        any('kill -15' in cmd for cmd in self.container_backend.exec_inputs))
    self.assertFalse(self.container_backend.was_stopped)

  def test_kill_swarming_bot_error_shutdown(self):
    self.container_backend.attrs = {
        'State': {
            'StartedAt': '2000-01-01T00:00:00.0000000000'
        }
    }
    # 12 hour uptime.
    now = datetime.strptime('2000-01-01T12:00:00.000000',
                            '%Y-%m-%dT%H:%M:%S.%f')
    self.container_backend.exec_outputs = ['omg failure']
    self.container.kill_swarming_bot(now=now, max_uptime=60)
    # Ensure the container was shutdown when the bot's pid couldn't be found
    # and its uptime was much larger than max_uptime.
    self.assertFalse(
        any('kill -15' in cmd for cmd in self.container_backend.exec_inputs))
    self.assertTrue(self.container_backend.was_stopped)

  def test_kill_swarming_bot_cant_kill(self):
    def _raise_requests_timeout(**_kwargs):
      raise requests.exceptions.ReadTimeout()

    self.container_backend.attrs = {
        'State': {
            'StartedAt': '2000-01-01T00:00:00.0000000000'
        }
    }
    # 1 hour uptime.
    now = datetime.strptime('2000-01-01T01:00:00.000000',
                            '%Y-%m-%dT%H:%M:%S.%f')
    self.container_backend.exec_outputs = ['omg failure']
    self.container_backend.stop = _raise_requests_timeout
    self.container.kill_swarming_bot(now=now)
    # Ensure nothing was killed when the bot's pid couldn't be found.
    self.assertFalse(
        any('kill -15' in cmd for cmd in self.container_backend.exec_inputs))
    self.assertFalse(self.container_backend.was_stopped)
    self.assertTrue(self.container_backend.was_deleted)

  def test_kill_swarming_bot_cant_remove(self):
    def _raise_requests_timeout(**_kwargs):
      raise requests.exceptions.ReadTimeout()
    def _raise_docker_api_error(**_kwargs):
      raise docker.errors.APIError('omg error')

    self.container_backend.attrs = {
        'State': {
            'StartedAt': '2000-01-01T00:00:00.0000000000'
        }
    }
    # 1 hour uptime.
    now = datetime.strptime('2000-01-01T01:00:00.000000',
                            '%Y-%m-%dT%H:%M:%S.%f')
    self.container_backend.exec_outputs = ['omg failure']
    self.container_backend.stop = _raise_requests_timeout
    self.container_backend.remove = _raise_docker_api_error
    with self.assertRaises(containers.FrozenContainerError):
      self.container.kill_swarming_bot(now=now)
    # Ensure nothing was killed when the bot's pid couldn't be found.
    self.assertFalse(
        any('kill -15' in cmd for cmd in self.container_backend.exec_inputs))
    self.assertFalse(self.container_backend.was_stopped)
    self.assertFalse(self.container_backend.was_deleted)

  def test_pause_unpause(self):
    self.container.pause()
    self.assertTrue(self.container_backend.is_paused)
    self.container.unpause()
    self.assertFalse(self.container_backend.is_paused)

  def test_exec_run(self):
    self.container_backend.exec_outputs = ['', '']
    self.container.exec_run('ls')
    self.container.exec_run('cd')
    self.assertEquals(self.container_backend.exec_inputs, ['ls', 'cd'])

  def test_attrs(self):
    self.container_backend.attrs = {'Id': '123'}
    self.assertEquals(self.container.attrs['Id'], '123')


if __name__ == '__main__':
  unittest.main()
