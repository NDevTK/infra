# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: go.chromium.org/luci/resultdb/proto/v1/test_result.proto
"""Generated protocol buffer code."""
from google.protobuf.internal import enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google.api import field_behavior_pb2 as google_dot_api_dot_field__behavior__pb2
from google.protobuf import duration_pb2 as google_dot_protobuf_dot_duration__pb2
from google.protobuf import struct_pb2 as google_dot_protobuf_dot_struct__pb2
from google.protobuf import timestamp_pb2 as google_dot_protobuf_dot_timestamp__pb2
from go.chromium.org.luci.resultdb.proto.v1 import common_pb2 as go_dot_chromium_dot_org_dot_luci_dot_resultdb_dot_proto_dot_v1_dot_common__pb2
from go.chromium.org.luci.resultdb.proto.v1 import test_metadata_pb2 as go_dot_chromium_dot_org_dot_luci_dot_resultdb_dot_proto_dot_v1_dot_test__metadata__pb2
from go.chromium.org.luci.resultdb.proto.v1 import failure_reason_pb2 as go_dot_chromium_dot_org_dot_luci_dot_resultdb_dot_proto_dot_v1_dot_failure__reason__pb2


DESCRIPTOR = _descriptor.FileDescriptor(
  name='go.chromium.org/luci/resultdb/proto/v1/test_result.proto',
  package='luci.resultdb.v1',
  syntax='proto3',
  serialized_options=b'Z/go.chromium.org/luci/resultdb/proto/v1;resultpb',
  create_key=_descriptor._internal_create_key,
  serialized_pb=b'\n8go.chromium.org/luci/resultdb/proto/v1/test_result.proto\x12\x10luci.resultdb.v1\x1a\x1fgoogle/api/field_behavior.proto\x1a\x1egoogle/protobuf/duration.proto\x1a\x1cgoogle/protobuf/struct.proto\x1a\x1fgoogle/protobuf/timestamp.proto\x1a\x33go.chromium.org/luci/resultdb/proto/v1/common.proto\x1a:go.chromium.org/luci/resultdb/proto/v1/test_metadata.proto\x1a;go.chromium.org/luci/resultdb/proto/v1/failure_reason.proto\"\x8d\x05\n\nTestResult\x12\x14\n\x04name\x18\x01 \x01(\tB\x06\xe0\x41\x03\xe0\x41\x05\x12\x14\n\x07test_id\x18\x02 \x01(\tB\x03\xe0\x41\x05\x12\x19\n\tresult_id\x18\x03 \x01(\tB\x06\xe0\x41\x05\xe0\x41\x02\x12/\n\x07variant\x18\x04 \x01(\x0b\x32\x19.luci.resultdb.v1.VariantB\x03\xe0\x41\x05\x12\x15\n\x08\x65xpected\x18\x05 \x01(\x08\x42\x03\xe0\x41\x05\x12\x31\n\x06status\x18\x06 \x01(\x0e\x32\x1c.luci.resultdb.v1.TestStatusB\x03\xe0\x41\x05\x12\x19\n\x0csummary_html\x18\x07 \x01(\tB\x03\xe0\x41\x05\x12\x33\n\nstart_time\x18\x08 \x01(\x0b\x32\x1a.google.protobuf.TimestampB\x03\xe0\x41\x05\x12\x30\n\x08\x64uration\x18\t \x01(\x0b\x32\x19.google.protobuf.DurationB\x03\xe0\x41\x05\x12/\n\x04tags\x18\n \x03(\x0b\x32\x1c.luci.resultdb.v1.StringPairB\x03\xe0\x41\x05\x12\x1c\n\x0cvariant_hash\x18\x0c \x01(\tB\x06\xe0\x41\x03\xe0\x41\x05\x12\x35\n\rtest_metadata\x18\r \x01(\x0b\x32\x1e.luci.resultdb.v1.TestMetadata\x12\x37\n\x0e\x66\x61ilure_reason\x18\x0e \x01(\x0b\x32\x1f.luci.resultdb.v1.FailureReason\x12+\n\nproperties\x18\x0f \x01(\x0b\x32\x17.google.protobuf.Struct\x12\x16\n\tis_masked\x18\x10 \x01(\x08\x42\x03\xe0\x41\x03\x12\x31\n\x0bskip_reason\x18\x12 \x01(\x0e\x32\x1c.luci.resultdb.v1.SkipReasonJ\x04\x08\x0b\x10\x0c\"\x90\x02\n\x0fTestExoneration\x12\x14\n\x04name\x18\x01 \x01(\tB\x06\xe0\x41\x03\xe0\x41\x05\x12\x0f\n\x07test_id\x18\x02 \x01(\t\x12*\n\x07variant\x18\x03 \x01(\x0b\x32\x19.luci.resultdb.v1.Variant\x12\x1e\n\x0e\x65xoneration_id\x18\x04 \x01(\tB\x06\xe0\x41\x03\xe0\x41\x05\x12\x1d\n\x10\x65xplanation_html\x18\x05 \x01(\tB\x03\xe0\x41\x05\x12\x19\n\x0cvariant_hash\x18\x06 \x01(\tB\x03\xe0\x41\x05\x12\x38\n\x06reason\x18\x07 \x01(\x0e\x32#.luci.resultdb.v1.ExonerationReasonB\x03\xe0\x41\x05\x12\x16\n\tis_masked\x18\x08 \x01(\x08\x42\x03\xe0\x41\x03*X\n\nTestStatus\x12\x16\n\x12STATUS_UNSPECIFIED\x10\x00\x12\x08\n\x04PASS\x10\x01\x12\x08\n\x04\x46\x41IL\x10\x02\x12\t\n\x05\x43RASH\x10\x03\x12\t\n\x05\x41\x42ORT\x10\x04\x12\x08\n\x04SKIP\x10\x05*S\n\nSkipReason\x12\x1b\n\x17SKIP_REASON_UNSPECIFIED\x10\x00\x12(\n$AUTOMATICALLY_DISABLED_FOR_FLAKINESS\x10\x01*\x8f\x01\n\x11\x45xonerationReason\x12\"\n\x1e\x45XONERATION_REASON_UNSPECIFIED\x10\x00\x12\x16\n\x12OCCURS_ON_MAINLINE\x10\x01\x12\x17\n\x13OCCURS_ON_OTHER_CLS\x10\x02\x12\x10\n\x0cNOT_CRITICAL\x10\x03\x12\x13\n\x0fUNEXPECTED_PASS\x10\x04\x42\x31Z/go.chromium.org/luci/resultdb/proto/v1;resultpbb\x06proto3'
  ,
  dependencies=[google_dot_api_dot_field__behavior__pb2.DESCRIPTOR,google_dot_protobuf_dot_duration__pb2.DESCRIPTOR,google_dot_protobuf_dot_struct__pb2.DESCRIPTOR,google_dot_protobuf_dot_timestamp__pb2.DESCRIPTOR,go_dot_chromium_dot_org_dot_luci_dot_resultdb_dot_proto_dot_v1_dot_common__pb2.DESCRIPTOR,go_dot_chromium_dot_org_dot_luci_dot_resultdb_dot_proto_dot_v1_dot_test__metadata__pb2.DESCRIPTOR,go_dot_chromium_dot_org_dot_luci_dot_resultdb_dot_proto_dot_v1_dot_failure__reason__pb2.DESCRIPTOR,])

_TESTSTATUS = _descriptor.EnumDescriptor(
  name='TestStatus',
  full_name='luci.resultdb.v1.TestStatus',
  filename=None,
  file=DESCRIPTOR,
  create_key=_descriptor._internal_create_key,
  values=[
    _descriptor.EnumValueDescriptor(
      name='STATUS_UNSPECIFIED', index=0, number=0,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='PASS', index=1, number=1,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='FAIL', index=2, number=2,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='CRASH', index=3, number=3,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='ABORT', index=4, number=4,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='SKIP', index=5, number=5,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
  ],
  containing_type=None,
  serialized_options=None,
  serialized_start=1311,
  serialized_end=1399,
)
_sym_db.RegisterEnumDescriptor(_TESTSTATUS)

TestStatus = enum_type_wrapper.EnumTypeWrapper(_TESTSTATUS)
_SKIPREASON = _descriptor.EnumDescriptor(
  name='SkipReason',
  full_name='luci.resultdb.v1.SkipReason',
  filename=None,
  file=DESCRIPTOR,
  create_key=_descriptor._internal_create_key,
  values=[
    _descriptor.EnumValueDescriptor(
      name='SKIP_REASON_UNSPECIFIED', index=0, number=0,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='AUTOMATICALLY_DISABLED_FOR_FLAKINESS', index=1, number=1,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
  ],
  containing_type=None,
  serialized_options=None,
  serialized_start=1401,
  serialized_end=1484,
)
_sym_db.RegisterEnumDescriptor(_SKIPREASON)

SkipReason = enum_type_wrapper.EnumTypeWrapper(_SKIPREASON)
_EXONERATIONREASON = _descriptor.EnumDescriptor(
  name='ExonerationReason',
  full_name='luci.resultdb.v1.ExonerationReason',
  filename=None,
  file=DESCRIPTOR,
  create_key=_descriptor._internal_create_key,
  values=[
    _descriptor.EnumValueDescriptor(
      name='EXONERATION_REASON_UNSPECIFIED', index=0, number=0,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='OCCURS_ON_MAINLINE', index=1, number=1,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='OCCURS_ON_OTHER_CLS', index=2, number=2,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='NOT_CRITICAL', index=3, number=3,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='UNEXPECTED_PASS', index=4, number=4,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
  ],
  containing_type=None,
  serialized_options=None,
  serialized_start=1487,
  serialized_end=1630,
)
_sym_db.RegisterEnumDescriptor(_EXONERATIONREASON)

ExonerationReason = enum_type_wrapper.EnumTypeWrapper(_EXONERATIONREASON)
STATUS_UNSPECIFIED = 0
PASS = 1
FAIL = 2
CRASH = 3
ABORT = 4
SKIP = 5
SKIP_REASON_UNSPECIFIED = 0
AUTOMATICALLY_DISABLED_FOR_FLAKINESS = 1
EXONERATION_REASON_UNSPECIFIED = 0
OCCURS_ON_MAINLINE = 1
OCCURS_ON_OTHER_CLS = 2
NOT_CRITICAL = 3
UNEXPECTED_PASS = 4



_TESTRESULT = _descriptor.Descriptor(
  name='TestResult',
  full_name='luci.resultdb.v1.TestResult',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='name', full_name='luci.resultdb.v1.TestResult.name', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=b'\340A\003\340A\005', file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='test_id', full_name='luci.resultdb.v1.TestResult.test_id', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=b'\340A\005', file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='result_id', full_name='luci.resultdb.v1.TestResult.result_id', index=2,
      number=3, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=b'\340A\005\340A\002', file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='variant', full_name='luci.resultdb.v1.TestResult.variant', index=3,
      number=4, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=b'\340A\005', file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='expected', full_name='luci.resultdb.v1.TestResult.expected', index=4,
      number=5, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=b'\340A\005', file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='status', full_name='luci.resultdb.v1.TestResult.status', index=5,
      number=6, type=14, cpp_type=8, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=b'\340A\005', file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='summary_html', full_name='luci.resultdb.v1.TestResult.summary_html', index=6,
      number=7, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=b'\340A\005', file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='start_time', full_name='luci.resultdb.v1.TestResult.start_time', index=7,
      number=8, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=b'\340A\005', file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='duration', full_name='luci.resultdb.v1.TestResult.duration', index=8,
      number=9, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=b'\340A\005', file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='tags', full_name='luci.resultdb.v1.TestResult.tags', index=9,
      number=10, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=b'\340A\005', file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='variant_hash', full_name='luci.resultdb.v1.TestResult.variant_hash', index=10,
      number=12, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=b'\340A\003\340A\005', file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='test_metadata', full_name='luci.resultdb.v1.TestResult.test_metadata', index=11,
      number=13, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='failure_reason', full_name='luci.resultdb.v1.TestResult.failure_reason', index=12,
      number=14, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='properties', full_name='luci.resultdb.v1.TestResult.properties', index=13,
      number=15, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='is_masked', full_name='luci.resultdb.v1.TestResult.is_masked', index=14,
      number=16, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=b'\340A\003', file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='skip_reason', full_name='luci.resultdb.v1.TestResult.skip_reason', index=15,
      number=18, type=14, cpp_type=8, label=1,
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
  serialized_start=381,
  serialized_end=1034,
)


_TESTEXONERATION = _descriptor.Descriptor(
  name='TestExoneration',
  full_name='luci.resultdb.v1.TestExoneration',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='name', full_name='luci.resultdb.v1.TestExoneration.name', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=b'\340A\003\340A\005', file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='test_id', full_name='luci.resultdb.v1.TestExoneration.test_id', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='variant', full_name='luci.resultdb.v1.TestExoneration.variant', index=2,
      number=3, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='exoneration_id', full_name='luci.resultdb.v1.TestExoneration.exoneration_id', index=3,
      number=4, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=b'\340A\003\340A\005', file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='explanation_html', full_name='luci.resultdb.v1.TestExoneration.explanation_html', index=4,
      number=5, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=b'\340A\005', file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='variant_hash', full_name='luci.resultdb.v1.TestExoneration.variant_hash', index=5,
      number=6, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=b'\340A\005', file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='reason', full_name='luci.resultdb.v1.TestExoneration.reason', index=6,
      number=7, type=14, cpp_type=8, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=b'\340A\005', file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='is_masked', full_name='luci.resultdb.v1.TestExoneration.is_masked', index=7,
      number=8, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=b'\340A\003', file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
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
  serialized_start=1037,
  serialized_end=1309,
)

_TESTRESULT.fields_by_name['variant'].message_type = go_dot_chromium_dot_org_dot_luci_dot_resultdb_dot_proto_dot_v1_dot_common__pb2._VARIANT
_TESTRESULT.fields_by_name['status'].enum_type = _TESTSTATUS
_TESTRESULT.fields_by_name['start_time'].message_type = google_dot_protobuf_dot_timestamp__pb2._TIMESTAMP
_TESTRESULT.fields_by_name['duration'].message_type = google_dot_protobuf_dot_duration__pb2._DURATION
_TESTRESULT.fields_by_name['tags'].message_type = go_dot_chromium_dot_org_dot_luci_dot_resultdb_dot_proto_dot_v1_dot_common__pb2._STRINGPAIR
_TESTRESULT.fields_by_name['test_metadata'].message_type = go_dot_chromium_dot_org_dot_luci_dot_resultdb_dot_proto_dot_v1_dot_test__metadata__pb2._TESTMETADATA
_TESTRESULT.fields_by_name['failure_reason'].message_type = go_dot_chromium_dot_org_dot_luci_dot_resultdb_dot_proto_dot_v1_dot_failure__reason__pb2._FAILUREREASON
_TESTRESULT.fields_by_name['properties'].message_type = google_dot_protobuf_dot_struct__pb2._STRUCT
_TESTRESULT.fields_by_name['skip_reason'].enum_type = _SKIPREASON
_TESTEXONERATION.fields_by_name['variant'].message_type = go_dot_chromium_dot_org_dot_luci_dot_resultdb_dot_proto_dot_v1_dot_common__pb2._VARIANT
_TESTEXONERATION.fields_by_name['reason'].enum_type = _EXONERATIONREASON
DESCRIPTOR.message_types_by_name['TestResult'] = _TESTRESULT
DESCRIPTOR.message_types_by_name['TestExoneration'] = _TESTEXONERATION
DESCRIPTOR.enum_types_by_name['TestStatus'] = _TESTSTATUS
DESCRIPTOR.enum_types_by_name['SkipReason'] = _SKIPREASON
DESCRIPTOR.enum_types_by_name['ExonerationReason'] = _EXONERATIONREASON
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

TestResult = _reflection.GeneratedProtocolMessageType('TestResult', (_message.Message,), {
  'DESCRIPTOR' : _TESTRESULT,
  '__module__' : 'go.chromium.org.luci.resultdb.proto.v1.test_result_pb2'
  # @@protoc_insertion_point(class_scope:luci.resultdb.v1.TestResult)
  })
_sym_db.RegisterMessage(TestResult)

TestExoneration = _reflection.GeneratedProtocolMessageType('TestExoneration', (_message.Message,), {
  'DESCRIPTOR' : _TESTEXONERATION,
  '__module__' : 'go.chromium.org.luci.resultdb.proto.v1.test_result_pb2'
  # @@protoc_insertion_point(class_scope:luci.resultdb.v1.TestExoneration)
  })
_sym_db.RegisterMessage(TestExoneration)


DESCRIPTOR._options = None
_TESTRESULT.fields_by_name['name']._options = None
_TESTRESULT.fields_by_name['test_id']._options = None
_TESTRESULT.fields_by_name['result_id']._options = None
_TESTRESULT.fields_by_name['variant']._options = None
_TESTRESULT.fields_by_name['expected']._options = None
_TESTRESULT.fields_by_name['status']._options = None
_TESTRESULT.fields_by_name['summary_html']._options = None
_TESTRESULT.fields_by_name['start_time']._options = None
_TESTRESULT.fields_by_name['duration']._options = None
_TESTRESULT.fields_by_name['tags']._options = None
_TESTRESULT.fields_by_name['variant_hash']._options = None
_TESTRESULT.fields_by_name['is_masked']._options = None
_TESTEXONERATION.fields_by_name['name']._options = None
_TESTEXONERATION.fields_by_name['exoneration_id']._options = None
_TESTEXONERATION.fields_by_name['explanation_html']._options = None
_TESTEXONERATION.fields_by_name['variant_hash']._options = None
_TESTEXONERATION.fields_by_name['reason']._options = None
_TESTEXONERATION.fields_by_name['is_masked']._options = None
# @@protoc_insertion_point(module_scope)
