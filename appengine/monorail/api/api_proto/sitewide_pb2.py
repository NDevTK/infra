# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: api/api_proto/sitewide.proto

import sys
_b=sys.version_info[0]<3 and (lambda x:x) or (lambda x:x.encode('latin1'))
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor.FileDescriptor(
  name='api/api_proto/sitewide.proto',
  package='monorail',
  syntax='proto3',
  serialized_options=None,
  serialized_pb=_b('\n\x1c\x61pi/api_proto/sitewide.proto\x12\x08monorail\"8\n\x13RefreshTokenRequest\x12\r\n\x05token\x18\x02 \x01(\t\x12\x12\n\ntoken_path\x18\x03 \x01(\t\"@\n\x14RefreshTokenResponse\x12\r\n\x05token\x18\x01 \x01(\t\x12\x19\n\x11token_expires_sec\x18\x02 \x01(\r\"\x18\n\x16GetServerStatusRequest\"Y\n\x17GetServerStatusResponse\x12\x16\n\x0e\x62\x61nner_message\x18\x01 \x01(\t\x12\x13\n\x0b\x62\x61nner_time\x18\x02 \x01(\x07\x12\x11\n\tread_only\x18\x03 \x01(\x08\x32\xb5\x01\n\x08Sitewide\x12O\n\x0cRefreshToken\x12\x1d.monorail.RefreshTokenRequest\x1a\x1e.monorail.RefreshTokenResponse\"\x00\x12X\n\x0fGetServerStatus\x12 .monorail.GetServerStatusRequest\x1a!.monorail.GetServerStatusResponse\"\x00\x62\x06proto3')
)




_REFRESHTOKENREQUEST = _descriptor.Descriptor(
  name='RefreshTokenRequest',
  full_name='monorail.RefreshTokenRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='token', full_name='monorail.RefreshTokenRequest.token', index=0,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='token_path', full_name='monorail.RefreshTokenRequest.token_path', index=1,
      number=3, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
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
  serialized_start=42,
  serialized_end=98,
)


_REFRESHTOKENRESPONSE = _descriptor.Descriptor(
  name='RefreshTokenResponse',
  full_name='monorail.RefreshTokenResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='token', full_name='monorail.RefreshTokenResponse.token', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='token_expires_sec', full_name='monorail.RefreshTokenResponse.token_expires_sec', index=1,
      number=2, type=13, cpp_type=3, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
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
  serialized_start=100,
  serialized_end=164,
)


_GETSERVERSTATUSREQUEST = _descriptor.Descriptor(
  name='GetServerStatusRequest',
  full_name='monorail.GetServerStatusRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
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
  serialized_start=166,
  serialized_end=190,
)


_GETSERVERSTATUSRESPONSE = _descriptor.Descriptor(
  name='GetServerStatusResponse',
  full_name='monorail.GetServerStatusResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='banner_message', full_name='monorail.GetServerStatusResponse.banner_message', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='banner_time', full_name='monorail.GetServerStatusResponse.banner_time', index=1,
      number=2, type=7, cpp_type=3, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='read_only', full_name='monorail.GetServerStatusResponse.read_only', index=2,
      number=3, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
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
  serialized_start=192,
  serialized_end=281,
)

DESCRIPTOR.message_types_by_name['RefreshTokenRequest'] = _REFRESHTOKENREQUEST
DESCRIPTOR.message_types_by_name['RefreshTokenResponse'] = _REFRESHTOKENRESPONSE
DESCRIPTOR.message_types_by_name['GetServerStatusRequest'] = _GETSERVERSTATUSREQUEST
DESCRIPTOR.message_types_by_name['GetServerStatusResponse'] = _GETSERVERSTATUSRESPONSE
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

RefreshTokenRequest = _reflection.GeneratedProtocolMessageType('RefreshTokenRequest', (_message.Message,), dict(
  DESCRIPTOR = _REFRESHTOKENREQUEST,
  __module__ = 'api.api_proto.sitewide_pb2'
  # @@protoc_insertion_point(class_scope:monorail.RefreshTokenRequest)
  ))
_sym_db.RegisterMessage(RefreshTokenRequest)

RefreshTokenResponse = _reflection.GeneratedProtocolMessageType('RefreshTokenResponse', (_message.Message,), dict(
  DESCRIPTOR = _REFRESHTOKENRESPONSE,
  __module__ = 'api.api_proto.sitewide_pb2'
  # @@protoc_insertion_point(class_scope:monorail.RefreshTokenResponse)
  ))
_sym_db.RegisterMessage(RefreshTokenResponse)

GetServerStatusRequest = _reflection.GeneratedProtocolMessageType('GetServerStatusRequest', (_message.Message,), dict(
  DESCRIPTOR = _GETSERVERSTATUSREQUEST,
  __module__ = 'api.api_proto.sitewide_pb2'
  # @@protoc_insertion_point(class_scope:monorail.GetServerStatusRequest)
  ))
_sym_db.RegisterMessage(GetServerStatusRequest)

GetServerStatusResponse = _reflection.GeneratedProtocolMessageType('GetServerStatusResponse', (_message.Message,), dict(
  DESCRIPTOR = _GETSERVERSTATUSRESPONSE,
  __module__ = 'api.api_proto.sitewide_pb2'
  # @@protoc_insertion_point(class_scope:monorail.GetServerStatusResponse)
  ))
_sym_db.RegisterMessage(GetServerStatusResponse)



_SITEWIDE = _descriptor.ServiceDescriptor(
  name='Sitewide',
  full_name='monorail.Sitewide',
  file=DESCRIPTOR,
  index=0,
  serialized_options=None,
  serialized_start=284,
  serialized_end=465,
  methods=[
  _descriptor.MethodDescriptor(
    name='RefreshToken',
    full_name='monorail.Sitewide.RefreshToken',
    index=0,
    containing_service=None,
    input_type=_REFRESHTOKENREQUEST,
    output_type=_REFRESHTOKENRESPONSE,
    serialized_options=None,
  ),
  _descriptor.MethodDescriptor(
    name='GetServerStatus',
    full_name='monorail.Sitewide.GetServerStatus',
    index=1,
    containing_service=None,
    input_type=_GETSERVERSTATUSREQUEST,
    output_type=_GETSERVERSTATUSRESPONSE,
    serialized_options=None,
  ),
])
_sym_db.RegisterServiceDescriptor(_SITEWIDE)

DESCRIPTOR.services_by_name['Sitewide'] = _SITEWIDE

# @@protoc_insertion_point(module_scope)
