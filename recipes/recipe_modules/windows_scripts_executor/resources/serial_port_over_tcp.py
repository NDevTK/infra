#!/usr/bin/python3

# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import argparse
import socket
import json
import sys
import base64
import select
import codecs
import time
import logging

HANDSHAKE_TIME = 2  # Time in seconds to wait before retrying handshake


# Extend argparse action to support dict type args
class ParseDict(argparse.Action):

  def __call__(self, parser, namespace, values, option_string=None):
    setattr(namespace, self.dest, dict())
    for element in values:
      k, v = element.split('=')
      getattr(namespace, self.dest)[k] = v


def send_expr(sock, expr, timeout=60, retries=1):
  """ send_expr sends a given expression to PS and returns the response

  Throws an exception if it exhausts all retry attempts.

  Args:
    sock: socket to send the expression over
    expr: dict representing the powershell expression request
    timeout: timeout in seconds to receive the response
    retries: number of times to retry the expression
  """
  while retries > 0:
    logging.info('Send req %s', expr)
    sock.sendall(json.dumps(expr).encode('utf-8'))
    res = resp(sock, timeout=timeout)
    if res:
      return res
    else:
      retries -= 1
  raise Exception('Timeout: {}'.format(expr))


def resp(sock, timeout=60):
  """ resp waits for response on socket. This is bound to be a complete json
  response. So wait until we get a json response and then return it.

  If there is no response after timeout number of seconds. It returns an empty
  response. Note that if the client powershell script is working as intended.
  We should never receive an empty response.

  Args:
    * sock: socket to read response from
    * timeout: time in seconds to wait for the response
  """
  data = ''
  response = {}
  timeout = timeout * (10**9)  # convert time in s to ns
  # Loop until we have timeout or we have received an incomplete message
  while timeout > 0:
    # record the start time. Using ns for int result.
    timer = time.monotonic_ns()
    # wait to receive any data, Timeout in seconds
    read_sock, _, _ = select.select([sock], [], [], timeout // (10**9))
    # update time spent and attempt to get rest of the json
    timeout -= (time.monotonic_ns() - timer)
    for s in read_sock:
      if s == sock:
        # loop until we get a valid json object
        try:
          data += sock.recv(8192).decode('utf-8')
          response = json.loads(data)
          # ping output is not encoded
          if 'Output' in response and response['Output']:
            # Output is always utf8 encoded
            response['Output'] = decode_logs(response['Output'], 'utf-8')
          if 'Logs' in response and response['Logs']:
            for f, log in response['Logs'].items():
              response['Logs'][f] = decode_logs(log, 'utf-8')
          if 'Error' in response and response['Error']:
            response['Error'] = decode_logs(response['Error'], 'utf-8')
          logging.info('Recv resp %s', response)
          return response
        except json.JSONDecodeError:
          # retry to get rest of the data
          pass


def handshake(sock, cont=True, retries=150):
  """ handshake performs a check on the powershell to see if we are connected

  Determines if the powershell is up. Sends an expression (int in 0-100 range)
  and waits for the expected output (the given integer). If all retries are
  exhausted, throws an error.

  Args:
    * sock: socket to connect to
    * cont: If true doesn't delete session after executing a handshake
    * retries: Number of times to retry the handshake. Default: 5 mins
  """
  cmd = {'Type': 'PING', 'Cont': cont}
  # check every 2 second(s) if the host is up. For 5 mins (default) [2*150 s]
  res = send_expr(sock, cmd, timeout=HANDSHAKE_TIME, retries=retries)
  if 'Output' in res and res['Output'] == "PONG":
    return res
  raise Exception('Handshake failed')


def decode_logs(log, encoding):
  ''' decode_logs decodes the given log file from base64 and decode the
  encoded string

  Args:
    * log: the contents of the log file to decode
    * encoding: encoding to read the string, ex utf-8
  '''
  log = base64.b64decode(log)
  return log.decode(encoding)


def main():
  desc = ''' Invoke an expression on the host machine.
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
      '-e',
      '--expr',
      metavar='expression',
      type=str,
      help='expression to execute')
  parser.add_argument(
      '-t',
      '--timeout',
      metavar='300',
      type=int,
      help='timeout in seconds',
      default=300)
  parser.add_argument(
      '-c',
      '--cont',
      action='store_true',
      default=False,
      help='continue the powershell session')
  parser.add_argument(
      '-l',
      '--let',
      action=ParseDict,
      nargs="*",
      metavar="foo=bar",
      help='Use the given context')
  parser.add_argument(
      '-d',
      '--debug',
      action='store_true',
      default=False,
      help='Drop a powershell prompt after executing command')
  parser.add_argument(
      '-L',
      '--log',
      metavar='C:\Windows\System32\Logs\Install.log',
      action="append",
      type=str)

  logging.basicConfig(
      stream=sys.stderr,
      level=logging.DEBUG,
      format='%(funcName)s [%(asctime)s] - %(message)s',
      style='%')

  args = parser.parse_args()

  if not args.debug and not args.expr:
    raise Exception('Nothing to do. No expression or debug flag given')

  logging.info('Running with %s', args)
  sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
  host, port = args.sock.split(':')
  sock.connect((host, int(port)))

  # Attempt to handshake for timeout secs.
  retries = args.timeout // HANDSHAKE_TIME
  res = handshake(sock, retries=retries, cont=True)

  # add given context to the session
  if args.let:
    for k, v in args.let.items():
      cmd = {'Type': 'Expr', 'Expr': '${} = {}'.format(k, v), 'Cont': True}
      response = send_expr(sock, cmd, timeout=args.timeout)

  if args.expr:
    cmd = {
        'Type': 'Expr',
        'Expr': args.expr,
        'Logs': args.log,
        'Cont': args.cont or args.debug
    }  # continue session if we are debugging
    response = send_expr(sock, cmd, timeout=args.timeout)
    json.dump(response, sys.stdout)

  if args.debug:
    # debug mode. Run powershell remotely.
    while True:
      cmd = {
          'Type': 'Expr',
          'Expr': input('PS {}>'.format(res['PWD'])),
          'Cont': True
      }
      res = send_expr(sock, cmd, timeout=args.timeout)
      if 'Output' in res and res['Output']:
        print(res['Output'])
      if 'Error' in res and res['Error']:
        print(res['Error'])

  sock.close()


main()
