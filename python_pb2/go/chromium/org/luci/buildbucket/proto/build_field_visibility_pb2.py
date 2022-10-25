# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: go.chromium.org/luci/buildbucket/proto/build_field_visibility.proto
"""Generated protocol buffer code."""
from google.protobuf.internal import enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google.protobuf import descriptor_pb2 as google_dot_protobuf_dot_descriptor__pb2


DESCRIPTOR = _descriptor.FileDescriptor(
  name='go.chromium.org/luci/buildbucket/proto/build_field_visibility.proto',
  package='buildbucket.v2',
  syntax='proto3',
  serialized_options=b'Z4go.chromium.org/luci/buildbucket/proto;buildbucketpb',
  create_key=_descriptor._internal_create_key,
  serialized_pb=b'\nCgo.chromium.org/luci/buildbucket/proto/build_field_visibility.proto\x12\x0e\x62uildbucket.v2\x1a google/protobuf/descriptor.proto*\x92\x01\n\x14\x42uildFieldVisibility\x12 \n\x1c\x46IELD_VISIBILITY_UNSPECIFIED\x10\x00\x12\x19\n\x15\x42UILDS_GET_PERMISSION\x10\x01\x12!\n\x1d\x42UILDS_GET_LIMITED_PERMISSION\x10\x02\x12\x1a\n\x16\x42UILDS_LIST_PERMISSION\x10\x03:[\n\x0cvisible_with\x12\x1d.google.protobuf.FieldOptions\x18\xe7\xc9\x37 \x01(\x0e\x32$.buildbucket.v2.BuildFieldVisibilityB6Z4go.chromium.org/luci/buildbucket/proto;buildbucketpbb\x06proto3'
  ,
  dependencies=[google_dot_protobuf_dot_descriptor__pb2.DESCRIPTOR,])

_BUILDFIELDVISIBILITY = _descriptor.EnumDescriptor(
  name='BuildFieldVisibility',
  full_name='buildbucket.v2.BuildFieldVisibility',
  filename=None,
  file=DESCRIPTOR,
  create_key=_descriptor._internal_create_key,
  values=[
    _descriptor.EnumValueDescriptor(
      name='FIELD_VISIBILITY_UNSPECIFIED', index=0, number=0,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='BUILDS_GET_PERMISSION', index=1, number=1,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='BUILDS_GET_LIMITED_PERMISSION', index=2, number=2,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='BUILDS_LIST_PERMISSION', index=3, number=3,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
  ],
  containing_type=None,
  serialized_options=None,
  serialized_start=122,
  serialized_end=268,
)
_sym_db.RegisterEnumDescriptor(_BUILDFIELDVISIBILITY)

BuildFieldVisibility = enum_type_wrapper.EnumTypeWrapper(_BUILDFIELDVISIBILITY)
FIELD_VISIBILITY_UNSPECIFIED = 0
BUILDS_GET_PERMISSION = 1
BUILDS_GET_LIMITED_PERMISSION = 2
BUILDS_LIST_PERMISSION = 3

VISIBLE_WITH_FIELD_NUMBER = 910567
visible_with = _descriptor.FieldDescriptor(
  name='visible_with', full_name='buildbucket.v2.visible_with', index=0,
  number=910567, type=14, cpp_type=8, label=1,
  has_default_value=False, default_value=0,
  message_type=None, enum_type=None, containing_type=None,
  is_extension=True, extension_scope=None,
  serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key)

DESCRIPTOR.enum_types_by_name['BuildFieldVisibility'] = _BUILDFIELDVISIBILITY
DESCRIPTOR.extensions_by_name['visible_with'] = visible_with
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

visible_with.enum_type = _BUILDFIELDVISIBILITY
google_dot_protobuf_dot_descriptor__pb2.FieldOptions.RegisterExtension(visible_with)

DESCRIPTOR._options = None
# @@protoc_insertion_point(module_scope)
