# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: go.chromium.org/luci/cv/api/bigquery/v1/attempt.proto
"""Generated protocol buffer code."""
from google.protobuf.internal import enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google.protobuf import timestamp_pb2 as google_dot_protobuf_dot_timestamp__pb2


DESCRIPTOR = _descriptor.FileDescriptor(
  name='go.chromium.org/luci/cv/api/bigquery/v1/attempt.proto',
  package='bigquery',
  syntax='proto3',
  serialized_options=b'Z0go.chromium.org/luci/cv/api/bigquery/v1;bigquery',
  create_key=_descriptor._internal_create_key,
  serialized_pb=b'\n5go.chromium.org/luci/cv/api/bigquery/v1/attempt.proto\x12\x08\x62igquery\x1a\x1fgoogle/protobuf/timestamp.proto\"\xa0\x03\n\x07\x41ttempt\x12\x0b\n\x03key\x18\x01 \x01(\t\x12\x14\n\x0cluci_project\x18\x02 \x01(\t\x12\x14\n\x0c\x63onfig_group\x18\x0b \x01(\t\x12\x14\n\x0c\x63l_group_key\x18\x03 \x01(\t\x12\x1f\n\x17\x65quivalent_cl_group_key\x18\x04 \x01(\t\x12.\n\nstart_time\x18\x05 \x01(\x0b\x32\x1a.google.protobuf.Timestamp\x12,\n\x08\x65nd_time\x18\x06 \x01(\x0b\x32\x1a.google.protobuf.Timestamp\x12.\n\x0egerrit_changes\x18\x07 \x03(\x0b\x32\x16.bigquery.GerritChange\x12\x1f\n\x06\x62uilds\x18\x08 \x03(\x0b\x32\x0f.bigquery.Build\x12\'\n\x06status\x18\t \x01(\x0e\x32\x17.bigquery.AttemptStatus\x12-\n\tsubstatus\x18\n \x01(\x0e\x32\x1a.bigquery.AttemptSubstatus\x12\x1e\n\x16has_custom_requirement\x18\x0c \x01(\x08\"\x8d\x03\n\x0cGerritChange\x12\x0c\n\x04host\x18\x01 \x01(\t\x12\x0f\n\x07project\x18\x02 \x01(\t\x12\x0e\n\x06\x63hange\x18\x03 \x01(\x03\x12\x10\n\x08patchset\x18\x04 \x01(\x03\x12$\n\x1c\x65\x61rliest_equivalent_patchset\x18\x05 \x01(\x03\x12\x30\n\x0ctrigger_time\x18\x06 \x01(\x0b\x32\x1a.google.protobuf.Timestamp\x12\x1c\n\x04mode\x18\x07 \x01(\x0e\x32\x0e.bigquery.Mode\x12:\n\rsubmit_status\x18\x08 \x01(\x0e\x32#.bigquery.GerritChange.SubmitStatus\x12\x11\n\x05owner\x18\t \x01(\tB\x02\x18\x01\x12\x14\n\x0cis_owner_bot\x18\n \x01(\x08\"a\n\x0cSubmitStatus\x12\x1d\n\x19SUBMIT_STATUS_UNSPECIFIED\x10\x00\x12\x0b\n\x07PENDING\x10\x01\x12\x0b\n\x07UNKNOWN\x10\x02\x12\x0b\n\x07\x46\x41ILURE\x10\x03\x12\x0b\n\x07SUCCESS\x10\x04\"\xab\x01\n\x05\x42uild\x12\n\n\x02id\x18\x01 \x01(\x03\x12\x0c\n\x04host\x18\x02 \x01(\t\x12&\n\x06origin\x18\x03 \x01(\x0e\x32\x16.bigquery.Build.Origin\x12\x10\n\x08\x63ritical\x18\x04 \x01(\x08\"N\n\x06Origin\x12\x16\n\x12ORIGIN_UNSPECIFIED\x10\x00\x12\x10\n\x0cNOT_REUSABLE\x10\x01\x12\x0e\n\nNOT_REUSED\x10\x02\x12\n\n\x06REUSED\x10\x03*=\n\x04Mode\x12\x14\n\x10MODE_UNSPECIFIED\x10\x00\x12\x0b\n\x07\x44RY_RUN\x10\x01\x12\x0c\n\x08\x46ULL_RUN\x10\x02\"\x04\x08\x03\x10\x03*v\n\rAttemptStatus\x12\x1e\n\x1a\x41TTEMPT_STATUS_UNSPECIFIED\x10\x00\x12\x0b\n\x07STARTED\x10\x01\x12\x0b\n\x07SUCCESS\x10\x02\x12\x0b\n\x07\x41\x42ORTED\x10\x03\x12\x0b\n\x07\x46\x41ILURE\x10\x04\x12\x11\n\rINFRA_FAILURE\x10\x05*\xe4\x01\n\x10\x41ttemptSubstatus\x12!\n\x1d\x41TTEMPT_SUBSTATUS_UNSPECIFIED\x10\x00\x12\x10\n\x0cNO_SUBSTATUS\x10\x01\x12\x12\n\x0e\x46\x41ILED_TRYJOBS\x10\x02\x12\x0f\n\x0b\x46\x41ILED_LINT\x10\x03\x12\x0e\n\nUNAPPROVED\x10\x04\x12\x15\n\x11PERMISSION_DENIED\x10\x05\x12\x1a\n\x16UNSATISFIED_DEPENDENCY\x10\x06\x12\x11\n\rMANUAL_CANCEL\x10\x07\x12 \n\x1c\x42UILDBUCKET_MISCONFIGURATION\x10\x08\x42\x32Z0go.chromium.org/luci/cv/api/bigquery/v1;bigqueryb\x06proto3'
  ,
  dependencies=[google_dot_protobuf_dot_timestamp__pb2.DESCRIPTOR,])

_MODE = _descriptor.EnumDescriptor(
  name='Mode',
  full_name='bigquery.Mode',
  filename=None,
  file=DESCRIPTOR,
  create_key=_descriptor._internal_create_key,
  values=[
    _descriptor.EnumValueDescriptor(
      name='MODE_UNSPECIFIED', index=0, number=0,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='DRY_RUN', index=1, number=1,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='FULL_RUN', index=2, number=2,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
  ],
  containing_type=None,
  serialized_options=None,
  serialized_start=1093,
  serialized_end=1154,
)
_sym_db.RegisterEnumDescriptor(_MODE)

Mode = enum_type_wrapper.EnumTypeWrapper(_MODE)
_ATTEMPTSTATUS = _descriptor.EnumDescriptor(
  name='AttemptStatus',
  full_name='bigquery.AttemptStatus',
  filename=None,
  file=DESCRIPTOR,
  create_key=_descriptor._internal_create_key,
  values=[
    _descriptor.EnumValueDescriptor(
      name='ATTEMPT_STATUS_UNSPECIFIED', index=0, number=0,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='STARTED', index=1, number=1,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='SUCCESS', index=2, number=2,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='ABORTED', index=3, number=3,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='FAILURE', index=4, number=4,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='INFRA_FAILURE', index=5, number=5,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
  ],
  containing_type=None,
  serialized_options=None,
  serialized_start=1156,
  serialized_end=1274,
)
_sym_db.RegisterEnumDescriptor(_ATTEMPTSTATUS)

AttemptStatus = enum_type_wrapper.EnumTypeWrapper(_ATTEMPTSTATUS)
_ATTEMPTSUBSTATUS = _descriptor.EnumDescriptor(
  name='AttemptSubstatus',
  full_name='bigquery.AttemptSubstatus',
  filename=None,
  file=DESCRIPTOR,
  create_key=_descriptor._internal_create_key,
  values=[
    _descriptor.EnumValueDescriptor(
      name='ATTEMPT_SUBSTATUS_UNSPECIFIED', index=0, number=0,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='NO_SUBSTATUS', index=1, number=1,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='FAILED_TRYJOBS', index=2, number=2,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='FAILED_LINT', index=3, number=3,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='UNAPPROVED', index=4, number=4,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='PERMISSION_DENIED', index=5, number=5,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='UNSATISFIED_DEPENDENCY', index=6, number=6,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='MANUAL_CANCEL', index=7, number=7,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='BUILDBUCKET_MISCONFIGURATION', index=8, number=8,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
  ],
  containing_type=None,
  serialized_options=None,
  serialized_start=1277,
  serialized_end=1505,
)
_sym_db.RegisterEnumDescriptor(_ATTEMPTSUBSTATUS)

AttemptSubstatus = enum_type_wrapper.EnumTypeWrapper(_ATTEMPTSUBSTATUS)
MODE_UNSPECIFIED = 0
DRY_RUN = 1
FULL_RUN = 2
ATTEMPT_STATUS_UNSPECIFIED = 0
STARTED = 1
SUCCESS = 2
ABORTED = 3
FAILURE = 4
INFRA_FAILURE = 5
ATTEMPT_SUBSTATUS_UNSPECIFIED = 0
NO_SUBSTATUS = 1
FAILED_TRYJOBS = 2
FAILED_LINT = 3
UNAPPROVED = 4
PERMISSION_DENIED = 5
UNSATISFIED_DEPENDENCY = 6
MANUAL_CANCEL = 7
BUILDBUCKET_MISCONFIGURATION = 8


_GERRITCHANGE_SUBMITSTATUS = _descriptor.EnumDescriptor(
  name='SubmitStatus',
  full_name='bigquery.GerritChange.SubmitStatus',
  filename=None,
  file=DESCRIPTOR,
  create_key=_descriptor._internal_create_key,
  values=[
    _descriptor.EnumValueDescriptor(
      name='SUBMIT_STATUS_UNSPECIFIED', index=0, number=0,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='PENDING', index=1, number=1,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='UNKNOWN', index=2, number=2,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='FAILURE', index=3, number=3,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='SUCCESS', index=4, number=4,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
  ],
  containing_type=None,
  serialized_options=None,
  serialized_start=820,
  serialized_end=917,
)
_sym_db.RegisterEnumDescriptor(_GERRITCHANGE_SUBMITSTATUS)

_BUILD_ORIGIN = _descriptor.EnumDescriptor(
  name='Origin',
  full_name='bigquery.Build.Origin',
  filename=None,
  file=DESCRIPTOR,
  create_key=_descriptor._internal_create_key,
  values=[
    _descriptor.EnumValueDescriptor(
      name='ORIGIN_UNSPECIFIED', index=0, number=0,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='NOT_REUSABLE', index=1, number=1,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='NOT_REUSED', index=2, number=2,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='REUSED', index=3, number=3,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
  ],
  containing_type=None,
  serialized_options=None,
  serialized_start=1013,
  serialized_end=1091,
)
_sym_db.RegisterEnumDescriptor(_BUILD_ORIGIN)


_ATTEMPT = _descriptor.Descriptor(
  name='Attempt',
  full_name='bigquery.Attempt',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='key', full_name='bigquery.Attempt.key', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='luci_project', full_name='bigquery.Attempt.luci_project', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='config_group', full_name='bigquery.Attempt.config_group', index=2,
      number=11, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='cl_group_key', full_name='bigquery.Attempt.cl_group_key', index=3,
      number=3, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='equivalent_cl_group_key', full_name='bigquery.Attempt.equivalent_cl_group_key', index=4,
      number=4, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='start_time', full_name='bigquery.Attempt.start_time', index=5,
      number=5, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='end_time', full_name='bigquery.Attempt.end_time', index=6,
      number=6, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='gerrit_changes', full_name='bigquery.Attempt.gerrit_changes', index=7,
      number=7, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='builds', full_name='bigquery.Attempt.builds', index=8,
      number=8, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='status', full_name='bigquery.Attempt.status', index=9,
      number=9, type=14, cpp_type=8, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='substatus', full_name='bigquery.Attempt.substatus', index=10,
      number=10, type=14, cpp_type=8, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='has_custom_requirement', full_name='bigquery.Attempt.has_custom_requirement', index=11,
      number=12, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
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
  serialized_start=101,
  serialized_end=517,
)


_GERRITCHANGE = _descriptor.Descriptor(
  name='GerritChange',
  full_name='bigquery.GerritChange',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='host', full_name='bigquery.GerritChange.host', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='project', full_name='bigquery.GerritChange.project', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='change', full_name='bigquery.GerritChange.change', index=2,
      number=3, type=3, cpp_type=2, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='patchset', full_name='bigquery.GerritChange.patchset', index=3,
      number=4, type=3, cpp_type=2, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='earliest_equivalent_patchset', full_name='bigquery.GerritChange.earliest_equivalent_patchset', index=4,
      number=5, type=3, cpp_type=2, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='trigger_time', full_name='bigquery.GerritChange.trigger_time', index=5,
      number=6, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='mode', full_name='bigquery.GerritChange.mode', index=6,
      number=7, type=14, cpp_type=8, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='submit_status', full_name='bigquery.GerritChange.submit_status', index=7,
      number=8, type=14, cpp_type=8, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='owner', full_name='bigquery.GerritChange.owner', index=8,
      number=9, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=b'\030\001', file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='is_owner_bot', full_name='bigquery.GerritChange.is_owner_bot', index=9,
      number=10, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
    _GERRITCHANGE_SUBMITSTATUS,
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=520,
  serialized_end=917,
)


_BUILD = _descriptor.Descriptor(
  name='Build',
  full_name='bigquery.Build',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='id', full_name='bigquery.Build.id', index=0,
      number=1, type=3, cpp_type=2, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='host', full_name='bigquery.Build.host', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='origin', full_name='bigquery.Build.origin', index=2,
      number=3, type=14, cpp_type=8, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='critical', full_name='bigquery.Build.critical', index=3,
      number=4, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
    _BUILD_ORIGIN,
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=920,
  serialized_end=1091,
)

_ATTEMPT.fields_by_name['start_time'].message_type = google_dot_protobuf_dot_timestamp__pb2._TIMESTAMP
_ATTEMPT.fields_by_name['end_time'].message_type = google_dot_protobuf_dot_timestamp__pb2._TIMESTAMP
_ATTEMPT.fields_by_name['gerrit_changes'].message_type = _GERRITCHANGE
_ATTEMPT.fields_by_name['builds'].message_type = _BUILD
_ATTEMPT.fields_by_name['status'].enum_type = _ATTEMPTSTATUS
_ATTEMPT.fields_by_name['substatus'].enum_type = _ATTEMPTSUBSTATUS
_GERRITCHANGE.fields_by_name['trigger_time'].message_type = google_dot_protobuf_dot_timestamp__pb2._TIMESTAMP
_GERRITCHANGE.fields_by_name['mode'].enum_type = _MODE
_GERRITCHANGE.fields_by_name['submit_status'].enum_type = _GERRITCHANGE_SUBMITSTATUS
_GERRITCHANGE_SUBMITSTATUS.containing_type = _GERRITCHANGE
_BUILD.fields_by_name['origin'].enum_type = _BUILD_ORIGIN
_BUILD_ORIGIN.containing_type = _BUILD
DESCRIPTOR.message_types_by_name['Attempt'] = _ATTEMPT
DESCRIPTOR.message_types_by_name['GerritChange'] = _GERRITCHANGE
DESCRIPTOR.message_types_by_name['Build'] = _BUILD
DESCRIPTOR.enum_types_by_name['Mode'] = _MODE
DESCRIPTOR.enum_types_by_name['AttemptStatus'] = _ATTEMPTSTATUS
DESCRIPTOR.enum_types_by_name['AttemptSubstatus'] = _ATTEMPTSUBSTATUS
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

Attempt = _reflection.GeneratedProtocolMessageType('Attempt', (_message.Message,), {
  'DESCRIPTOR' : _ATTEMPT,
  '__module__' : 'go.chromium.org.luci.cv.api.bigquery.v1.attempt_pb2'
  # @@protoc_insertion_point(class_scope:bigquery.Attempt)
  })
_sym_db.RegisterMessage(Attempt)

GerritChange = _reflection.GeneratedProtocolMessageType('GerritChange', (_message.Message,), {
  'DESCRIPTOR' : _GERRITCHANGE,
  '__module__' : 'go.chromium.org.luci.cv.api.bigquery.v1.attempt_pb2'
  # @@protoc_insertion_point(class_scope:bigquery.GerritChange)
  })
_sym_db.RegisterMessage(GerritChange)

Build = _reflection.GeneratedProtocolMessageType('Build', (_message.Message,), {
  'DESCRIPTOR' : _BUILD,
  '__module__' : 'go.chromium.org.luci.cv.api.bigquery.v1.attempt_pb2'
  # @@protoc_insertion_point(class_scope:bigquery.Build)
  })
_sym_db.RegisterMessage(Build)


DESCRIPTOR._options = None
_GERRITCHANGE.fields_by_name['owner']._options = None
# @@protoc_insertion_point(module_scope)
