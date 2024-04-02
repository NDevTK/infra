# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: go.chromium.org/luci/resultdb/proto/v1/artifact.proto
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google.api import field_behavior_pb2 as google_dot_api_dot_field__behavior__pb2
from google.protobuf import timestamp_pb2 as google_dot_protobuf_dot_timestamp__pb2
from go.chromium.org.luci.resultdb.proto.v1 import test_result_pb2 as go_dot_chromium_dot_org_dot_luci_dot_resultdb_dot_proto_dot_v1_dot_test__result__pb2


DESCRIPTOR = _descriptor.FileDescriptor(
  name='go.chromium.org/luci/resultdb/proto/v1/artifact.proto',
  package='luci.resultdb.v1',
  syntax='proto3',
  serialized_options=b'Z/go.chromium.org/luci/resultdb/proto/v1;resultpb',
  create_key=_descriptor._internal_create_key,
  serialized_pb=b'\n5go.chromium.org/luci/resultdb/proto/v1/artifact.proto\x12\x10luci.resultdb.v1\x1a\x1fgoogle/api/field_behavior.proto\x1a\x1fgoogle/protobuf/timestamp.proto\x1a\x38go.chromium.org/luci/resultdb/proto/v1/test_result.proto\"\xff\x01\n\x08\x41rtifact\x12\x0c\n\x04name\x18\x01 \x01(\t\x12\x13\n\x0b\x61rtifact_id\x18\x02 \x01(\t\x12\x11\n\tfetch_url\x18\x03 \x01(\t\x12\x38\n\x14\x66\x65tch_url_expiration\x18\x04 \x01(\x0b\x32\x1a.google.protobuf.Timestamp\x12\x14\n\x0c\x63ontent_type\x18\x05 \x01(\t\x12\x12\n\nsize_bytes\x18\x06 \x01(\x03\x12\x15\n\x08\x63ontents\x18\x07 \x01(\x0c\x42\x03\xe0\x41\x04\x12\x0f\n\x07gcs_uri\x18\x08 \x01(\t\x12\x31\n\x0btest_status\x18\t \x01(\x0e\x32\x1c.luci.resultdb.v1.TestStatusB1Z/go.chromium.org/luci/resultdb/proto/v1;resultpbb\x06proto3'
  ,
  dependencies=[google_dot_api_dot_field__behavior__pb2.DESCRIPTOR,google_dot_protobuf_dot_timestamp__pb2.DESCRIPTOR,go_dot_chromium_dot_org_dot_luci_dot_resultdb_dot_proto_dot_v1_dot_test__result__pb2.DESCRIPTOR,])




_ARTIFACT = _descriptor.Descriptor(
  name='Artifact',
  full_name='luci.resultdb.v1.Artifact',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='name', full_name='luci.resultdb.v1.Artifact.name', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='artifact_id', full_name='luci.resultdb.v1.Artifact.artifact_id', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='fetch_url', full_name='luci.resultdb.v1.Artifact.fetch_url', index=2,
      number=3, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='fetch_url_expiration', full_name='luci.resultdb.v1.Artifact.fetch_url_expiration', index=3,
      number=4, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='content_type', full_name='luci.resultdb.v1.Artifact.content_type', index=4,
      number=5, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='size_bytes', full_name='luci.resultdb.v1.Artifact.size_bytes', index=5,
      number=6, type=3, cpp_type=2, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='contents', full_name='luci.resultdb.v1.Artifact.contents', index=6,
      number=7, type=12, cpp_type=9, label=1,
      has_default_value=False, default_value=b"",
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=b'\340A\004', file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='gcs_uri', full_name='luci.resultdb.v1.Artifact.gcs_uri', index=7,
      number=8, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='test_status', full_name='luci.resultdb.v1.Artifact.test_status', index=8,
      number=9, type=14, cpp_type=8, label=1,
      has_default_value=False, default_value=0,
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
  serialized_start=200,
  serialized_end=455,
)

_ARTIFACT.fields_by_name['fetch_url_expiration'].message_type = google_dot_protobuf_dot_timestamp__pb2._TIMESTAMP
_ARTIFACT.fields_by_name['test_status'].enum_type = go_dot_chromium_dot_org_dot_luci_dot_resultdb_dot_proto_dot_v1_dot_test__result__pb2._TESTSTATUS
DESCRIPTOR.message_types_by_name['Artifact'] = _ARTIFACT
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

Artifact = _reflection.GeneratedProtocolMessageType('Artifact', (_message.Message,), {
  'DESCRIPTOR' : _ARTIFACT,
  '__module__' : 'go.chromium.org.luci.resultdb.proto.v1.artifact_pb2'
  # @@protoc_insertion_point(class_scope:luci.resultdb.v1.Artifact)
  })
_sym_db.RegisterMessage(Artifact)


DESCRIPTOR._options = None
_ARTIFACT.fields_by_name['contents']._options = None
# @@protoc_insertion_point(module_scope)
