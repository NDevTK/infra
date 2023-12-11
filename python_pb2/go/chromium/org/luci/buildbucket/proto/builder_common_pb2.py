# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: go.chromium.org/luci/buildbucket/proto/builder_common.proto
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from go.chromium.org.luci.buildbucket.proto import common_pb2 as go_dot_chromium_dot_org_dot_luci_dot_buildbucket_dot_proto_dot_common__pb2
from go.chromium.org.luci.buildbucket.proto import field_option_pb2 as go_dot_chromium_dot_org_dot_luci_dot_buildbucket_dot_proto_dot_field__option__pb2
from go.chromium.org.luci.buildbucket.proto import project_config_pb2 as go_dot_chromium_dot_org_dot_luci_dot_buildbucket_dot_proto_dot_project__config__pb2


DESCRIPTOR = _descriptor.FileDescriptor(
  name='go.chromium.org/luci/buildbucket/proto/builder_common.proto',
  package='buildbucket.v2',
  syntax='proto3',
  serialized_options=b'Z4go.chromium.org/luci/buildbucket/proto;buildbucketpb',
  create_key=_descriptor._internal_create_key,
  serialized_pb=b'\n;go.chromium.org/luci/buildbucket/proto/builder_common.proto\x12\x0e\x62uildbucket.v2\x1a\x33go.chromium.org/luci/buildbucket/proto/common.proto\x1a\x39go.chromium.org/luci/buildbucket/proto/field_option.proto\x1a;go.chromium.org/luci/buildbucket/proto/project_config.proto\"\x7f\n\tBuilderID\x12%\n\x07project\x18\x01 \x01(\tB\x14\x9a\xc3\x1a\x10SetBuilderHealth\x12$\n\x06\x62ucket\x18\x02 \x01(\tB\x14\x9a\xc3\x1a\x10SetBuilderHealth\x12%\n\x07\x62uilder\x18\x03 \x01(\tB\x14\x9a\xc3\x1a\x10SetBuilderHealth\"N\n\x0f\x42uilderMetadata\x12\r\n\x05owner\x18\x01 \x01(\t\x12,\n\x06health\x18\x02 \x01(\x0b\x32\x1c.buildbucket.v2.HealthStatus\"\x93\x01\n\x0b\x42uilderItem\x12%\n\x02id\x18\x01 \x01(\x0b\x32\x19.buildbucket.v2.BuilderID\x12*\n\x06\x63onfig\x18\x02 \x01(\x0b\x32\x1a.buildbucket.BuilderConfig\x12\x31\n\x08metadata\x18\x03 \x01(\x0b\x32\x1f.buildbucket.v2.BuilderMetadataB6Z4go.chromium.org/luci/buildbucket/proto;buildbucketpbb\x06proto3'
  ,
  dependencies=[go_dot_chromium_dot_org_dot_luci_dot_buildbucket_dot_proto_dot_common__pb2.DESCRIPTOR,go_dot_chromium_dot_org_dot_luci_dot_buildbucket_dot_proto_dot_field__option__pb2.DESCRIPTOR,go_dot_chromium_dot_org_dot_luci_dot_buildbucket_dot_proto_dot_project__config__pb2.DESCRIPTOR,])




_BUILDERID = _descriptor.Descriptor(
  name='BuilderID',
  full_name='buildbucket.v2.BuilderID',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='project', full_name='buildbucket.v2.BuilderID.project', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=b'\232\303\032\020SetBuilderHealth', file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='bucket', full_name='buildbucket.v2.BuilderID.bucket', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=b'\232\303\032\020SetBuilderHealth', file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='builder', full_name='buildbucket.v2.BuilderID.builder', index=2,
      number=3, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=b'\232\303\032\020SetBuilderHealth', file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=252,
  serialized_end=379,
)


_BUILDERMETADATA = _descriptor.Descriptor(
  name='BuilderMetadata',
  full_name='buildbucket.v2.BuilderMetadata',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='owner', full_name='buildbucket.v2.BuilderMetadata.owner', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='health', full_name='buildbucket.v2.BuilderMetadata.health', index=1,
      number=2, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=381,
  serialized_end=459,
)


_BUILDERITEM = _descriptor.Descriptor(
  name='BuilderItem',
  full_name='buildbucket.v2.BuilderItem',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='id', full_name='buildbucket.v2.BuilderItem.id', index=0,
      number=1, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='config', full_name='buildbucket.v2.BuilderItem.config', index=1,
      number=2, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='metadata', full_name='buildbucket.v2.BuilderItem.metadata', index=2,
      number=3, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=462,
  serialized_end=609,
)

_BUILDERMETADATA.fields_by_name['health'].message_type = go_dot_chromium_dot_org_dot_luci_dot_buildbucket_dot_proto_dot_common__pb2._HEALTHSTATUS
_BUILDERITEM.fields_by_name['id'].message_type = _BUILDERID
_BUILDERITEM.fields_by_name['config'].message_type = go_dot_chromium_dot_org_dot_luci_dot_buildbucket_dot_proto_dot_project__config__pb2._BUILDERCONFIG
_BUILDERITEM.fields_by_name['metadata'].message_type = _BUILDERMETADATA
DESCRIPTOR.message_types_by_name['BuilderID'] = _BUILDERID
DESCRIPTOR.message_types_by_name['BuilderMetadata'] = _BUILDERMETADATA
DESCRIPTOR.message_types_by_name['BuilderItem'] = _BUILDERITEM
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

BuilderID = _reflection.GeneratedProtocolMessageType('BuilderID', (_message.Message,), {
  'DESCRIPTOR' : _BUILDERID,
  '__module__' : 'go.chromium.org.luci.buildbucket.proto.builder_common_pb2'
  # @@protoc_insertion_point(class_scope:buildbucket.v2.BuilderID)
  })
_sym_db.RegisterMessage(BuilderID)

BuilderMetadata = _reflection.GeneratedProtocolMessageType('BuilderMetadata', (_message.Message,), {
  'DESCRIPTOR' : _BUILDERMETADATA,
  '__module__' : 'go.chromium.org.luci.buildbucket.proto.builder_common_pb2'
  # @@protoc_insertion_point(class_scope:buildbucket.v2.BuilderMetadata)
  })
_sym_db.RegisterMessage(BuilderMetadata)

BuilderItem = _reflection.GeneratedProtocolMessageType('BuilderItem', (_message.Message,), {
  'DESCRIPTOR' : _BUILDERITEM,
  '__module__' : 'go.chromium.org.luci.buildbucket.proto.builder_common_pb2'
  # @@protoc_insertion_point(class_scope:buildbucket.v2.BuilderItem)
  })
_sym_db.RegisterMessage(BuilderItem)


DESCRIPTOR._options = None
_BUILDERID.fields_by_name['project']._options = None
_BUILDERID.fields_by_name['bucket']._options = None
_BUILDERID.fields_by_name['builder']._options = None
# @@protoc_insertion_point(module_scope)