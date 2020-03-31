# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: api/v1/api_proto/projects.proto

import sys
_b=sys.version_info[0]<3 and (lambda x:x) or (lambda x:x.encode('latin1'))
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google_proto.google.api import field_behavior_pb2 as google__proto_dot_google_dot_api_dot_field__behavior__pb2
from google_proto.google.api import resource_pb2 as google__proto_dot_google_dot_api_dot_resource__pb2
from google_proto.google.api import annotations_pb2 as google__proto_dot_google_dot_api_dot_annotations__pb2
from api.v1.api_proto import project_objects_pb2 as api_dot_v1_dot_api__proto_dot_project__objects__pb2


DESCRIPTOR = _descriptor.FileDescriptor(
  name='api/v1/api_proto/projects.proto',
  package='monorail.v1',
  syntax='proto3',
  serialized_options=None,
  serialized_pb=_b('\n\x1f\x61pi/v1/api_proto/projects.proto\x12\x0bmonorail.v1\x1a,google_proto/google/api/field_behavior.proto\x1a&google_proto/google/api/resource.proto\x1a)google_proto/google/api/annotations.proto\x1a&api/v1/api_proto/project_objects.proto\"q\n\x19ListIssueTemplatesRequest\x12-\n\x06parent\x18\x01 \x01(\tB\x1d\xfa\x41\x17\n\x15\x61pi.crbug.com/Project\xe0\x41\x02\x12\x11\n\tpage_size\x18\x02 \x01(\x05\x12\x12\n\npage_token\x18\x03 \x01(\t\"d\n\x1aListIssueTemplatesResponse\x12-\n\ttemplates\x18\x01 \x03(\x0b\x32\x1a.monorail.v1.IssueTemplate\x12\x17\n\x0fnext_page_token\x18\x02 \x01(\t\"O\n\x1e\x46\x65tchProjectEnvironmentRequest\x12-\n\x06parent\x18\x01 \x01(\tB\x1d\xfa\x41\x17\n\x15\x61pi.crbug.com/Project\xe0\x41\x02\"\xea\x05\n\x1f\x46\x65tchProjectEnvironmentResponse\x12+\n\x0bstatus_defs\x18\x01 \x03(\x0b\x32\x16.monorail.v1.StatusDef\x12)\n\nlabel_defs\x18\x02 \x03(\x0b\x32\x15.monorail.v1.LabelDef\x12\x31\n\x0e\x63omponent_defs\x18\x03 \x03(\x0b\x32\x19.monorail.v1.ComponentDef\x12)\n\nfield_defs\x18\x04 \x03(\x0b\x32\x15.monorail.v1.FieldDef\x12/\n\rapproval_defs\x18\x05 \x03(\x0b\x32\x18.monorail.v1.ApprovalDef\x12\x34\n\x14statuses_offer_merge\x18\x06 \x03(\x0b\x32\x16.monorail.v1.StatusDef\x12 \n\x18\x65xclusive_label_prefixes\x18\x07 \x03(\t\x12\x15\n\rdefault_query\x18\t \x01(\t\x12\x14\n\x0c\x64\x65\x66\x61ult_sort\x18\n \x01(\t\x12\x18\n\x10\x64\x65\x66\x61ult_col_spec\x18\x0b \x01(\t\x12\x16\n\x0e\x64\x65\x66\x61ult_x_attr\x18\x0c \x01(\t\x12\x16\n\x0e\x64\x65\x66\x61ult_y_attr\x18\r \x01(\t\x12\x1d\n\x15project_thumbnail_url\x18\x0e \x01(\t\x12\x1b\n\x13revision_url_format\x18\x0f \x01(\t\x12\x1e\n\x16\x63ustom_issue_entry_url\x18\x10 \x01(\t\x12@\n\x1c\x64\x65\x66\x61ult_template_for_members\x18\x11 \x01(\x0b\x32\x1a.monorail.v1.IssueTemplate\x12\x44\n default_template_for_non_members\x18\x12 \x01(\x0b\x32\x1a.monorail.v1.IssueTemplate\x12\x14\n\x0cproject_name\x18\x13 \x01(\t\x12\x17\n\x0fproject_summary\x18\x14 \x01(\t2\xe2\x02\n\x08Projects\x12\x9f\x01\n\x12ListIssueTemplates\x12&.monorail.v1.ListIssueTemplatesRequest\x1a\'.monorail.v1.ListIssueTemplatesResponse\"8\x82\xd3\xe4\x93\x02\x32\"-/prpc/monorail.v1.Projects/ListIssueTemplates:\x01*\x12\xb3\x01\n\x17\x46\x65tchProjectEnvironment\x12+.monorail.v1.FetchProjectEnvironmentRequest\x1a,.monorail.v1.FetchProjectEnvironmentResponse\"=\x82\xd3\xe4\x93\x02\x37\"2/prpc/monorail.v1.Projects/FetchProjectEnvironment:\x01*b\x06proto3')
  ,
  dependencies=[google__proto_dot_google_dot_api_dot_field__behavior__pb2.DESCRIPTOR,google__proto_dot_google_dot_api_dot_resource__pb2.DESCRIPTOR,google__proto_dot_google_dot_api_dot_annotations__pb2.DESCRIPTOR,api_dot_v1_dot_api__proto_dot_project__objects__pb2.DESCRIPTOR,])




_LISTISSUETEMPLATESREQUEST = _descriptor.Descriptor(
  name='ListIssueTemplatesRequest',
  full_name='monorail.v1.ListIssueTemplatesRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='parent', full_name='monorail.v1.ListIssueTemplatesRequest.parent', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=_b('\372A\027\n\025api.crbug.com/Project\340A\002'), file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='page_size', full_name='monorail.v1.ListIssueTemplatesRequest.page_size', index=1,
      number=2, type=5, cpp_type=1, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='page_token', full_name='monorail.v1.ListIssueTemplatesRequest.page_token', index=2,
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
  serialized_start=217,
  serialized_end=330,
)


_LISTISSUETEMPLATESRESPONSE = _descriptor.Descriptor(
  name='ListIssueTemplatesResponse',
  full_name='monorail.v1.ListIssueTemplatesResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='templates', full_name='monorail.v1.ListIssueTemplatesResponse.templates', index=0,
      number=1, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='next_page_token', full_name='monorail.v1.ListIssueTemplatesResponse.next_page_token', index=1,
      number=2, type=9, cpp_type=9, label=1,
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
  serialized_start=332,
  serialized_end=432,
)


_FETCHPROJECTENVIRONMENTREQUEST = _descriptor.Descriptor(
  name='FetchProjectEnvironmentRequest',
  full_name='monorail.v1.FetchProjectEnvironmentRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='parent', full_name='monorail.v1.FetchProjectEnvironmentRequest.parent', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=_b('\372A\027\n\025api.crbug.com/Project\340A\002'), file=DESCRIPTOR),
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
  serialized_start=434,
  serialized_end=513,
)


_FETCHPROJECTENVIRONMENTRESPONSE = _descriptor.Descriptor(
  name='FetchProjectEnvironmentResponse',
  full_name='monorail.v1.FetchProjectEnvironmentResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='status_defs', full_name='monorail.v1.FetchProjectEnvironmentResponse.status_defs', index=0,
      number=1, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='label_defs', full_name='monorail.v1.FetchProjectEnvironmentResponse.label_defs', index=1,
      number=2, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='component_defs', full_name='monorail.v1.FetchProjectEnvironmentResponse.component_defs', index=2,
      number=3, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='field_defs', full_name='monorail.v1.FetchProjectEnvironmentResponse.field_defs', index=3,
      number=4, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='approval_defs', full_name='monorail.v1.FetchProjectEnvironmentResponse.approval_defs', index=4,
      number=5, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='statuses_offer_merge', full_name='monorail.v1.FetchProjectEnvironmentResponse.statuses_offer_merge', index=5,
      number=6, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='exclusive_label_prefixes', full_name='monorail.v1.FetchProjectEnvironmentResponse.exclusive_label_prefixes', index=6,
      number=7, type=9, cpp_type=9, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='default_query', full_name='monorail.v1.FetchProjectEnvironmentResponse.default_query', index=7,
      number=9, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='default_sort', full_name='monorail.v1.FetchProjectEnvironmentResponse.default_sort', index=8,
      number=10, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='default_col_spec', full_name='monorail.v1.FetchProjectEnvironmentResponse.default_col_spec', index=9,
      number=11, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='default_x_attr', full_name='monorail.v1.FetchProjectEnvironmentResponse.default_x_attr', index=10,
      number=12, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='default_y_attr', full_name='monorail.v1.FetchProjectEnvironmentResponse.default_y_attr', index=11,
      number=13, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='project_thumbnail_url', full_name='monorail.v1.FetchProjectEnvironmentResponse.project_thumbnail_url', index=12,
      number=14, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='revision_url_format', full_name='monorail.v1.FetchProjectEnvironmentResponse.revision_url_format', index=13,
      number=15, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='custom_issue_entry_url', full_name='monorail.v1.FetchProjectEnvironmentResponse.custom_issue_entry_url', index=14,
      number=16, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='default_template_for_members', full_name='monorail.v1.FetchProjectEnvironmentResponse.default_template_for_members', index=15,
      number=17, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='default_template_for_non_members', full_name='monorail.v1.FetchProjectEnvironmentResponse.default_template_for_non_members', index=16,
      number=18, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='project_name', full_name='monorail.v1.FetchProjectEnvironmentResponse.project_name', index=17,
      number=19, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='project_summary', full_name='monorail.v1.FetchProjectEnvironmentResponse.project_summary', index=18,
      number=20, type=9, cpp_type=9, label=1,
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
  serialized_start=516,
  serialized_end=1262,
)

_LISTISSUETEMPLATESRESPONSE.fields_by_name['templates'].message_type = api_dot_v1_dot_api__proto_dot_project__objects__pb2._ISSUETEMPLATE
_FETCHPROJECTENVIRONMENTRESPONSE.fields_by_name['status_defs'].message_type = api_dot_v1_dot_api__proto_dot_project__objects__pb2._STATUSDEF
_FETCHPROJECTENVIRONMENTRESPONSE.fields_by_name['label_defs'].message_type = api_dot_v1_dot_api__proto_dot_project__objects__pb2._LABELDEF
_FETCHPROJECTENVIRONMENTRESPONSE.fields_by_name['component_defs'].message_type = api_dot_v1_dot_api__proto_dot_project__objects__pb2._COMPONENTDEF
_FETCHPROJECTENVIRONMENTRESPONSE.fields_by_name['field_defs'].message_type = api_dot_v1_dot_api__proto_dot_project__objects__pb2._FIELDDEF
_FETCHPROJECTENVIRONMENTRESPONSE.fields_by_name['approval_defs'].message_type = api_dot_v1_dot_api__proto_dot_project__objects__pb2._APPROVALDEF
_FETCHPROJECTENVIRONMENTRESPONSE.fields_by_name['statuses_offer_merge'].message_type = api_dot_v1_dot_api__proto_dot_project__objects__pb2._STATUSDEF
_FETCHPROJECTENVIRONMENTRESPONSE.fields_by_name['default_template_for_members'].message_type = api_dot_v1_dot_api__proto_dot_project__objects__pb2._ISSUETEMPLATE
_FETCHPROJECTENVIRONMENTRESPONSE.fields_by_name['default_template_for_non_members'].message_type = api_dot_v1_dot_api__proto_dot_project__objects__pb2._ISSUETEMPLATE
DESCRIPTOR.message_types_by_name['ListIssueTemplatesRequest'] = _LISTISSUETEMPLATESREQUEST
DESCRIPTOR.message_types_by_name['ListIssueTemplatesResponse'] = _LISTISSUETEMPLATESRESPONSE
DESCRIPTOR.message_types_by_name['FetchProjectEnvironmentRequest'] = _FETCHPROJECTENVIRONMENTREQUEST
DESCRIPTOR.message_types_by_name['FetchProjectEnvironmentResponse'] = _FETCHPROJECTENVIRONMENTRESPONSE
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

ListIssueTemplatesRequest = _reflection.GeneratedProtocolMessageType('ListIssueTemplatesRequest', (_message.Message,), dict(
  DESCRIPTOR = _LISTISSUETEMPLATESREQUEST,
  __module__ = 'api.v1.api_proto.projects_pb2'
  # @@protoc_insertion_point(class_scope:monorail.v1.ListIssueTemplatesRequest)
  ))
_sym_db.RegisterMessage(ListIssueTemplatesRequest)

ListIssueTemplatesResponse = _reflection.GeneratedProtocolMessageType('ListIssueTemplatesResponse', (_message.Message,), dict(
  DESCRIPTOR = _LISTISSUETEMPLATESRESPONSE,
  __module__ = 'api.v1.api_proto.projects_pb2'
  # @@protoc_insertion_point(class_scope:monorail.v1.ListIssueTemplatesResponse)
  ))
_sym_db.RegisterMessage(ListIssueTemplatesResponse)

FetchProjectEnvironmentRequest = _reflection.GeneratedProtocolMessageType('FetchProjectEnvironmentRequest', (_message.Message,), dict(
  DESCRIPTOR = _FETCHPROJECTENVIRONMENTREQUEST,
  __module__ = 'api.v1.api_proto.projects_pb2'
  # @@protoc_insertion_point(class_scope:monorail.v1.FetchProjectEnvironmentRequest)
  ))
_sym_db.RegisterMessage(FetchProjectEnvironmentRequest)

FetchProjectEnvironmentResponse = _reflection.GeneratedProtocolMessageType('FetchProjectEnvironmentResponse', (_message.Message,), dict(
  DESCRIPTOR = _FETCHPROJECTENVIRONMENTRESPONSE,
  __module__ = 'api.v1.api_proto.projects_pb2'
  # @@protoc_insertion_point(class_scope:monorail.v1.FetchProjectEnvironmentResponse)
  ))
_sym_db.RegisterMessage(FetchProjectEnvironmentResponse)


_LISTISSUETEMPLATESREQUEST.fields_by_name['parent']._options = None
_FETCHPROJECTENVIRONMENTREQUEST.fields_by_name['parent']._options = None

_PROJECTS = _descriptor.ServiceDescriptor(
  name='Projects',
  full_name='monorail.v1.Projects',
  file=DESCRIPTOR,
  index=0,
  serialized_options=None,
  serialized_start=1265,
  serialized_end=1619,
  methods=[
  _descriptor.MethodDescriptor(
    name='ListIssueTemplates',
    full_name='monorail.v1.Projects.ListIssueTemplates',
    index=0,
    containing_service=None,
    input_type=_LISTISSUETEMPLATESREQUEST,
    output_type=_LISTISSUETEMPLATESRESPONSE,
    serialized_options=_b('\202\323\344\223\0022\"-/prpc/monorail.v1.Projects/ListIssueTemplates:\001*'),
  ),
  _descriptor.MethodDescriptor(
    name='FetchProjectEnvironment',
    full_name='monorail.v1.Projects.FetchProjectEnvironment',
    index=1,
    containing_service=None,
    input_type=_FETCHPROJECTENVIRONMENTREQUEST,
    output_type=_FETCHPROJECTENVIRONMENTRESPONSE,
    serialized_options=_b('\202\323\344\223\0027\"2/prpc/monorail.v1.Projects/FetchProjectEnvironment:\001*'),
  ),
])
_sym_db.RegisterServiceDescriptor(_PROJECTS)

DESCRIPTOR.services_by_name['Projects'] = _PROJECTS

# @@protoc_insertion_point(module_scope)
