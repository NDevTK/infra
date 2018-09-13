# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: third_party_packages_ng/spec.proto

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
  name='third_party_packages_ng/spec.proto',
  package='',
  syntax='proto3',
  serialized_pb=_b('\n\"third_party_packages_ng/spec.proto\"\xcf\x05\n\x04Spec\x12\x1c\n\x06\x63reate\x18\x01 \x03(\x0b\x32\x0c.Spec.Create\x12\x1c\n\x06upload\x18\x02 \x01(\x0b\x32\x0c.Spec.Upload\x1a\xd9\x04\n\x06\x43reate\x12\x13\n\x0bplatform_re\x18\x01 \x01(\t\x12#\n\x06source\x18\x02 \x01(\x0b\x32\x13.Spec.Create.Source\x12!\n\x05\x62uild\x18\x03 \x01(\x0b\x32\x12.Spec.Create.Build\x12%\n\x07package\x18\x04 \x01(\x0b\x32\x14.Spec.Create.Package\x12#\n\x06verify\x18\x05 \x01(\x0b\x32\x13.Spec.Create.Verify\x12\x13\n\x0bunsupported\x18\x06 \x01(\x08\x1a\xc4\x01\n\x06Source\x12\x19\n\x03git\x18\x01 \x01(\x0b\x32\n.GitSourceH\x00\x12\x1b\n\x04\x63ipd\x18\x02 \x01(\x0b\x32\x0b.CipdSourceH\x00\x12\x1f\n\x06script\x18\x03 \x01(\x0b\x32\r.ScriptSourceH\x00\x12\x0e\n\x06subdir\x18\x04 \x01(\t\x12\x16\n\x0eunpack_archive\x18\x05 \x01(\x08\x12\x18\n\x10no_archive_prune\x18\x06 \x01(\x08\x12\x15\n\rpatch_version\x18\x07 \x01(\tB\x08\n\x06method\x1a\x33\n\x05\x42uild\x12\x0f\n\x07install\x18\x01 \x03(\t\x12\x0c\n\x04tool\x18\x02 \x03(\t\x12\x0b\n\x03\x64\x65p\x18\x03 \x03(\t\x1a}\n\x07Package\x12\x36\n\x0cinstall_mode\x18\x01 \x01(\x0e\x32 .Spec.Create.Package.InstallMode\x12\x14\n\x0cversion_file\x18\x02 \x01(\t\"$\n\x0bInstallMode\x12\x08\n\x04\x63opy\x10\x00\x12\x0b\n\x07symlink\x10\x01\x1a\x16\n\x06Verify\x12\x0c\n\x04test\x18\x01 \x03(\t\x1a/\n\x06Upload\x12\x12\n\npkg_prefix\x18\x01 \x01(\t\x12\x11\n\tuniversal\x18\x02 \x01(\x08\"W\n\tGitSource\x12\x0c\n\x04repo\x18\x01 \x01(\t\x12\x13\n\x0btag_pattern\x18\x02 \x01(\t\x12\x14\n\x0cversion_join\x18\x03 \x01(\t\x12\x11\n\tpatch_dir\x18\x04 \x03(\t\"2\n\nCipdSource\x12\x0b\n\x03pkg\x18\x01 \x01(\t\x12\x17\n\x0f\x64\x65\x66\x61ult_version\x18\x02 \x01(\t\"\x1c\n\x0cScriptSource\x12\x0c\n\x04name\x18\x01 \x01(\tb\x06proto3')
)
_sym_db.RegisterFileDescriptor(DESCRIPTOR)



_SPEC_CREATE_PACKAGE_INSTALLMODE = _descriptor.EnumDescriptor(
  name='InstallMode',
  full_name='Spec.Create.Package.InstallMode',
  filename=None,
  file=DESCRIPTOR,
  values=[
    _descriptor.EnumValueDescriptor(
      name='copy', index=0, number=0,
      options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='symlink', index=1, number=1,
      options=None,
      type=None),
  ],
  containing_type=None,
  options=None,
  serialized_start=649,
  serialized_end=685,
)
_sym_db.RegisterEnumDescriptor(_SPEC_CREATE_PACKAGE_INSTALLMODE)


_SPEC_CREATE_SOURCE = _descriptor.Descriptor(
  name='Source',
  full_name='Spec.Create.Source',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='git', full_name='Spec.Create.Source.git', index=0,
      number=1, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='cipd', full_name='Spec.Create.Source.cipd', index=1,
      number=2, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='script', full_name='Spec.Create.Source.script', index=2,
      number=3, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='subdir', full_name='Spec.Create.Source.subdir', index=3,
      number=4, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='unpack_archive', full_name='Spec.Create.Source.unpack_archive', index=4,
      number=5, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='no_archive_prune', full_name='Spec.Create.Source.no_archive_prune', index=5,
      number=6, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='patch_version', full_name='Spec.Create.Source.patch_version', index=6,
      number=7, type=9, cpp_type=9, label=1,
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
    _descriptor.OneofDescriptor(
      name='method', full_name='Spec.Create.Source.method',
      index=0, containing_type=None, fields=[]),
  ],
  serialized_start=309,
  serialized_end=505,
)

_SPEC_CREATE_BUILD = _descriptor.Descriptor(
  name='Build',
  full_name='Spec.Create.Build',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='install', full_name='Spec.Create.Build.install', index=0,
      number=1, type=9, cpp_type=9, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='tool', full_name='Spec.Create.Build.tool', index=1,
      number=2, type=9, cpp_type=9, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='dep', full_name='Spec.Create.Build.dep', index=2,
      number=3, type=9, cpp_type=9, label=3,
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
  serialized_start=507,
  serialized_end=558,
)

_SPEC_CREATE_PACKAGE = _descriptor.Descriptor(
  name='Package',
  full_name='Spec.Create.Package',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='install_mode', full_name='Spec.Create.Package.install_mode', index=0,
      number=1, type=14, cpp_type=8, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='version_file', full_name='Spec.Create.Package.version_file', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
    _SPEC_CREATE_PACKAGE_INSTALLMODE,
  ],
  options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=560,
  serialized_end=685,
)

_SPEC_CREATE_VERIFY = _descriptor.Descriptor(
  name='Verify',
  full_name='Spec.Create.Verify',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='test', full_name='Spec.Create.Verify.test', index=0,
      number=1, type=9, cpp_type=9, label=3,
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
  serialized_start=687,
  serialized_end=709,
)

_SPEC_CREATE = _descriptor.Descriptor(
  name='Create',
  full_name='Spec.Create',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='platform_re', full_name='Spec.Create.platform_re', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='source', full_name='Spec.Create.source', index=1,
      number=2, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='build', full_name='Spec.Create.build', index=2,
      number=3, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='package', full_name='Spec.Create.package', index=3,
      number=4, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='verify', full_name='Spec.Create.verify', index=4,
      number=5, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='unsupported', full_name='Spec.Create.unsupported', index=5,
      number=6, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
  ],
  extensions=[
  ],
  nested_types=[_SPEC_CREATE_SOURCE, _SPEC_CREATE_BUILD, _SPEC_CREATE_PACKAGE, _SPEC_CREATE_VERIFY, ],
  enum_types=[
  ],
  options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=108,
  serialized_end=709,
)

_SPEC_UPLOAD = _descriptor.Descriptor(
  name='Upload',
  full_name='Spec.Upload',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='pkg_prefix', full_name='Spec.Upload.pkg_prefix', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='universal', full_name='Spec.Upload.universal', index=1,
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
  serialized_start=711,
  serialized_end=758,
)

_SPEC = _descriptor.Descriptor(
  name='Spec',
  full_name='Spec',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='create', full_name='Spec.create', index=0,
      number=1, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='upload', full_name='Spec.upload', index=1,
      number=2, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
  ],
  extensions=[
  ],
  nested_types=[_SPEC_CREATE, _SPEC_UPLOAD, ],
  enum_types=[
  ],
  options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=39,
  serialized_end=758,
)


_GITSOURCE = _descriptor.Descriptor(
  name='GitSource',
  full_name='GitSource',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='repo', full_name='GitSource.repo', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='tag_pattern', full_name='GitSource.tag_pattern', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='version_join', full_name='GitSource.version_join', index=2,
      number=3, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='patch_dir', full_name='GitSource.patch_dir', index=3,
      number=4, type=9, cpp_type=9, label=3,
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
  serialized_start=760,
  serialized_end=847,
)


_CIPDSOURCE = _descriptor.Descriptor(
  name='CipdSource',
  full_name='CipdSource',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='pkg', full_name='CipdSource.pkg', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None),
    _descriptor.FieldDescriptor(
      name='default_version', full_name='CipdSource.default_version', index=1,
      number=2, type=9, cpp_type=9, label=1,
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
  serialized_start=849,
  serialized_end=899,
)


_SCRIPTSOURCE = _descriptor.Descriptor(
  name='ScriptSource',
  full_name='ScriptSource',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='name', full_name='ScriptSource.name', index=0,
      number=1, type=9, cpp_type=9, label=1,
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
  serialized_start=901,
  serialized_end=929,
)

_SPEC_CREATE_SOURCE.fields_by_name['git'].message_type = _GITSOURCE
_SPEC_CREATE_SOURCE.fields_by_name['cipd'].message_type = _CIPDSOURCE
_SPEC_CREATE_SOURCE.fields_by_name['script'].message_type = _SCRIPTSOURCE
_SPEC_CREATE_SOURCE.containing_type = _SPEC_CREATE
_SPEC_CREATE_SOURCE.oneofs_by_name['method'].fields.append(
  _SPEC_CREATE_SOURCE.fields_by_name['git'])
_SPEC_CREATE_SOURCE.fields_by_name['git'].containing_oneof = _SPEC_CREATE_SOURCE.oneofs_by_name['method']
_SPEC_CREATE_SOURCE.oneofs_by_name['method'].fields.append(
  _SPEC_CREATE_SOURCE.fields_by_name['cipd'])
_SPEC_CREATE_SOURCE.fields_by_name['cipd'].containing_oneof = _SPEC_CREATE_SOURCE.oneofs_by_name['method']
_SPEC_CREATE_SOURCE.oneofs_by_name['method'].fields.append(
  _SPEC_CREATE_SOURCE.fields_by_name['script'])
_SPEC_CREATE_SOURCE.fields_by_name['script'].containing_oneof = _SPEC_CREATE_SOURCE.oneofs_by_name['method']
_SPEC_CREATE_BUILD.containing_type = _SPEC_CREATE
_SPEC_CREATE_PACKAGE.fields_by_name['install_mode'].enum_type = _SPEC_CREATE_PACKAGE_INSTALLMODE
_SPEC_CREATE_PACKAGE.containing_type = _SPEC_CREATE
_SPEC_CREATE_PACKAGE_INSTALLMODE.containing_type = _SPEC_CREATE_PACKAGE
_SPEC_CREATE_VERIFY.containing_type = _SPEC_CREATE
_SPEC_CREATE.fields_by_name['source'].message_type = _SPEC_CREATE_SOURCE
_SPEC_CREATE.fields_by_name['build'].message_type = _SPEC_CREATE_BUILD
_SPEC_CREATE.fields_by_name['package'].message_type = _SPEC_CREATE_PACKAGE
_SPEC_CREATE.fields_by_name['verify'].message_type = _SPEC_CREATE_VERIFY
_SPEC_CREATE.containing_type = _SPEC
_SPEC_UPLOAD.containing_type = _SPEC
_SPEC.fields_by_name['create'].message_type = _SPEC_CREATE
_SPEC.fields_by_name['upload'].message_type = _SPEC_UPLOAD
DESCRIPTOR.message_types_by_name['Spec'] = _SPEC
DESCRIPTOR.message_types_by_name['GitSource'] = _GITSOURCE
DESCRIPTOR.message_types_by_name['CipdSource'] = _CIPDSOURCE
DESCRIPTOR.message_types_by_name['ScriptSource'] = _SCRIPTSOURCE

Spec = _reflection.GeneratedProtocolMessageType('Spec', (_message.Message,), dict(

  Create = _reflection.GeneratedProtocolMessageType('Create', (_message.Message,), dict(

    Source = _reflection.GeneratedProtocolMessageType('Source', (_message.Message,), dict(
      DESCRIPTOR = _SPEC_CREATE_SOURCE,
      __module__ = 'third_party_packages_ng.spec_pb2'
      # @@protoc_insertion_point(class_scope:Spec.Create.Source)
      ))
    ,

    Build = _reflection.GeneratedProtocolMessageType('Build', (_message.Message,), dict(
      DESCRIPTOR = _SPEC_CREATE_BUILD,
      __module__ = 'third_party_packages_ng.spec_pb2'
      # @@protoc_insertion_point(class_scope:Spec.Create.Build)
      ))
    ,

    Package = _reflection.GeneratedProtocolMessageType('Package', (_message.Message,), dict(
      DESCRIPTOR = _SPEC_CREATE_PACKAGE,
      __module__ = 'third_party_packages_ng.spec_pb2'
      # @@protoc_insertion_point(class_scope:Spec.Create.Package)
      ))
    ,

    Verify = _reflection.GeneratedProtocolMessageType('Verify', (_message.Message,), dict(
      DESCRIPTOR = _SPEC_CREATE_VERIFY,
      __module__ = 'third_party_packages_ng.spec_pb2'
      # @@protoc_insertion_point(class_scope:Spec.Create.Verify)
      ))
    ,
    DESCRIPTOR = _SPEC_CREATE,
    __module__ = 'third_party_packages_ng.spec_pb2'
    # @@protoc_insertion_point(class_scope:Spec.Create)
    ))
  ,

  Upload = _reflection.GeneratedProtocolMessageType('Upload', (_message.Message,), dict(
    DESCRIPTOR = _SPEC_UPLOAD,
    __module__ = 'third_party_packages_ng.spec_pb2'
    # @@protoc_insertion_point(class_scope:Spec.Upload)
    ))
  ,
  DESCRIPTOR = _SPEC,
  __module__ = 'third_party_packages_ng.spec_pb2'
  # @@protoc_insertion_point(class_scope:Spec)
  ))
_sym_db.RegisterMessage(Spec)
_sym_db.RegisterMessage(Spec.Create)
_sym_db.RegisterMessage(Spec.Create.Source)
_sym_db.RegisterMessage(Spec.Create.Build)
_sym_db.RegisterMessage(Spec.Create.Package)
_sym_db.RegisterMessage(Spec.Create.Verify)
_sym_db.RegisterMessage(Spec.Upload)

GitSource = _reflection.GeneratedProtocolMessageType('GitSource', (_message.Message,), dict(
  DESCRIPTOR = _GITSOURCE,
  __module__ = 'third_party_packages_ng.spec_pb2'
  # @@protoc_insertion_point(class_scope:GitSource)
  ))
_sym_db.RegisterMessage(GitSource)

CipdSource = _reflection.GeneratedProtocolMessageType('CipdSource', (_message.Message,), dict(
  DESCRIPTOR = _CIPDSOURCE,
  __module__ = 'third_party_packages_ng.spec_pb2'
  # @@protoc_insertion_point(class_scope:CipdSource)
  ))
_sym_db.RegisterMessage(CipdSource)

ScriptSource = _reflection.GeneratedProtocolMessageType('ScriptSource', (_message.Message,), dict(
  DESCRIPTOR = _SCRIPTSOURCE,
  __module__ = 'third_party_packages_ng.spec_pb2'
  # @@protoc_insertion_point(class_scope:ScriptSource)
  ))
_sym_db.RegisterMessage(ScriptSource)


# @@protoc_insertion_point(module_scope)
