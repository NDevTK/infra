# Generated by the pRPC protocol buffer compiler plugin.  DO NOT EDIT!
# source: api/api_proto/issues.proto

import base64
import zlib

from google.protobuf import descriptor_pb2

# Includes description of the api/api_proto/issues.proto and all of its transitive
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


IssuesServiceDescription = {
  'file_descriptor_set': FILE_DESCRIPTOR_SET,
  'file_descriptor': _INDEX[u'api/api_proto/issues.proto']['descriptor'],
  'service_descriptor': _INDEX[u'api/api_proto/issues.proto']['services'][u'Issues'],
}
