# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: chromeperf/pinpoint/comparison.proto

from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor.FileDescriptor(
  name='chromeperf/pinpoint/comparison.proto',
  package='chromeperf.pinpoint',
  syntax='proto3',
  serialized_options=None,
  create_key=_descriptor._internal_create_key,
  serialized_pb=b'\n$chromeperf/pinpoint/comparison.proto\x12\x13\x63hromeperf.pinpoint\"\xef\x01\n\nComparison\x12=\n\x06result\x18\x01 \x01(\x0e\x32-.chromeperf.pinpoint.Comparison.CompareResult\x12\x0f\n\x07p_value\x18\x02 \x01(\x01\x12\x15\n\rlow_threshold\x18\x03 \x01(\x01\x12\x16\n\x0ehigh_threshold\x18\x04 \x01(\x01\"b\n\rCompareResult\x12\x1e\n\x1a\x43OMPARE_RESULT_UNSPECIFIED\x10\x00\x12\r\n\tDIFFERENT\x10\x01\x12\x08\n\x04SAME\x10\x02\x12\x0b\n\x07UNKNOWN\x10\x03\x12\x0b\n\x07PENDING\x10\x04\x62\x06proto3'
)



_COMPARISON_COMPARERESULT = _descriptor.EnumDescriptor(
  name='CompareResult',
  full_name='chromeperf.pinpoint.Comparison.CompareResult',
  filename=None,
  file=DESCRIPTOR,
  create_key=_descriptor._internal_create_key,
  values=[
    _descriptor.EnumValueDescriptor(
      name='COMPARE_RESULT_UNSPECIFIED', index=0, number=0,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='DIFFERENT', index=1, number=1,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='SAME', index=2, number=2,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='UNKNOWN', index=3, number=3,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='PENDING', index=4, number=4,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
  ],
  containing_type=None,
  serialized_options=None,
  serialized_start=203,
  serialized_end=301,
)
_sym_db.RegisterEnumDescriptor(_COMPARISON_COMPARERESULT)


_COMPARISON = _descriptor.Descriptor(
  name='Comparison',
  full_name='chromeperf.pinpoint.Comparison',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='result', full_name='chromeperf.pinpoint.Comparison.result', index=0,
      number=1, type=14, cpp_type=8, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='p_value', full_name='chromeperf.pinpoint.Comparison.p_value', index=1,
      number=2, type=1, cpp_type=5, label=1,
      has_default_value=False, default_value=float(0),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='low_threshold', full_name='chromeperf.pinpoint.Comparison.low_threshold', index=2,
      number=3, type=1, cpp_type=5, label=1,
      has_default_value=False, default_value=float(0),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='high_threshold', full_name='chromeperf.pinpoint.Comparison.high_threshold', index=3,
      number=4, type=1, cpp_type=5, label=1,
      has_default_value=False, default_value=float(0),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
    _COMPARISON_COMPARERESULT,
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=62,
  serialized_end=301,
)

_COMPARISON.fields_by_name['result'].enum_type = _COMPARISON_COMPARERESULT
_COMPARISON_COMPARERESULT.containing_type = _COMPARISON
DESCRIPTOR.message_types_by_name['Comparison'] = _COMPARISON
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

Comparison = _reflection.GeneratedProtocolMessageType('Comparison', (_message.Message,), {
  'DESCRIPTOR' : _COMPARISON,
  '__module__' : 'chromeperf.pinpoint.comparison_pb2'
  # @@protoc_insertion_point(class_scope:chromeperf.pinpoint.Comparison)
  })
_sym_db.RegisterMessage(Comparison)


# @@protoc_insertion_point(module_scope)