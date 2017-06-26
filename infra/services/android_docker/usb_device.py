# Copyright 2015 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""High-level wrapper of a usb device.

Wraps python-libusb1 to provide config/descriptions for a device. Does not
provide any low-level IO support over the bus.
"""

import collections
import os
import sys
import libusb1
import logging
import subprocess
import time
import usb1

from battor import battor_error
from devil.utils import battor_device_mapping
from devil.utils import find_usb_devices
from devil.utils import usb_hubs


if sys.platform != 'linux2':
  raise NotImplementedError('This library only supported on linux systems.')


_SUPPORTED_INTERFACES = [
  # (interface class, interface subclass, interface protocol)
  (255, 66, 1),  # ADB's definition.
  (255, 66, 3),  # Fastboot's definition.
]


def is_android_device(device):
  if not any(i in _SUPPORTED_INTERFACES for i in device.interfaces):
    return False
  if not device.serial:
    return False
  return True


def get_android_devices(filter_devices):
  ctx = usb1.USBContext()
  usb_devices = [USBDevice(d) for d in ctx.getDeviceList(skip_on_error=True)]
  android_devices = [d for d in usb_devices if is_android_device(d)]

  if not android_devices:
    logging.error('Unable to find devices: %s', filter_devices or 'all')
    return []

  # Determine the order in which the devices are physically plugged in. Can
  # only be done once all devices have been discovered.
  assign_physical_ports(android_devices)

  # Remove devices with duplicate serials. This can wreak havoc in container
  # management logic since each device's container is identified by its device's
  # presumed-to-be-unique serial.
  device_count = collections.defaultdict(int)
  for d in android_devices:
    device_count[d.serial] += 1
  for serial, count in device_count.iteritems():
    if count > 1:
      logging.error(
          'Ignoring device %s due to it appearing %d times.', serial, count)
  android_devices = [d for d in android_devices if device_count[d.serial] == 1]

  # Filter out the requested devices only after the physical ports have been
  # assigned.
  if filter_devices:
    android_devices = [d for d in android_devices if d.serial in filter_devices]
    if not android_devices:
      logging.error('Requested devices %s not found on host.', filter_devices)
      return []

  # Scan for connected battors and link each one to its android device if any
  # are present.
  battor_serial_map = {}
  try:
    battor_serial_map = battor_device_mapping.GenerateSerialMap()
  except battor_error.BattOrError:
    # No battors connected. Exit quietly since this is the case on most bots.
    pass
  except Exception:
    logging.exception('Unable to scan for battors; none will be added.')
  if battor_serial_map:
    # This loop may take several seconds, so log the performance.
    # TODO(bpastene): Maybe parallelize?
    start = time.time()
    for device in android_devices:
      try:
        battor_tty_path = battor_device_mapping.GetBattOrPathFromPhoneSerial(
            device.serial, serial_map=battor_serial_map)
      except battor_error.BattOrError:
        logging.exception(
            'Unable to get battor path for device %s.', device.serial)
        continue
      device.battor = BattorTTYDevice(
          battor_tty_path, battor_serial_map.get(device.serial, 'unknown'))
    logging.debug('Fetched battor serials in %.2fs.', time.time() - start)

  return android_devices


def assign_physical_ports(devices):
  """Based on usbfs port list, try to assign each device its physical port num.

  This corresponds to the order in which they're plugged into an external hub.
  """
  port_mapping = {}
  for hub in find_usb_devices.GetAllPhysicalPortToSerialMaps(
      usb_hubs.ALL_HUBS, fast=True):
    # Reverse the mapping.
    port_mapping.update({device: port for port, device in hub.iteritems()})
  for d in devices:
    d.physical_port = port_mapping.get(d.serial)


class USBDevice(object):
  def __init__(self, libusb_device):
    self._libusb_device = libusb_device

    self.port = libusb_device.getPortNumber()
    self.bus = libusb_device.getBusNumber()
    self.dev = libusb_device.getDeviceAddress()
    self.physical_port = None
    self._serial = None
    self._port_list = None

    # libusb exposes product and vendor IDs as decimal but sysfs reports
    # them as hex. Convert to hex for easy string comparison.
    self.product = hex(libusb_device.getProductID())[2:]

    # libusb doesn't expose major and minor numbers, so stat the device file.
    self.major = None
    self.minor = None
    self.dev_file_path = os.path.join(
        '/dev/bus/usb', '%03d' % self.bus, '%03d' % self.dev)
    try:
      st = os.stat(self.dev_file_path)
      self.major = os.major(st.st_rdev)
      self.minor = os.minor(st.st_rdev)
    except OSError:
      pass

    self.battor = None

  def __str__(self):
    return self.serial or self.port_list

  def __repr__(self):
    return self.serial or self.port_list

  @property
  def serial(self):
    if not self._serial:
      try:
        self._serial = self._libusb_device.getSerialNumber()
      except usb1.USBError:
        self._serial = None
    return self._serial

  @property
  def port_list(self):
    if not self._port_list:
      try:
        self._port_list = self._libusb_device.getPortNumberList()
      except usb1.USBError:
        self._port_list = None
    return self._port_list

  @property
  def interfaces(self):
    for setting in self._libusb_device.iterSettings():
      yield (setting.getClass(), setting.getSubClass(), setting.getProtocol())


class BattorTTYDevice(object):
  """Represents the virtual tty character device file of a battor.

  This differs from an android device because communication with it is done via
  its virtual tty device (/dev/tty*) and not its bus device (/dev/bus/usb/*),
  which seems to go unused.
  """
  def __init__(self, tty_path, serial):
    self.tty_path = tty_path
    self.serial = serial

    self.major = None
    self.minor = None
    try:
      st = os.stat(self.tty_path)
      self.major = os.major(st.st_rdev)
      self.minor = os.minor(st.st_rdev)
    except OSError:
      pass

    self._syspath = None

  @property
  def syspath(self):
    """Fetches the sysfs path of the battor by querying udev via udevadm.

    Expected output of udevadm resembles the following:
    E: DEVNAME=/dev/ttyUSB0
    E: DEVPATH=/devices/pci0000:00/0000:00:1c.7/0000:08:00.0/usb3/.../ttyUSB0
    E: ID_BUS=usb

    This extracts and returns the DEVPATH string.
    """
    if self._syspath is not None:
      return self._syspath
    try:
      out = subprocess.check_output(['/sbin/udevadm', 'info', self.tty_path])
    except subprocess.CalledProcessError:
      logging.exception('Unable to fetch syspath of battor %s', self.tty_path)
      return None
    for line in out.splitlines():
      if 'DEVPATH=' in line:
        self._syspath = line.split('DEVPATH=')[1]
        return self._syspath
