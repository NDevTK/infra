# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: acquisition_task.proto

import sys
_b=sys.version_info[0]<3 and (lambda x:x) or (lambda x:x.encode('latin1'))
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
from google.protobuf import descriptor_pb2
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor.FileDescriptor(
  name='acquisition_task.proto',
  package='ts_mon.proto',
  syntax='proto2',
  serialized_pb=_b('\n\x16\x61\x63quisition_task.proto\x12\x0cts_mon.proto\"\xd3\x01\n\x04Task\x12\x19\n\x11proxy_environment\x18\x05 \x01(\t\x12\x18\n\x10\x61\x63quisition_name\x18\n \x01(\t\x12\x14\n\x0cservice_name\x18\x14 \x01(\t\x12\x10\n\x08job_name\x18\x1e \x01(\t\x12\x13\n\x0b\x64\x61ta_center\x18( \x01(\t\x12\x11\n\thost_name\x18\x32 \x01(\t\x12\x10\n\x08task_num\x18< \x01(\x05\x12\x12\n\nproxy_zone\x18\x46 \x01(\t\" \n\x06TypeId\x12\x16\n\x0fMESSAGE_TYPE_ID\x10\xd5\x9d\x9e\x10')
)
_sym_db.RegisterFileDescriptor(DESCRIPTOR)



_TASK_TYPEID = _descriptor.EnumDescriptor(
  name='TypeId',
  full_name='ts_mon.proto.Task.TypeId',
  filename=None,
  file=DESCRIPTOR,
  values=[
    _descriptor.EnumValueDescriptor(
      name='MESSAGE_TYPE_ID', index=0, number=34049749,
      options=None,
      type=None),
  ],
  containing_type=None,
  options=None,
  serialized_start=220,
  serialized_end=252,
)
_sym_db.RegisterEnumDescriptor(_TASK_TYPEID)


_TASK = _descriptor.Descriptor(
  name='Task',
  full_name='ts_mon.proto.Task',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='proxy_environment', full_name='ts_mon.proto.Task.proxy_environment', index=0,
      number=5, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='acquisition_name', full_name='ts_mon.proto.Task.acquisition_name', index=1,
      number=10, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='service_name', full_name='ts_mon.proto.Task.service_name', index=2,
      number=20, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='job_name', full_name='ts_mon.proto.Task.job_name', index=3,
      number=30, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='data_center', full_name='ts_mon.proto.Task.data_center', index=4,
      number=40, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='host_name', full_name='ts_mon.proto.Task.host_name', index=5,
      number=50, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='task_num', full_name='ts_mon.proto.Task.task_num', index=6,
      number=60, type=5, cpp_type=1, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='proxy_zone', full_name='ts_mon.proto.Task.proxy_zone', index=7,
      number=70, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
    _TASK_TYPEID,
  ],
  options=None,
  is_extendable=False,
  syntax='proto2',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=41,
  serialized_end=252,
)

_TASK_TYPEID.containing_type = _TASK
DESCRIPTOR.message_types_by_name['Task'] = _TASK

Task = _reflection.GeneratedProtocolMessageType('Task', (_message.Message,), dict(
  DESCRIPTOR = _TASK,
  __module__ = 'acquisition_task_pb2'
  # @@protoc_insertion_point(class_scope:ts_mon.proto.Task)
  ))
_sym_db.RegisterMessage(Task)


# @@protoc_insertion_point(module_scope)
