# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: api/v3/api_proto/permission_objects.proto
"""Generated protocol buffer code."""
from google.protobuf.internal import builder as _builder
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n)api/v3/api_proto/permission_objects.proto\x12\x0bmonorail.v3\"O\n\rPermissionSet\x12\x10\n\x08resource\x18\x01 \x01(\t\x12,\n\x0bpermissions\x18\x02 \x03(\x0e\x32\x17.monorail.v3.Permission*\x90\x01\n\nPermission\x12\x1a\n\x16PERMISSION_UNSPECIFIED\x10\x00\x12\x10\n\x0cHOTLIST_EDIT\x10\x01\x12\x16\n\x12HOTLIST_ADMINISTER\x10\x02\x12\x0e\n\nISSUE_EDIT\x10\x03\x12\x12\n\x0e\x46IELD_DEF_EDIT\x10\x04\x12\x18\n\x14\x46IELD_DEF_VALUE_EDIT\x10\x05\x42#Z!infra/monorailv2/api/v3/api_protob\x06proto3')

_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, globals())
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'api.v3.api_proto.permission_objects_pb2', globals())
if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'Z!infra/monorailv2/api/v3/api_proto'
  _PERMISSION._serialized_start=140
  _PERMISSION._serialized_end=284
  _PERMISSIONSET._serialized_start=58
  _PERMISSIONSET._serialized_end=137
# @@protoc_insertion_point(module_scope)