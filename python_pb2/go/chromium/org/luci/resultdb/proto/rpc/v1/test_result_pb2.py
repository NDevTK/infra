# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: go.chromium.org/luci/resultdb/proto/rpc/v1/test_result.proto

import sys
_b=sys.version_info[0]<3 and (lambda x:x) or (lambda x:x.encode('latin1'))
from google.protobuf.internal import enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google.api import field_behavior_pb2 as google_dot_api_dot_field__behavior__pb2
from google.protobuf import duration_pb2 as google_dot_protobuf_dot_duration__pb2
from google.protobuf import timestamp_pb2 as google_dot_protobuf_dot_timestamp__pb2
from go.chromium.org.luci.resultdb.proto.type import common_pb2 as go_dot_chromium_dot_org_dot_luci_dot_resultdb_dot_proto_dot_type_dot_common__pb2


DESCRIPTOR = _descriptor.FileDescriptor(
  name='go.chromium.org/luci/resultdb/proto/rpc/v1/test_result.proto',
  package='luci.resultdb.rpc.v1',
  syntax='proto3',
  serialized_options=None,
  serialized_pb=_b('\n<go.chromium.org/luci/resultdb/proto/rpc/v1/test_result.proto\x12\x14luci.resultdb.rpc.v1\x1a\x1fgoogle/api/field_behavior.proto\x1a\x1egoogle/protobuf/duration.proto\x1a\x1fgoogle/protobuf/timestamp.proto\x1a\x35go.chromium.org/luci/resultdb/proto/type/common.proto\"\x86\x04\n\nTestResult\x12\x14\n\x04name\x18\x01 \x01(\tB\x06\xe0\x41\x03\xe0\x41\x05\x12\x14\n\x07test_id\x18\x02 \x01(\tB\x03\xe0\x41\x05\x12\x19\n\tresult_id\x18\x03 \x01(\tB\x06\xe0\x41\x05\xe0\x41\x02\x12\x31\n\x07variant\x18\x04 \x01(\x0b\x32\x1b.luci.resultdb.type.VariantB\x03\xe0\x41\x05\x12\x15\n\x08\x65xpected\x18\x05 \x01(\x08\x42\x03\xe0\x41\x05\x12\x35\n\x06status\x18\x06 \x01(\x0e\x32 .luci.resultdb.rpc.v1.TestStatusB\x03\xe0\x41\x05\x12\x19\n\x0csummary_html\x18\x07 \x01(\tB\x03\xe0\x41\x05\x12\x33\n\nstart_time\x18\x08 \x01(\x0b\x32\x1a.google.protobuf.TimestampB\x03\xe0\x41\x05\x12\x30\n\x08\x64uration\x18\t \x01(\x0b\x32\x19.google.protobuf.DurationB\x03\xe0\x41\x05\x12\x31\n\x04tags\x18\n \x03(\x0b\x32\x1e.luci.resultdb.type.StringPairB\x03\xe0\x41\x05\x12<\n\x0finput_artifacts\x18\x0b \x03(\x0b\x32\x1e.luci.resultdb.rpc.v1.ArtifactB\x03\xe0\x41\x05\x12=\n\x10output_artifacts\x18\x0c \x03(\x0b\x32\x1e.luci.resultdb.rpc.v1.ArtifactB\x03\xe0\x41\x05\"\x98\x01\n\x08\x41rtifact\x12\x11\n\x04name\x18\x01 \x01(\tB\x03\xe0\x41\x05\x12\x11\n\tfetch_url\x18\x02 \x01(\t\x12\x38\n\x14\x66\x65tch_url_expiration\x18\x03 \x01(\x0b\x32\x1a.google.protobuf.Timestamp\x12\x19\n\x0c\x63ontent_type\x18\x04 \x01(\tB\x03\xe0\x41\x05\x12\x11\n\x04size\x18\x05 \x01(\x03\x42\x03\xe0\x41\x05\"\xa5\x01\n\x0fTestExoneration\x12\x14\n\x04name\x18\x01 \x01(\tB\x06\xe0\x41\x03\xe0\x41\x05\x12\x0f\n\x07test_id\x18\x02 \x01(\t\x12,\n\x07variant\x18\x03 \x01(\x0b\x32\x1b.luci.resultdb.type.Variant\x12\x1e\n\x0e\x65xoneration_id\x18\x04 \x01(\tB\x06\xe0\x41\x03\xe0\x41\x05\x12\x1d\n\x10\x65xplanation_html\x18\x05 \x01(\tB\x03\xe0\x41\x05*X\n\nTestStatus\x12\x16\n\x12STATUS_UNSPECIFIED\x10\x00\x12\x08\n\x04PASS\x10\x01\x12\x08\n\x04\x46\x41IL\x10\x02\x12\t\n\x05\x43RASH\x10\x03\x12\t\n\x05\x41\x42ORT\x10\x04\x12\x08\n\x04SKIP\x10\x05\x62\x06proto3')
  ,
  dependencies=[google_dot_api_dot_field__behavior__pb2.DESCRIPTOR,google_dot_protobuf_dot_duration__pb2.DESCRIPTOR,google_dot_protobuf_dot_timestamp__pb2.DESCRIPTOR,go_dot_chromium_dot_org_dot_luci_dot_resultdb_dot_proto_dot_type_dot_common__pb2.DESCRIPTOR,])

_TESTSTATUS = _descriptor.EnumDescriptor(
  name='TestStatus',
  full_name='luci.resultdb.rpc.v1.TestStatus',
  filename=None,
  file=DESCRIPTOR,
  values=[
    _descriptor.EnumValueDescriptor(
      name='STATUS_UNSPECIFIED', index=0, number=0,
      serialized_options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='PASS', index=1, number=1,
      serialized_options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='FAIL', index=2, number=2,
      serialized_options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='CRASH', index=3, number=3,
      serialized_options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='ABORT', index=4, number=4,
      serialized_options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='SKIP', index=5, number=5,
      serialized_options=None,
      type=None),
  ],
  containing_type=None,
  serialized_options=None,
  serialized_start=1083,
  serialized_end=1171,
)
_sym_db.RegisterEnumDescriptor(_TESTSTATUS)

TestStatus = enum_type_wrapper.EnumTypeWrapper(_TESTSTATUS)
STATUS_UNSPECIFIED = 0
PASS = 1
FAIL = 2
CRASH = 3
ABORT = 4
SKIP = 5



_TESTRESULT = _descriptor.Descriptor(
  name='TestResult',
  full_name='luci.resultdb.rpc.v1.TestResult',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='name', full_name='luci.resultdb.rpc.v1.TestResult.name', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=_b('\340A\003\340A\005'), file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='test_id', full_name='luci.resultdb.rpc.v1.TestResult.test_id', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=_b('\340A\005'), file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='result_id', full_name='luci.resultdb.rpc.v1.TestResult.result_id', index=2,
      number=3, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=_b('\340A\005\340A\002'), file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='variant', full_name='luci.resultdb.rpc.v1.TestResult.variant', index=3,
      number=4, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=_b('\340A\005'), file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='expected', full_name='luci.resultdb.rpc.v1.TestResult.expected', index=4,
      number=5, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=_b('\340A\005'), file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='status', full_name='luci.resultdb.rpc.v1.TestResult.status', index=5,
      number=6, type=14, cpp_type=8, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=_b('\340A\005'), file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='summary_html', full_name='luci.resultdb.rpc.v1.TestResult.summary_html', index=6,
      number=7, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=_b('\340A\005'), file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='start_time', full_name='luci.resultdb.rpc.v1.TestResult.start_time', index=7,
      number=8, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=_b('\340A\005'), file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='duration', full_name='luci.resultdb.rpc.v1.TestResult.duration', index=8,
      number=9, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=_b('\340A\005'), file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='tags', full_name='luci.resultdb.rpc.v1.TestResult.tags', index=9,
      number=10, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=_b('\340A\005'), file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='input_artifacts', full_name='luci.resultdb.rpc.v1.TestResult.input_artifacts', index=10,
      number=11, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=_b('\340A\005'), file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='output_artifacts', full_name='luci.resultdb.rpc.v1.TestResult.output_artifacts', index=11,
      number=12, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=_b('\340A\005'), file=DESCRIPTOR),
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
  serialized_start=240,
  serialized_end=758,
)


_ARTIFACT = _descriptor.Descriptor(
  name='Artifact',
  full_name='luci.resultdb.rpc.v1.Artifact',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='name', full_name='luci.resultdb.rpc.v1.Artifact.name', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=_b('\340A\005'), file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='fetch_url', full_name='luci.resultdb.rpc.v1.Artifact.fetch_url', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='fetch_url_expiration', full_name='luci.resultdb.rpc.v1.Artifact.fetch_url_expiration', index=2,
      number=3, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='content_type', full_name='luci.resultdb.rpc.v1.Artifact.content_type', index=3,
      number=4, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=_b('\340A\005'), file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='size', full_name='luci.resultdb.rpc.v1.Artifact.size', index=4,
      number=5, type=3, cpp_type=2, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=_b('\340A\005'), file=DESCRIPTOR),
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
  serialized_start=761,
  serialized_end=913,
)


_TESTEXONERATION = _descriptor.Descriptor(
  name='TestExoneration',
  full_name='luci.resultdb.rpc.v1.TestExoneration',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='name', full_name='luci.resultdb.rpc.v1.TestExoneration.name', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=_b('\340A\003\340A\005'), file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='test_id', full_name='luci.resultdb.rpc.v1.TestExoneration.test_id', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='variant', full_name='luci.resultdb.rpc.v1.TestExoneration.variant', index=2,
      number=3, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='exoneration_id', full_name='luci.resultdb.rpc.v1.TestExoneration.exoneration_id', index=3,
      number=4, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=_b('\340A\003\340A\005'), file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='explanation_html', full_name='luci.resultdb.rpc.v1.TestExoneration.explanation_html', index=4,
      number=5, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=_b('\340A\005'), file=DESCRIPTOR),
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
  serialized_start=916,
  serialized_end=1081,
)

_TESTRESULT.fields_by_name['variant'].message_type = go_dot_chromium_dot_org_dot_luci_dot_resultdb_dot_proto_dot_type_dot_common__pb2._VARIANT
_TESTRESULT.fields_by_name['status'].enum_type = _TESTSTATUS
_TESTRESULT.fields_by_name['start_time'].message_type = google_dot_protobuf_dot_timestamp__pb2._TIMESTAMP
_TESTRESULT.fields_by_name['duration'].message_type = google_dot_protobuf_dot_duration__pb2._DURATION
_TESTRESULT.fields_by_name['tags'].message_type = go_dot_chromium_dot_org_dot_luci_dot_resultdb_dot_proto_dot_type_dot_common__pb2._STRINGPAIR
_TESTRESULT.fields_by_name['input_artifacts'].message_type = _ARTIFACT
_TESTRESULT.fields_by_name['output_artifacts'].message_type = _ARTIFACT
_ARTIFACT.fields_by_name['fetch_url_expiration'].message_type = google_dot_protobuf_dot_timestamp__pb2._TIMESTAMP
_TESTEXONERATION.fields_by_name['variant'].message_type = go_dot_chromium_dot_org_dot_luci_dot_resultdb_dot_proto_dot_type_dot_common__pb2._VARIANT
DESCRIPTOR.message_types_by_name['TestResult'] = _TESTRESULT
DESCRIPTOR.message_types_by_name['Artifact'] = _ARTIFACT
DESCRIPTOR.message_types_by_name['TestExoneration'] = _TESTEXONERATION
DESCRIPTOR.enum_types_by_name['TestStatus'] = _TESTSTATUS
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

TestResult = _reflection.GeneratedProtocolMessageType('TestResult', (_message.Message,), dict(
  DESCRIPTOR = _TESTRESULT,
  __module__ = 'go.chromium.org.luci.resultdb.proto.rpc.v1.test_result_pb2'
  # @@protoc_insertion_point(class_scope:luci.resultdb.rpc.v1.TestResult)
  ))
_sym_db.RegisterMessage(TestResult)

Artifact = _reflection.GeneratedProtocolMessageType('Artifact', (_message.Message,), dict(
  DESCRIPTOR = _ARTIFACT,
  __module__ = 'go.chromium.org.luci.resultdb.proto.rpc.v1.test_result_pb2'
  # @@protoc_insertion_point(class_scope:luci.resultdb.rpc.v1.Artifact)
  ))
_sym_db.RegisterMessage(Artifact)

TestExoneration = _reflection.GeneratedProtocolMessageType('TestExoneration', (_message.Message,), dict(
  DESCRIPTOR = _TESTEXONERATION,
  __module__ = 'go.chromium.org.luci.resultdb.proto.rpc.v1.test_result_pb2'
  # @@protoc_insertion_point(class_scope:luci.resultdb.rpc.v1.TestExoneration)
  ))
_sym_db.RegisterMessage(TestExoneration)


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
_TESTRESULT.fields_by_name['input_artifacts']._options = None
_TESTRESULT.fields_by_name['output_artifacts']._options = None
_ARTIFACT.fields_by_name['name']._options = None
_ARTIFACT.fields_by_name['content_type']._options = None
_ARTIFACT.fields_by_name['size']._options = None
_TESTEXONERATION.fields_by_name['name']._options = None
_TESTEXONERATION.fields_by_name['exoneration_id']._options = None
_TESTEXONERATION.fields_by_name['explanation_html']._options = None
# @@protoc_insertion_point(module_scope)
