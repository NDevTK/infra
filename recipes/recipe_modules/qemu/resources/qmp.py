#!/usr/bin/python3

# Copyright 2021 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import argparse
import socket
import json
import sys


# Extend argparse action to support dict type args
class ParseDict(argparse.Action):

  def __call__(self, parser, namespace, values, option_string=None):
    setattr(namespace, self.dest, dict())
    for element in values:
      k, v = element.split('=')
      getattr(namespace, self.dest)[k] = v


def send_cmd(sock, cmd):
  """ send_cmd sends a given command to qmp socket and returns the response
  """
  sock.sendall(json.dumps(cmd).encode('utf-8'))
  return resp(sock)


def resp(sock):
  """ resp waits for response on socket. This is bound to be a complete json
      response. So wait until we get a json response and then return it.
  """
  incomplete = True
  data = ''
  while incomplete:
    # loop until we get a valid json object
    try:
      data += sock.recv(1024).decode('utf-8')
      json.loads(data)
      incomplete = False
    except json.JSONDecodeError:
      # might be incomplete json
      incomplete = True
  return json.loads(data)


def main():
  desc = ''' Run a QMP command and return the results as a dictionary.
             See: https://www.qemu.org/docs/master/interop/qemu-qmp-ref.html
         '''

  parser = argparse.ArgumentParser(description=desc)
  parser.add_argument(
      '-s',
      '--sock',
      metavar='host:port',
      type=str,
      help='ipv4 host and port',
      required=True)
  parser.add_argument(
      '-c',
      '--cmd',
      metavar='command',
      type=str,
      help='qmp command to query',
      required=True)
  parser.add_argument(
      '-a', '--args', metavar='device=ide1-cd0', nargs="*", action=ParseDict)

  args = parser.parse_args()

  sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
  host, port = args.sock.split(':')
  sock.connect((host, int(port)))

  # read the first response on connect and ignore it
  resp(sock)

  # enable capabilities for qmp
  send_cmd(sock, {'execute': 'qmp_capabilities'})

  # execute the given command and return the response
  cmd = {'execute': args.cmd}
  if args.args:
    cmd['arguments'] = args.args
  json.dump(send_cmd(sock, cmd), sys.stdout)

  sock.close()


main()
