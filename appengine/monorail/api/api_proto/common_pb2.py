# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: api/api_proto/common.proto

import sys
_b=sys.version_info[0]<3 and (lambda x:x) or (lambda x:x.encode('latin1'))
from google.protobuf.internal import enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
from google.protobuf import descriptor_pb2
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor.FileDescriptor(
  name='api/api_proto/common.proto',
  package='monorail',
  syntax='proto3',
  serialized_pb=_b('\n\x1a\x61pi/api_proto/common.proto\x12\x08monorail\"W\n\x0cRequestTrace\x12\r\n\x05token\x18\x01 \x01(\t\x12\x0e\n\x06reason\x18\x02 \x01(\t\x12\x12\n\nrequest_id\x18\x03 \x01(\t\x12\x14\n\x0ctest_account\x18\x04 \x01(\t\"0\n\x0c\x43omponentRef\x12\x0c\n\x04path\x18\x01 \x01(\t\x12\x12\n\nis_derived\x18\x02 \x01(\x08\"j\n\x08\x46ieldRef\x12\x10\n\x08\x66ield_id\x18\x01 \x01(\x04\x12\x12\n\nfield_name\x18\x02 \x01(\t\x12!\n\x04type\x18\x03 \x01(\x0e\x32\x13.monorail.FieldType\x12\x15\n\rapproval_name\x18\x04 \x01(\t\"-\n\x08LabelRef\x12\r\n\x05label\x18\x01 \x01(\t\x12\x12\n\nis_derived\x18\x02 \x01(\x08\"C\n\tStatusRef\x12\x0e\n\x06status\x18\x01 \x01(\t\x12\x12\n\nmeans_open\x18\x02 \x01(\x08\x12\x12\n\nis_derived\x18\x03 \x01(\x08\"2\n\x08IssueRef\x12\x14\n\x0cproject_name\x18\x01 \x01(\t\x12\x10\n\x08local_id\x18\x02 \x01(\r\"D\n\x07UserRef\x12\x0f\n\x07user_id\x18\x01 \x01(\x04\x12\x14\n\x0c\x64isplay_name\x18\x02 \x01(\t\x12\x12\n\nis_derived\x18\x03 \x01(\x08\"P\n\nHotlistRef\x12\x12\n\nhotlist_id\x18\x01 \x01(\x04\x12\x0c\n\x04name\x18\x02 \x01(\t\x12 \n\x05owner\x18\x03 \x01(\x0b\x32\x11.monorail.UserRef*\x91\x01\n\tFieldType\x12\x0b\n\x07NO_TYPE\x10\x00\x12\r\n\tENUM_TYPE\x10\x01\x12\x0c\n\x08INT_TYPE\x10\x02\x12\x0c\n\x08STR_TYPE\x10\x03\x12\r\n\tUSER_TYPE\x10\x04\x12\r\n\tDATE_TYPE\x10\x05\x12\r\n\tBOOL_TYPE\x10\x06\x12\x0c\n\x08URL_TYPE\x10\x07\x12\x11\n\rAPPROVAL_TYPE\x10\x08\x62\x06proto3')
)
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

_FIELDTYPE = _descriptor.EnumDescriptor(
  name='FieldType',
  full_name='monorail.FieldType',
  filename=None,
  file=DESCRIPTOR,
  values=[
    _descriptor.EnumValueDescriptor(
      name='NO_TYPE', index=0, number=0,
      options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='ENUM_TYPE', index=1, number=1,
      options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='INT_TYPE', index=2, number=2,
      options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='STR_TYPE', index=3, number=3,
      options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='USER_TYPE', index=4, number=4,
      options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='DATE_TYPE', index=5, number=5,
      options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='BOOL_TYPE', index=6, number=6,
      options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='URL_TYPE', index=7, number=7,
      options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='APPROVAL_TYPE', index=8, number=8,
      options=None,
      type=None),
  ],
  containing_type=None,
  options=None,
  serialized_start=608,
  serialized_end=753,
)
_sym_db.RegisterEnumDescriptor(_FIELDTYPE)

FieldType = enum_type_wrapper.EnumTypeWrapper(_FIELDTYPE)
NO_TYPE = 0
ENUM_TYPE = 1
INT_TYPE = 2
STR_TYPE = 3
USER_TYPE = 4
DATE_TYPE = 5
BOOL_TYPE = 6
URL_TYPE = 7
APPROVAL_TYPE = 8



_REQUESTTRACE = _descriptor.Descriptor(
  name='RequestTrace',
  full_name='monorail.RequestTrace',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='token', full_name='monorail.RequestTrace.token', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='reason', full_name='monorail.RequestTrace.reason', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='request_id', full_name='monorail.RequestTrace.request_id', index=2,
      number=3, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='test_account', full_name='monorail.RequestTrace.test_account', index=3,
      number=4, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=40,
  serialized_end=127,
)


_COMPONENTREF = _descriptor.Descriptor(
  name='ComponentRef',
  full_name='monorail.ComponentRef',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='path', full_name='monorail.ComponentRef.path', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='is_derived', full_name='monorail.ComponentRef.is_derived', index=1,
      number=2, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=129,
  serialized_end=177,
)


_FIELDREF = _descriptor.Descriptor(
  name='FieldRef',
  full_name='monorail.FieldRef',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='field_id', full_name='monorail.FieldRef.field_id', index=0,
      number=1, type=4, cpp_type=4, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='field_name', full_name='monorail.FieldRef.field_name', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='type', full_name='monorail.FieldRef.type', index=2,
      number=3, type=14, cpp_type=8, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='approval_name', full_name='monorail.FieldRef.approval_name', index=3,
      number=4, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=179,
  serialized_end=285,
)


_LABELREF = _descriptor.Descriptor(
  name='LabelRef',
  full_name='monorail.LabelRef',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='label', full_name='monorail.LabelRef.label', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='is_derived', full_name='monorail.LabelRef.is_derived', index=1,
      number=2, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=287,
  serialized_end=332,
)


_STATUSREF = _descriptor.Descriptor(
  name='StatusRef',
  full_name='monorail.StatusRef',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='status', full_name='monorail.StatusRef.status', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='means_open', full_name='monorail.StatusRef.means_open', index=1,
      number=2, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='is_derived', full_name='monorail.StatusRef.is_derived', index=2,
      number=3, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=334,
  serialized_end=401,
)


_ISSUEREF = _descriptor.Descriptor(
  name='IssueRef',
  full_name='monorail.IssueRef',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='project_name', full_name='monorail.IssueRef.project_name', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='local_id', full_name='monorail.IssueRef.local_id', index=1,
      number=2, type=13, cpp_type=3, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=403,
  serialized_end=453,
)


_USERREF = _descriptor.Descriptor(
  name='UserRef',
  full_name='monorail.UserRef',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='user_id', full_name='monorail.UserRef.user_id', index=0,
      number=1, type=4, cpp_type=4, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='display_name', full_name='monorail.UserRef.display_name', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='is_derived', full_name='monorail.UserRef.is_derived', index=2,
      number=3, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=455,
  serialized_end=523,
)


_HOTLISTREF = _descriptor.Descriptor(
  name='HotlistRef',
  full_name='monorail.HotlistRef',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='hotlist_id', full_name='monorail.HotlistRef.hotlist_id', index=0,
      number=1, type=4, cpp_type=4, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='name', full_name='monorail.HotlistRef.name', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='owner', full_name='monorail.HotlistRef.owner', index=2,
      number=3, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=525,
  serialized_end=605,
)

_FIELDREF.fields_by_name['type'].enum_type = _FIELDTYPE
_HOTLISTREF.fields_by_name['owner'].message_type = _USERREF
DESCRIPTOR.message_types_by_name['RequestTrace'] = _REQUESTTRACE
DESCRIPTOR.message_types_by_name['ComponentRef'] = _COMPONENTREF
DESCRIPTOR.message_types_by_name['FieldRef'] = _FIELDREF
DESCRIPTOR.message_types_by_name['LabelRef'] = _LABELREF
DESCRIPTOR.message_types_by_name['StatusRef'] = _STATUSREF
DESCRIPTOR.message_types_by_name['IssueRef'] = _ISSUEREF
DESCRIPTOR.message_types_by_name['UserRef'] = _USERREF
DESCRIPTOR.message_types_by_name['HotlistRef'] = _HOTLISTREF
DESCRIPTOR.enum_types_by_name['FieldType'] = _FIELDTYPE

RequestTrace = _reflection.GeneratedProtocolMessageType('RequestTrace', (_message.Message,), dict(
  DESCRIPTOR = _REQUESTTRACE,
  __module__ = 'api.api_proto.common_pb2'
  # @@protoc_insertion_point(class_scope:monorail.RequestTrace)
  ))
_sym_db.RegisterMessage(RequestTrace)

ComponentRef = _reflection.GeneratedProtocolMessageType('ComponentRef', (_message.Message,), dict(
  DESCRIPTOR = _COMPONENTREF,
  __module__ = 'api.api_proto.common_pb2'
  # @@protoc_insertion_point(class_scope:monorail.ComponentRef)
  ))
_sym_db.RegisterMessage(ComponentRef)

FieldRef = _reflection.GeneratedProtocolMessageType('FieldRef', (_message.Message,), dict(
  DESCRIPTOR = _FIELDREF,
  __module__ = 'api.api_proto.common_pb2'
  # @@protoc_insertion_point(class_scope:monorail.FieldRef)
  ))
_sym_db.RegisterMessage(FieldRef)

LabelRef = _reflection.GeneratedProtocolMessageType('LabelRef', (_message.Message,), dict(
  DESCRIPTOR = _LABELREF,
  __module__ = 'api.api_proto.common_pb2'
  # @@protoc_insertion_point(class_scope:monorail.LabelRef)
  ))
_sym_db.RegisterMessage(LabelRef)

StatusRef = _reflection.GeneratedProtocolMessageType('StatusRef', (_message.Message,), dict(
  DESCRIPTOR = _STATUSREF,
  __module__ = 'api.api_proto.common_pb2'
  # @@protoc_insertion_point(class_scope:monorail.StatusRef)
  ))
_sym_db.RegisterMessage(StatusRef)

IssueRef = _reflection.GeneratedProtocolMessageType('IssueRef', (_message.Message,), dict(
  DESCRIPTOR = _ISSUEREF,
  __module__ = 'api.api_proto.common_pb2'
  # @@protoc_insertion_point(class_scope:monorail.IssueRef)
  ))
_sym_db.RegisterMessage(IssueRef)

UserRef = _reflection.GeneratedProtocolMessageType('UserRef', (_message.Message,), dict(
  DESCRIPTOR = _USERREF,
  __module__ = 'api.api_proto.common_pb2'
  # @@protoc_insertion_point(class_scope:monorail.UserRef)
  ))
_sym_db.RegisterMessage(UserRef)

HotlistRef = _reflection.GeneratedProtocolMessageType('HotlistRef', (_message.Message,), dict(
  DESCRIPTOR = _HOTLISTREF,
  __module__ = 'api.api_proto.common_pb2'
  # @@protoc_insertion_point(class_scope:monorail.HotlistRef)
  ))
_sym_db.RegisterMessage(HotlistRef)


# @@protoc_insertion_point(module_scope)
