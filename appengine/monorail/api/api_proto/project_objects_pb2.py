# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: api/api_proto/project_objects.proto

import sys
_b=sys.version_info[0]<3 and (lambda x:x) or (lambda x:x.encode('latin1'))
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
from google.protobuf import descriptor_pb2
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from api.api_proto import common_pb2 as api_dot_api__proto_dot_common__pb2


DESCRIPTOR = _descriptor.FileDescriptor(
  name='api/api_proto/project_objects.proto',
  package='monorail',
  syntax='proto3',
  serialized_pb=_b('\n#api/api_proto/project_objects.proto\x12\x08monorail\x1a\x1a\x61pi/api_proto/common.proto\"=\n\x07Project\x12\x0c\n\x04name\x18\x01 \x01(\t\x12\x0f\n\x07summary\x18\x02 \x01(\t\x12\x13\n\x0b\x64\x65scription\x18\x03 \x01(\t\"d\n\tStatusDef\x12\x0e\n\x06status\x18\x01 \x01(\t\x12\x12\n\nmeans_open\x18\x02 \x01(\x08\x12\x0c\n\x04rank\x18\x03 \x01(\r\x12\x11\n\tdocstring\x18\x04 \x01(\t\x12\x12\n\ndeprecated\x18\x05 \x01(\x08\"N\n\x08LabelDef\x12\r\n\x05label\x18\x01 \x01(\t\x12\x0c\n\x04rank\x18\x02 \x01(\r\x12\x11\n\tdocstring\x18\x03 \x01(\t\x12\x12\n\ndeprecated\x18\x04 \x01(\x08\"\xaa\x02\n\x0c\x43omponentDef\x12\x0c\n\x04path\x18\x01 \x01(\t\x12\x11\n\tdocstring\x18\x02 \x01(\t\x12%\n\nadmin_refs\x18\x03 \x03(\x0b\x32\x11.monorail.UserRef\x12\"\n\x07\x63\x63_refs\x18\x04 \x03(\x0b\x32\x11.monorail.UserRef\x12\x12\n\ndeprecated\x18\x05 \x01(\x08\x12\x0f\n\x07\x63reated\x18\x06 \x01(\x07\x12&\n\x0b\x63reator_ref\x18\x07 \x01(\x0b\x32\x11.monorail.UserRef\x12\x10\n\x08modified\x18\x08 \x01(\x07\x12\'\n\x0cmodifier_ref\x18\t \x01(\x0b\x32\x11.monorail.UserRef\x12&\n\nlabel_refs\x18\n \x03(\x0b\x32\x12.monorail.LabelRef\"\xdb\x01\n\x08\x46ieldDef\x12%\n\tfield_ref\x18\x01 \x01(\x0b\x32\x12.monorail.FieldRef\x12\x17\n\x0f\x61pplicable_type\x18\x02 \x01(\t\x12\x13\n\x0bis_required\x18\x03 \x01(\x08\x12\x10\n\x08is_niche\x18\x04 \x01(\x08\x12\x16\n\x0eis_multivalued\x18\x05 \x01(\x08\x12\x11\n\tdocstring\x18\x06 \x01(\t\x12%\n\nadmin_refs\x18\x07 \x03(\x0b\x32\x11.monorail.UserRef\x12\x16\n\x0eis_phase_field\x18\x08 \x01(\x08\"[\n\x0c\x46ieldOptions\x12%\n\tfield_ref\x18\x01 \x01(\x0b\x32\x12.monorail.FieldRef\x12$\n\tuser_refs\x18\x02 \x03(\x0b\x32\x11.monorail.UserRef\"n\n\x0b\x41pprovalDef\x12%\n\tfield_ref\x18\x01 \x01(\x0b\x32\x12.monorail.FieldRef\x12(\n\rapprover_refs\x18\x02 \x03(\x0b\x32\x11.monorail.UserRef\x12\x0e\n\x06survey\x18\x03 \x01(\t\"\xe6\x02\n\x06\x43onfig\x12\x14\n\x0cproject_name\x18\x01 \x01(\t\x12(\n\x0bstatus_defs\x18\x02 \x03(\x0b\x32\x13.monorail.StatusDef\x12\x31\n\x14statuses_offer_merge\x18\x03 \x03(\x0b\x32\x13.monorail.StatusRef\x12&\n\nlabel_defs\x18\x04 \x03(\x0b\x32\x12.monorail.LabelDef\x12 \n\x18\x65xclusive_label_prefixes\x18\x05 \x03(\t\x12.\n\x0e\x63omponent_defs\x18\x06 \x03(\x0b\x32\x16.monorail.ComponentDef\x12&\n\nfield_defs\x18\x07 \x03(\x0b\x32\x12.monorail.FieldDef\x12,\n\rapproval_defs\x18\x08 \x03(\x0b\x32\x15.monorail.ApprovalDef\x12\x19\n\x11restrict_to_known\x18\t \x01(\x08\x62\x06proto3')
  ,
  dependencies=[api_dot_api__proto_dot_common__pb2.DESCRIPTOR,])
_sym_db.RegisterFileDescriptor(DESCRIPTOR)




_PROJECT = _descriptor.Descriptor(
  name='Project',
  full_name='monorail.Project',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='name', full_name='monorail.Project.name', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='summary', full_name='monorail.Project.summary', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='description', full_name='monorail.Project.description', index=2,
      number=3, type=9, cpp_type=9, label=1,
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
  serialized_start=77,
  serialized_end=138,
)


_STATUSDEF = _descriptor.Descriptor(
  name='StatusDef',
  full_name='monorail.StatusDef',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='status', full_name='monorail.StatusDef.status', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='means_open', full_name='monorail.StatusDef.means_open', index=1,
      number=2, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='rank', full_name='monorail.StatusDef.rank', index=2,
      number=3, type=13, cpp_type=3, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='docstring', full_name='monorail.StatusDef.docstring', index=3,
      number=4, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='deprecated', full_name='monorail.StatusDef.deprecated', index=4,
      number=5, type=8, cpp_type=7, label=1,
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
  serialized_start=140,
  serialized_end=240,
)


_LABELDEF = _descriptor.Descriptor(
  name='LabelDef',
  full_name='monorail.LabelDef',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='label', full_name='monorail.LabelDef.label', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='rank', full_name='monorail.LabelDef.rank', index=1,
      number=2, type=13, cpp_type=3, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='docstring', full_name='monorail.LabelDef.docstring', index=2,
      number=3, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='deprecated', full_name='monorail.LabelDef.deprecated', index=3,
      number=4, type=8, cpp_type=7, label=1,
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
  serialized_start=242,
  serialized_end=320,
)


_COMPONENTDEF = _descriptor.Descriptor(
  name='ComponentDef',
  full_name='monorail.ComponentDef',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='path', full_name='monorail.ComponentDef.path', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='docstring', full_name='monorail.ComponentDef.docstring', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='admin_refs', full_name='monorail.ComponentDef.admin_refs', index=2,
      number=3, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='cc_refs', full_name='monorail.ComponentDef.cc_refs', index=3,
      number=4, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='deprecated', full_name='monorail.ComponentDef.deprecated', index=4,
      number=5, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='created', full_name='monorail.ComponentDef.created', index=5,
      number=6, type=7, cpp_type=3, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='creator_ref', full_name='monorail.ComponentDef.creator_ref', index=6,
      number=7, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='modified', full_name='monorail.ComponentDef.modified', index=7,
      number=8, type=7, cpp_type=3, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='modifier_ref', full_name='monorail.ComponentDef.modifier_ref', index=8,
      number=9, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='label_refs', full_name='monorail.ComponentDef.label_refs', index=9,
      number=10, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
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
  serialized_start=323,
  serialized_end=621,
)


_FIELDDEF = _descriptor.Descriptor(
  name='FieldDef',
  full_name='monorail.FieldDef',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='field_ref', full_name='monorail.FieldDef.field_ref', index=0,
      number=1, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='applicable_type', full_name='monorail.FieldDef.applicable_type', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='is_required', full_name='monorail.FieldDef.is_required', index=2,
      number=3, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='is_niche', full_name='monorail.FieldDef.is_niche', index=3,
      number=4, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='is_multivalued', full_name='monorail.FieldDef.is_multivalued', index=4,
      number=5, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='docstring', full_name='monorail.FieldDef.docstring', index=5,
      number=6, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='admin_refs', full_name='monorail.FieldDef.admin_refs', index=6,
      number=7, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='is_phase_field', full_name='monorail.FieldDef.is_phase_field', index=7,
      number=8, type=8, cpp_type=7, label=1,
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
  serialized_start=624,
  serialized_end=843,
)


_FIELDOPTIONS = _descriptor.Descriptor(
  name='FieldOptions',
  full_name='monorail.FieldOptions',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='field_ref', full_name='monorail.FieldOptions.field_ref', index=0,
      number=1, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='user_refs', full_name='monorail.FieldOptions.user_refs', index=1,
      number=2, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
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
  serialized_start=845,
  serialized_end=936,
)


_APPROVALDEF = _descriptor.Descriptor(
  name='ApprovalDef',
  full_name='monorail.ApprovalDef',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='field_ref', full_name='monorail.ApprovalDef.field_ref', index=0,
      number=1, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='approver_refs', full_name='monorail.ApprovalDef.approver_refs', index=1,
      number=2, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='survey', full_name='monorail.ApprovalDef.survey', index=2,
      number=3, type=9, cpp_type=9, label=1,
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
  serialized_start=938,
  serialized_end=1048,
)


_CONFIG = _descriptor.Descriptor(
  name='Config',
  full_name='monorail.Config',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='project_name', full_name='monorail.Config.project_name', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='status_defs', full_name='monorail.Config.status_defs', index=1,
      number=2, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='statuses_offer_merge', full_name='monorail.Config.statuses_offer_merge', index=2,
      number=3, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='label_defs', full_name='monorail.Config.label_defs', index=3,
      number=4, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='exclusive_label_prefixes', full_name='monorail.Config.exclusive_label_prefixes', index=4,
      number=5, type=9, cpp_type=9, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='component_defs', full_name='monorail.Config.component_defs', index=5,
      number=6, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='field_defs', full_name='monorail.Config.field_defs', index=6,
      number=7, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='approval_defs', full_name='monorail.Config.approval_defs', index=7,
      number=8, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='restrict_to_known', full_name='monorail.Config.restrict_to_known', index=8,
      number=9, type=8, cpp_type=7, label=1,
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
  serialized_start=1051,
  serialized_end=1409,
)

_COMPONENTDEF.fields_by_name['admin_refs'].message_type = api_dot_api__proto_dot_common__pb2._USERREF
_COMPONENTDEF.fields_by_name['cc_refs'].message_type = api_dot_api__proto_dot_common__pb2._USERREF
_COMPONENTDEF.fields_by_name['creator_ref'].message_type = api_dot_api__proto_dot_common__pb2._USERREF
_COMPONENTDEF.fields_by_name['modifier_ref'].message_type = api_dot_api__proto_dot_common__pb2._USERREF
_COMPONENTDEF.fields_by_name['label_refs'].message_type = api_dot_api__proto_dot_common__pb2._LABELREF
_FIELDDEF.fields_by_name['field_ref'].message_type = api_dot_api__proto_dot_common__pb2._FIELDREF
_FIELDDEF.fields_by_name['admin_refs'].message_type = api_dot_api__proto_dot_common__pb2._USERREF
_FIELDOPTIONS.fields_by_name['field_ref'].message_type = api_dot_api__proto_dot_common__pb2._FIELDREF
_FIELDOPTIONS.fields_by_name['user_refs'].message_type = api_dot_api__proto_dot_common__pb2._USERREF
_APPROVALDEF.fields_by_name['field_ref'].message_type = api_dot_api__proto_dot_common__pb2._FIELDREF
_APPROVALDEF.fields_by_name['approver_refs'].message_type = api_dot_api__proto_dot_common__pb2._USERREF
_CONFIG.fields_by_name['status_defs'].message_type = _STATUSDEF
_CONFIG.fields_by_name['statuses_offer_merge'].message_type = api_dot_api__proto_dot_common__pb2._STATUSREF
_CONFIG.fields_by_name['label_defs'].message_type = _LABELDEF
_CONFIG.fields_by_name['component_defs'].message_type = _COMPONENTDEF
_CONFIG.fields_by_name['field_defs'].message_type = _FIELDDEF
_CONFIG.fields_by_name['approval_defs'].message_type = _APPROVALDEF
DESCRIPTOR.message_types_by_name['Project'] = _PROJECT
DESCRIPTOR.message_types_by_name['StatusDef'] = _STATUSDEF
DESCRIPTOR.message_types_by_name['LabelDef'] = _LABELDEF
DESCRIPTOR.message_types_by_name['ComponentDef'] = _COMPONENTDEF
DESCRIPTOR.message_types_by_name['FieldDef'] = _FIELDDEF
DESCRIPTOR.message_types_by_name['FieldOptions'] = _FIELDOPTIONS
DESCRIPTOR.message_types_by_name['ApprovalDef'] = _APPROVALDEF
DESCRIPTOR.message_types_by_name['Config'] = _CONFIG

Project = _reflection.GeneratedProtocolMessageType('Project', (_message.Message,), dict(
  DESCRIPTOR = _PROJECT,
  __module__ = 'api.api_proto.project_objects_pb2'
  # @@protoc_insertion_point(class_scope:monorail.Project)
  ))
_sym_db.RegisterMessage(Project)

StatusDef = _reflection.GeneratedProtocolMessageType('StatusDef', (_message.Message,), dict(
  DESCRIPTOR = _STATUSDEF,
  __module__ = 'api.api_proto.project_objects_pb2'
  # @@protoc_insertion_point(class_scope:monorail.StatusDef)
  ))
_sym_db.RegisterMessage(StatusDef)

LabelDef = _reflection.GeneratedProtocolMessageType('LabelDef', (_message.Message,), dict(
  DESCRIPTOR = _LABELDEF,
  __module__ = 'api.api_proto.project_objects_pb2'
  # @@protoc_insertion_point(class_scope:monorail.LabelDef)
  ))
_sym_db.RegisterMessage(LabelDef)

ComponentDef = _reflection.GeneratedProtocolMessageType('ComponentDef', (_message.Message,), dict(
  DESCRIPTOR = _COMPONENTDEF,
  __module__ = 'api.api_proto.project_objects_pb2'
  # @@protoc_insertion_point(class_scope:monorail.ComponentDef)
  ))
_sym_db.RegisterMessage(ComponentDef)

FieldDef = _reflection.GeneratedProtocolMessageType('FieldDef', (_message.Message,), dict(
  DESCRIPTOR = _FIELDDEF,
  __module__ = 'api.api_proto.project_objects_pb2'
  # @@protoc_insertion_point(class_scope:monorail.FieldDef)
  ))
_sym_db.RegisterMessage(FieldDef)

FieldOptions = _reflection.GeneratedProtocolMessageType('FieldOptions', (_message.Message,), dict(
  DESCRIPTOR = _FIELDOPTIONS,
  __module__ = 'api.api_proto.project_objects_pb2'
  # @@protoc_insertion_point(class_scope:monorail.FieldOptions)
  ))
_sym_db.RegisterMessage(FieldOptions)

ApprovalDef = _reflection.GeneratedProtocolMessageType('ApprovalDef', (_message.Message,), dict(
  DESCRIPTOR = _APPROVALDEF,
  __module__ = 'api.api_proto.project_objects_pb2'
  # @@protoc_insertion_point(class_scope:monorail.ApprovalDef)
  ))
_sym_db.RegisterMessage(ApprovalDef)

Config = _reflection.GeneratedProtocolMessageType('Config', (_message.Message,), dict(
  DESCRIPTOR = _CONFIG,
  __module__ = 'api.api_proto.project_objects_pb2'
  # @@protoc_insertion_point(class_scope:monorail.Config)
  ))
_sym_db.RegisterMessage(Config)


# @@protoc_insertion_point(module_scope)
