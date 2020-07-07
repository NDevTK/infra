# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: api/v3/api_proto/frontend.proto

from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google_proto.google.api import field_behavior_pb2 as google__proto_dot_google_dot_api_dot_field__behavior__pb2
from google_proto.google.api import resource_pb2 as google__proto_dot_google_dot_api_dot_resource__pb2
from api.v3.api_proto import project_objects_pb2 as api_dot_v3_dot_api__proto_dot_project__objects__pb2


DESCRIPTOR = _descriptor.FileDescriptor(
  name='api/v3/api_proto/frontend.proto',
  package='monorail.v3',
  syntax='proto3',
  serialized_options=b'Z\020api/v3/api_proto',
  create_key=_descriptor._internal_create_key,
  serialized_pb=b'\n\x1f\x61pi/v3/api_proto/frontend.proto\x12\x0bmonorail.v3\x1a,google_proto/google/api/field_behavior.proto\x1a&google_proto/google/api/resource.proto\x1a&api/v3/api_proto/project_objects.proto\"P\n\x1fGatherProjectEnvironmentRequest\x12-\n\x06parent\x18\x01 \x01(\tB\x1d\xfa\x41\x17\n\x15\x61pi.crbug.com/Project\xe0\x41\x02\"\x99\x03\n GatherProjectEnvironmentResponse\x12%\n\x07project\x18\x01 \x01(\x0b\x32\x14.monorail.v3.Project\x12\x32\n\x0eproject_config\x18\x02 \x01(\x0b\x32\x1a.monorail.v3.ProjectConfig\x12(\n\x08statuses\x18\x03 \x03(\x0b\x32\x16.monorail.v3.StatusDef\x12\x30\n\x11well_known_labels\x18\x04 \x03(\x0b\x32\x15.monorail.v3.LabelDef\x12-\n\ncomponents\x18\x05 \x03(\x0b\x32\x19.monorail.v3.ComponentDef\x12%\n\x06\x66ields\x18\x06 \x03(\x0b\x32\x15.monorail.v3.FieldDef\x12\x31\n\x0f\x61pproval_fields\x18\x07 \x03(\x0b\x32\x18.monorail.v3.ApprovalDef\x12\x35\n\rsaved_queries\x18\x08 \x03(\x0b\x32\x1e.monorail.v3.ProjectSavedQuery\"O\n&GatherProjectMembershipsForUserRequest\x12%\n\x04user\x18\x01 \x01(\tB\x17\xfa\x41\x14\n\x12\x61pi.crbug.com/User\"b\n\'GatherProjectMembershipsForUserResponse\x12\x37\n\x13project_memberships\x18\x01 \x03(\x0b\x32\x1a.monorail.v3.ProjectMember2\x96\x02\n\x08\x46rontend\x12y\n\x18GatherProjectEnvironment\x12,.monorail.v3.GatherProjectEnvironmentRequest\x1a-.monorail.v3.GatherProjectEnvironmentResponse\"\x00\x12\x8e\x01\n\x1fGatherProjectMembershipsForUser\x12\x33.monorail.v3.GatherProjectMembershipsForUserRequest\x1a\x34.monorail.v3.GatherProjectMembershipsForUserResponse\"\x00\x42\x12Z\x10\x61pi/v3/api_protob\x06proto3'
  ,
  dependencies=[google__proto_dot_google_dot_api_dot_field__behavior__pb2.DESCRIPTOR,google__proto_dot_google_dot_api_dot_resource__pb2.DESCRIPTOR,api_dot_v3_dot_api__proto_dot_project__objects__pb2.DESCRIPTOR,])




_GATHERPROJECTENVIRONMENTREQUEST = _descriptor.Descriptor(
  name='GatherProjectEnvironmentRequest',
  full_name='monorail.v3.GatherProjectEnvironmentRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='parent', full_name='monorail.v3.GatherProjectEnvironmentRequest.parent', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=b'\372A\027\n\025api.crbug.com/Project\340A\002', file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
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
  serialized_start=174,
  serialized_end=254,
)


_GATHERPROJECTENVIRONMENTRESPONSE = _descriptor.Descriptor(
  name='GatherProjectEnvironmentResponse',
  full_name='monorail.v3.GatherProjectEnvironmentResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='project', full_name='monorail.v3.GatherProjectEnvironmentResponse.project', index=0,
      number=1, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='project_config', full_name='monorail.v3.GatherProjectEnvironmentResponse.project_config', index=1,
      number=2, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='statuses', full_name='monorail.v3.GatherProjectEnvironmentResponse.statuses', index=2,
      number=3, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='well_known_labels', full_name='monorail.v3.GatherProjectEnvironmentResponse.well_known_labels', index=3,
      number=4, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='components', full_name='monorail.v3.GatherProjectEnvironmentResponse.components', index=4,
      number=5, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='fields', full_name='monorail.v3.GatherProjectEnvironmentResponse.fields', index=5,
      number=6, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='approval_fields', full_name='monorail.v3.GatherProjectEnvironmentResponse.approval_fields', index=6,
      number=7, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='saved_queries', full_name='monorail.v3.GatherProjectEnvironmentResponse.saved_queries', index=7,
      number=8, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
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
  serialized_start=257,
  serialized_end=666,
)


_GATHERPROJECTMEMBERSHIPSFORUSERREQUEST = _descriptor.Descriptor(
  name='GatherProjectMembershipsForUserRequest',
  full_name='monorail.v3.GatherProjectMembershipsForUserRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='user', full_name='monorail.v3.GatherProjectMembershipsForUserRequest.user', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=b'\372A\024\n\022api.crbug.com/User', file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
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
  serialized_start=668,
  serialized_end=747,
)


_GATHERPROJECTMEMBERSHIPSFORUSERRESPONSE = _descriptor.Descriptor(
  name='GatherProjectMembershipsForUserResponse',
  full_name='monorail.v3.GatherProjectMembershipsForUserResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='project_memberships', full_name='monorail.v3.GatherProjectMembershipsForUserResponse.project_memberships', index=0,
      number=1, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
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
  serialized_start=749,
  serialized_end=847,
)

_GATHERPROJECTENVIRONMENTRESPONSE.fields_by_name['project'].message_type = api_dot_v3_dot_api__proto_dot_project__objects__pb2._PROJECT
_GATHERPROJECTENVIRONMENTRESPONSE.fields_by_name['project_config'].message_type = api_dot_v3_dot_api__proto_dot_project__objects__pb2._PROJECTCONFIG
_GATHERPROJECTENVIRONMENTRESPONSE.fields_by_name['statuses'].message_type = api_dot_v3_dot_api__proto_dot_project__objects__pb2._STATUSDEF
_GATHERPROJECTENVIRONMENTRESPONSE.fields_by_name['well_known_labels'].message_type = api_dot_v3_dot_api__proto_dot_project__objects__pb2._LABELDEF
_GATHERPROJECTENVIRONMENTRESPONSE.fields_by_name['components'].message_type = api_dot_v3_dot_api__proto_dot_project__objects__pb2._COMPONENTDEF
_GATHERPROJECTENVIRONMENTRESPONSE.fields_by_name['fields'].message_type = api_dot_v3_dot_api__proto_dot_project__objects__pb2._FIELDDEF
_GATHERPROJECTENVIRONMENTRESPONSE.fields_by_name['approval_fields'].message_type = api_dot_v3_dot_api__proto_dot_project__objects__pb2._APPROVALDEF
_GATHERPROJECTENVIRONMENTRESPONSE.fields_by_name['saved_queries'].message_type = api_dot_v3_dot_api__proto_dot_project__objects__pb2._PROJECTSAVEDQUERY
_GATHERPROJECTMEMBERSHIPSFORUSERRESPONSE.fields_by_name['project_memberships'].message_type = api_dot_v3_dot_api__proto_dot_project__objects__pb2._PROJECTMEMBER
DESCRIPTOR.message_types_by_name['GatherProjectEnvironmentRequest'] = _GATHERPROJECTENVIRONMENTREQUEST
DESCRIPTOR.message_types_by_name['GatherProjectEnvironmentResponse'] = _GATHERPROJECTENVIRONMENTRESPONSE
DESCRIPTOR.message_types_by_name['GatherProjectMembershipsForUserRequest'] = _GATHERPROJECTMEMBERSHIPSFORUSERREQUEST
DESCRIPTOR.message_types_by_name['GatherProjectMembershipsForUserResponse'] = _GATHERPROJECTMEMBERSHIPSFORUSERRESPONSE
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

GatherProjectEnvironmentRequest = _reflection.GeneratedProtocolMessageType('GatherProjectEnvironmentRequest', (_message.Message,), {
  'DESCRIPTOR' : _GATHERPROJECTENVIRONMENTREQUEST,
  '__module__' : 'api.v3.api_proto.frontend_pb2'
  # @@protoc_insertion_point(class_scope:monorail.v3.GatherProjectEnvironmentRequest)
  })
_sym_db.RegisterMessage(GatherProjectEnvironmentRequest)

GatherProjectEnvironmentResponse = _reflection.GeneratedProtocolMessageType('GatherProjectEnvironmentResponse', (_message.Message,), {
  'DESCRIPTOR' : _GATHERPROJECTENVIRONMENTRESPONSE,
  '__module__' : 'api.v3.api_proto.frontend_pb2'
  # @@protoc_insertion_point(class_scope:monorail.v3.GatherProjectEnvironmentResponse)
  })
_sym_db.RegisterMessage(GatherProjectEnvironmentResponse)

GatherProjectMembershipsForUserRequest = _reflection.GeneratedProtocolMessageType('GatherProjectMembershipsForUserRequest', (_message.Message,), {
  'DESCRIPTOR' : _GATHERPROJECTMEMBERSHIPSFORUSERREQUEST,
  '__module__' : 'api.v3.api_proto.frontend_pb2'
  # @@protoc_insertion_point(class_scope:monorail.v3.GatherProjectMembershipsForUserRequest)
  })
_sym_db.RegisterMessage(GatherProjectMembershipsForUserRequest)

GatherProjectMembershipsForUserResponse = _reflection.GeneratedProtocolMessageType('GatherProjectMembershipsForUserResponse', (_message.Message,), {
  'DESCRIPTOR' : _GATHERPROJECTMEMBERSHIPSFORUSERRESPONSE,
  '__module__' : 'api.v3.api_proto.frontend_pb2'
  # @@protoc_insertion_point(class_scope:monorail.v3.GatherProjectMembershipsForUserResponse)
  })
_sym_db.RegisterMessage(GatherProjectMembershipsForUserResponse)


DESCRIPTOR._options = None
_GATHERPROJECTENVIRONMENTREQUEST.fields_by_name['parent']._options = None
_GATHERPROJECTMEMBERSHIPSFORUSERREQUEST.fields_by_name['user']._options = None

_FRONTEND = _descriptor.ServiceDescriptor(
  name='Frontend',
  full_name='monorail.v3.Frontend',
  file=DESCRIPTOR,
  index=0,
  serialized_options=None,
  create_key=_descriptor._internal_create_key,
  serialized_start=850,
  serialized_end=1128,
  methods=[
  _descriptor.MethodDescriptor(
    name='GatherProjectEnvironment',
    full_name='monorail.v3.Frontend.GatherProjectEnvironment',
    index=0,
    containing_service=None,
    input_type=_GATHERPROJECTENVIRONMENTREQUEST,
    output_type=_GATHERPROJECTENVIRONMENTRESPONSE,
    serialized_options=None,
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='GatherProjectMembershipsForUser',
    full_name='monorail.v3.Frontend.GatherProjectMembershipsForUser',
    index=1,
    containing_service=None,
    input_type=_GATHERPROJECTMEMBERSHIPSFORUSERREQUEST,
    output_type=_GATHERPROJECTMEMBERSHIPSFORUSERRESPONSE,
    serialized_options=None,
    create_key=_descriptor._internal_create_key,
  ),
])
_sym_db.RegisterServiceDescriptor(_FRONTEND)

DESCRIPTOR.services_by_name['Frontend'] = _FRONTEND

# @@protoc_insertion_point(module_scope)
