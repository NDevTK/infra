# Generated by the pRPC protocol buffer compiler plugin.  DO NOT EDIT!
# source: go.chromium.org/luci/cv/api/migration/migration.proto

import base64
import zlib

from google.protobuf import descriptor_pb2

# Includes description of the go.chromium.org/luci/cv/api/migration/migration.proto and all of its transitive
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


MigrationServiceDescription = {
  'file_descriptor_set': FILE_DESCRIPTOR_SET,
  'file_descriptor': _INDEX[u'go.chromium.org/luci/cv/api/migration/migration.proto']['descriptor'],
  'service_descriptor': _INDEX[u'go.chromium.org/luci/cv/api/migration/migration.proto']['services'][u'Migration'],
}
