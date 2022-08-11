# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from recipe_engine.post_process import DropExpectation, StatusSuccess
from recipe_engine.post_process import StatusFailure, StepCommandRE
from recipe_engine.post_process import StatusException
from recipe_engine.recipe_api import Property
from RECIPE_MODULES.infra.windows_scripts_executor import test_helper as t
from PB.recipes.infra.windows_image_builder import vm, sources

DEPS = [
    'qemu', 'recipe_engine/raw_io', 'recipe_engine/path',
    'recipe_engine/properties', 'recipe_engine/platform'
]

PYTHON_VERSION_COMPATIBILITY = 'PY3'

PROPERTIES = vm.QEMU_VM


def RunSteps(api, qemu_vm):
  api.qemu.init('latest')
  api.qemu.start_vm(arch='amd64', qemu_vm=qemu_vm, kvm=False)
  api.qemu.start_vm(arch='amd64', qemu_vm=qemu_vm, kvm=True)


def GenTests(api):

  WIN_ISO = t.VM_DRIVE(
      name='win10.iso',
      ip=sources.Src(
          cipd_src=sources.CIPDSrc(
              package='infra/win10',
              refs='latest',
              platform='windows-amd64',
          )),
      op=None,
      interface='none',
      media='cdrom',
      readonly=True,
      fmt='raw',  # Not required. For test purpose only
  )

  VM = t.VM_CONFIG(
      name='test_vm',
      machine='pc-q35-2.10',
      cpu='Cascadelake-Server',
      smp='cpus=4',
      memory=8192,
      drives=[WIN_ISO],
      device=['ide-cd,drive=win10.iso'],
      extra_args=['-display', 'none']  # run headless
  )

  yield (
      api.test('Test start vm same guest and host arch') +
      api.post_process(StatusSuccess) + api.properties(VM.qemu_vm) +
      api.step_data('Start vm test_vm',
                    api.raw_io.output('VNC server running on 127.0.0.1:5900')) +
      # Verify that kvm was enabled for the first call
      api.post_process(StepCommandRE, 'Start vm test_vm (2)', [
          '.*qemu-system-x86_64', '-qmp', 'tcp:localhost:4444,server,nowait',
          '--serial', 'tcp:localhost:4445,server,nowait', '-daemonize', '-name',
          'test_vm', '--enable-kvm', '-m', '8192M', '-cpu',
          'Cascadelake-Server', '-machine', 'pc-q35-2.10', '-smp', 'cpus=4',
          '-device', 'ide-cd,drive=win10.iso', '-drive',
          'file=\[CLEANUP\]\/qemu\/workdir/disks/win10.iso,'
          'id=win10.iso,if=none,media=cdrom,format=raw,readonly,', '-display',
          'none'
      ]) +
      # Verify that vm was run with emulation (no --enable-kvm)
      api.post_process(StepCommandRE, 'Start vm test_vm', [
          '.*qemu-system-x86_64', '-qmp', 'tcp:localhost:4444,server,nowait',
          '--serial', 'tcp:localhost:4445,server,nowait', '-daemonize', '-name',
          'test_vm', '-m', '8192M', '-cpu', 'Cascadelake-Server', '-machine',
          'pc-q35-2.10', '-smp', 'cpus=4', '-device', 'ide-cd,drive=win10.iso',
          '-drive', 'file=\[CLEANUP\]\/qemu\/workdir/disks/win10.iso,'
          'id=win10.iso,if=none,media=cdrom,format=raw,readonly,', '-display',
          'none'
      ]) + api.post_process(DropExpectation))

  yield (
      api.test('Test start vm different guest and host arch') +
      api.platform('linux', 64, 'arm') + api.expect_exception('QEMUError') +
      api.post_process(StatusException) + api.properties(VM.qemu_vm) +
      api.step_data('Start vm test_vm',
                    api.raw_io.output('VNC server running on 127.0.0.1:5900')) +
      # Verify that first VM was run with emulation (no --enable-kvm)
      api.post_process(StepCommandRE, 'Start vm test_vm', [
          '.*qemu-system-x86_64', '-qmp', 'tcp:localhost:4444,server,nowait',
          '--serial', 'tcp:localhost:4445,server,nowait', '-daemonize', '-name',
          'test_vm', '-m', '8192M', '-cpu', 'Cascadelake-Server', '-machine',
          'pc-q35-2.10', '-smp', 'cpus=4', '-device', 'ide-cd,drive=win10.iso',
          '-drive', 'file=\[CLEANUP\]\/qemu\/workdir/disks/win10.iso,'
          'id=win10.iso,if=none,media=cdrom,format=raw,readonly,', '-display',
          'none'
      ]) + api.post_process(DropExpectation))

  yield (api.test('Test start vm fail') + api.post_process(StatusFailure) +
         api.properties(VM.qemu_vm) + api.step_data(
             'Start vm test_vm',
             api.raw_io.output('Failed to start vm'),
             retcode=1) + api.post_process(DropExpectation))
