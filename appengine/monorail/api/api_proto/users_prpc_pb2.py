# Generated by the pRPC protocol buffer compiler plugin.  DO NOT EDIT!
# source: api/api_proto/users.proto

import base64
import zlib

from google.protobuf import descriptor_pb2

# Includes description of the api/api_proto/users.proto and all of its transitive
# dependencies. Includes source code info.
FILE_DESCRIPTOR_SET = descriptor_pb2.FileDescriptorSet()
FILE_DESCRIPTOR_SET.ParseFromString(zlib.decompress(base64.b64decode(
    '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')))
_INDEX = {
    f.name: {
      'descriptor': f,
      'services': {s.name: s for s in f.service},
    }
    for f in FILE_DESCRIPTOR_SET.file
}


UsersServiceDescription = {
  'file_descriptor_set': FILE_DESCRIPTOR_SET,
  'file_descriptor': _INDEX[u'api/api_proto/users.proto']['descriptor'],
  'service_descriptor': _INDEX[u'api/api_proto/users.proto']['services'][u'Users'],
}
