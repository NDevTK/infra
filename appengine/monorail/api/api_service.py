# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is govered by a BSD-style
# license that can be found in the LICENSE file or at
# https://developers.google.com/open-source/licenses/bsd

# This file implements a pRPC API for Monorail.
#
# See the pRPC spec here: https://godoc.org/github.com/luci/luci-go/grpc/prpc
#
# Each Servicer corresponds to a service defined in a .proto file in this
# directory. Each method on that Servicer corresponds to one of the rpcs
# defined on the service.
#
# All APIs are served under the /prpc/* path space. Each service gets its own
# namespace under that, and each method is an individual endpoints. For example,
#   POST https://bugs.chromium.org/prpc/monorail.Users/GetUser
# would be a call to the UsersServicer.GetUser method.
#
# Note that this is not a RESTful API, although it is CRUDy. All requests are
# POSTs, all methods take exactly one input, and all methods return exactly
# one output.
#
# TODO(agable): Actually integrate the rpcexplorer.
# You can use the API Explorer here: https://bugs.chromium.org/rpcexplorer

from components import prpc

from api import monorail_pb2
from api import monorail_prpc_pb2

class UsersServicer(object):

  DESCRIPTION = monorail_prpc_pb2.UsersServiceDescription

  def GetUser(self, request, _context):
    ret = monorail_pb2.User()
    ret.email = request.email
    ret.id = hash(request.email)
    return ret


class IssuesServicer(object):

  DESCRIPTION = monorail_prpc_pb2.IssuesServiceDescription

  def ComponentBurndown(self, request, _context):
    res = monorail_pb2.ComponentBurndownResponse()
    res.project = request.project
    res.component_path = request.component_path
    return res


def RegisterApiHandlers(registry):
  server = prpc.Server()
  server.add_service(UsersServicer())
  registry.routes.extend(server.get_routes())
