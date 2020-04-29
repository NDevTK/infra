# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: api/v1/api_proto/user_objects.proto

import sys
_b=sys.version_info[0]<3 and (lambda x:x) or (lambda x:x.encode('latin1'))
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google_proto.google.api import resource_pb2 as google__proto_dot_google_dot_api_dot_resource__pb2
from google_proto.google.api import field_behavior_pb2 as google__proto_dot_google_dot_api_dot_field__behavior__pb2


DESCRIPTOR = _descriptor.FileDescriptor(
  name='api/v1/api_proto/user_objects.proto',
  package='monorail.v1',
  syntax='proto3',
  serialized_options=None,
  serialized_pb=_b('\n#api/v1/api_proto/user_objects.proto\x12\x0bmonorail.v1\x1a&google_proto/google/api/resource.proto\x1a,google_proto/google/api/field_behavior.proto\"w\n\x04User\x12\x0c\n\x04name\x18\x01 \x01(\t\x12\x19\n\x0c\x64isplay_name\x18\x02 \x01(\tB\x03\xe0\x41\x03\x12\x1c\n\x14\x61vailability_message\x18\x03 \x01(\t:(\xea\x41%\n\x12\x61pi.crbug.com/User\x12\x0fusers/{user_id}\"\x9a\t\n\x0cUserSettings\x12-\n\x04name\x18\x01 \x01(\tB\x1f\xfa\x41\x1c\n\x1a\x61pi.crbug.com/UserSettings\x12:\n\tsite_role\x18\x02 \x01(\x0e\x32\".monorail.v1.UserSettings.SiteRoleB\x03\xe0\x41\x03\x12:\n\x16linked_secondary_users\x18\x03 \x03(\tB\x1a\xfa\x41\x14\n\x12\x61pi.crbug.com/User\xe0\x41\x03\x12>\n\x0bsite_access\x18\x04 \x01(\x0b\x32$.monorail.v1.UserSettings.SiteAccessB\x03\xe0\x41\x03\x12I\n\x13notification_traits\x18\x05 \x03(\x0e\x32,.monorail.v1.UserSettings.NotificationTraits\x12?\n\x0eprivacy_traits\x18\x06 \x03(\x0e\x32\'.monorail.v1.UserSettings.PrivacyTraits\x12P\n\x17site_interaction_traits\x18\x07 \x03(\x0e\x32/.monorail.v1.UserSettings.SiteInteractionTraits\x1a\x98\x01\n\nSiteAccess\x12;\n\x06status\x18\x01 \x01(\x0e\x32+.monorail.v1.UserSettings.SiteAccess.Status\x12\x0e\n\x06reason\x18\x02 \x01(\t\"=\n\x06Status\x12\x16\n\x12STATUS_UNSPECIFIED\x10\x00\x12\x0f\n\x0b\x46ULL_ACCESS\x10\x01\x12\n\n\x06\x42\x41NNED\x10\x02\"<\n\x08SiteRole\x12\x19\n\x15SITE_ROLE_UNSPECIFIED\x10\x00\x12\n\n\x06NORMAL\x10\x01\x12\t\n\x05\x41\x44MIN\x10\x02\"\xea\x01\n\x12NotificationTraits\x12#\n\x1fNOTIFICATION_TRAITS_UNSPECIFIED\x10\x00\x12\'\n#NOTIFY_ON_OWNED_OR_CC_ISSUE_CHANGES\x10\x01\x12#\n\x1fNOTIFY_ON_STARRED_ISSUE_CHANGES\x10\x02\x12\"\n\x1eNOTIFY_ON_STARRED_NOTIFY_DATES\x10\x03\x12\x18\n\x14\x43OMPACT_SUBJECT_LINE\x10\x04\x12#\n\x1fGMAIL_INCLUDE_ISSUE_LINK_BUTTON\x10\x05\"B\n\rPrivacyTraits\x12\x1e\n\x1aPRIVACY_TRAITS_UNSPECIFIED\x10\x00\x12\x11\n\rOBSCURE_EMAIL\x10\x01\"\x81\x01\n\x15SiteInteractionTraits\x12\'\n#SITE_INTERACTION_TRAITS_UNSPECIFIED\x10\x00\x12&\n\"REPORT_RESTRICT_VIEW_GOOGLE_ISSUES\x10\x01\x12\x17\n\x13PUBLIC_ISSUE_BANNER\x10\x02:7\xea\x41\x34\n\x1a\x61pi.crbug.com/UserSettings\x12\x16usersettings/{user_id}\"\xf0\x02\n\x0eUserSavedQuery\x12\x0c\n\x04name\x18\x01 \x01(\t\x12\x14\n\x0c\x64isplay_name\x18\x02 \x01(\t\x12\r\n\x05query\x18\x03 \x01(\t\x12,\n\x08projects\x18\x04 \x03(\tB\x1a\xfa\x41\x17\n\x15\x61pi.crbug.com/Project\x12G\n\x11subscription_mode\x18\x05 \x01(\x0e\x32,.monorail.v1.UserSavedQuery.SubscriptionMode\"f\n\x10SubscriptionMode\x12!\n\x1dSUBSCRIPTION_MODE_UNSPECIFIED\x10\x00\x12\x13\n\x0fNO_NOTIFICATION\x10\x01\x12\x1a\n\x16IMMEDIATE_NOTIFICATION\x10\x02:L\xea\x41I\n\x1c\x61pi.crbug.com/UserSavedQuery\x12)users/{user_id}/savedQueries/{savedQuery}\"h\n\x0bProjectStar\x12\x0c\n\x04name\x18\x01 \x01(\t:K\xea\x41H\n\x19\x61pi.crbug.com/ProjectStar\x12+users/{user_id}/projectStars/{project_name}b\x06proto3')
  ,
  dependencies=[google__proto_dot_google_dot_api_dot_resource__pb2.DESCRIPTOR,google__proto_dot_google_dot_api_dot_field__behavior__pb2.DESCRIPTOR,])



_USERSETTINGS_SITEACCESS_STATUS = _descriptor.EnumDescriptor(
  name='Status',
  full_name='monorail.v1.UserSettings.SiteAccess.Status',
  filename=None,
  file=DESCRIPTOR,
  values=[
    _descriptor.EnumValueDescriptor(
      name='STATUS_UNSPECIFIED', index=0, number=0,
      serialized_options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='FULL_ACCESS', index=1, number=1,
      serialized_options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='BANNED', index=2, number=2,
      serialized_options=None,
      type=None),
  ],
  containing_type=None,
  serialized_options=None,
  serialized_start=821,
  serialized_end=882,
)
_sym_db.RegisterEnumDescriptor(_USERSETTINGS_SITEACCESS_STATUS)

_USERSETTINGS_SITEROLE = _descriptor.EnumDescriptor(
  name='SiteRole',
  full_name='monorail.v1.UserSettings.SiteRole',
  filename=None,
  file=DESCRIPTOR,
  values=[
    _descriptor.EnumValueDescriptor(
      name='SITE_ROLE_UNSPECIFIED', index=0, number=0,
      serialized_options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='NORMAL', index=1, number=1,
      serialized_options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='ADMIN', index=2, number=2,
      serialized_options=None,
      type=None),
  ],
  containing_type=None,
  serialized_options=None,
  serialized_start=884,
  serialized_end=944,
)
_sym_db.RegisterEnumDescriptor(_USERSETTINGS_SITEROLE)

_USERSETTINGS_NOTIFICATIONTRAITS = _descriptor.EnumDescriptor(
  name='NotificationTraits',
  full_name='monorail.v1.UserSettings.NotificationTraits',
  filename=None,
  file=DESCRIPTOR,
  values=[
    _descriptor.EnumValueDescriptor(
      name='NOTIFICATION_TRAITS_UNSPECIFIED', index=0, number=0,
      serialized_options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='NOTIFY_ON_OWNED_OR_CC_ISSUE_CHANGES', index=1, number=1,
      serialized_options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='NOTIFY_ON_STARRED_ISSUE_CHANGES', index=2, number=2,
      serialized_options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='NOTIFY_ON_STARRED_NOTIFY_DATES', index=3, number=3,
      serialized_options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='COMPACT_SUBJECT_LINE', index=4, number=4,
      serialized_options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='GMAIL_INCLUDE_ISSUE_LINK_BUTTON', index=5, number=5,
      serialized_options=None,
      type=None),
  ],
  containing_type=None,
  serialized_options=None,
  serialized_start=947,
  serialized_end=1181,
)
_sym_db.RegisterEnumDescriptor(_USERSETTINGS_NOTIFICATIONTRAITS)

_USERSETTINGS_PRIVACYTRAITS = _descriptor.EnumDescriptor(
  name='PrivacyTraits',
  full_name='monorail.v1.UserSettings.PrivacyTraits',
  filename=None,
  file=DESCRIPTOR,
  values=[
    _descriptor.EnumValueDescriptor(
      name='PRIVACY_TRAITS_UNSPECIFIED', index=0, number=0,
      serialized_options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='OBSCURE_EMAIL', index=1, number=1,
      serialized_options=None,
      type=None),
  ],
  containing_type=None,
  serialized_options=None,
  serialized_start=1183,
  serialized_end=1249,
)
_sym_db.RegisterEnumDescriptor(_USERSETTINGS_PRIVACYTRAITS)

_USERSETTINGS_SITEINTERACTIONTRAITS = _descriptor.EnumDescriptor(
  name='SiteInteractionTraits',
  full_name='monorail.v1.UserSettings.SiteInteractionTraits',
  filename=None,
  file=DESCRIPTOR,
  values=[
    _descriptor.EnumValueDescriptor(
      name='SITE_INTERACTION_TRAITS_UNSPECIFIED', index=0, number=0,
      serialized_options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='REPORT_RESTRICT_VIEW_GOOGLE_ISSUES', index=1, number=1,
      serialized_options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='PUBLIC_ISSUE_BANNER', index=2, number=2,
      serialized_options=None,
      type=None),
  ],
  containing_type=None,
  serialized_options=None,
  serialized_start=1252,
  serialized_end=1381,
)
_sym_db.RegisterEnumDescriptor(_USERSETTINGS_SITEINTERACTIONTRAITS)

_USERSAVEDQUERY_SUBSCRIPTIONMODE = _descriptor.EnumDescriptor(
  name='SubscriptionMode',
  full_name='monorail.v1.UserSavedQuery.SubscriptionMode',
  filename=None,
  file=DESCRIPTOR,
  values=[
    _descriptor.EnumValueDescriptor(
      name='SUBSCRIPTION_MODE_UNSPECIFIED', index=0, number=0,
      serialized_options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='NO_NOTIFICATION', index=1, number=1,
      serialized_options=None,
      type=None),
    _descriptor.EnumValueDescriptor(
      name='IMMEDIATE_NOTIFICATION', index=2, number=2,
      serialized_options=None,
      type=None),
  ],
  containing_type=None,
  serialized_options=None,
  serialized_start=1629,
  serialized_end=1731,
)
_sym_db.RegisterEnumDescriptor(_USERSAVEDQUERY_SUBSCRIPTIONMODE)


_USER = _descriptor.Descriptor(
  name='User',
  full_name='monorail.v1.User',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='name', full_name='monorail.v1.User.name', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='display_name', full_name='monorail.v1.User.display_name', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=_b('\340A\003'), file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='availability_message', full_name='monorail.v1.User.availability_message', index=2,
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
  serialized_options=_b('\352A%\n\022api.crbug.com/User\022\017users/{user_id}'),
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=138,
  serialized_end=257,
)


_USERSETTINGS_SITEACCESS = _descriptor.Descriptor(
  name='SiteAccess',
  full_name='monorail.v1.UserSettings.SiteAccess',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='status', full_name='monorail.v1.UserSettings.SiteAccess.status', index=0,
      number=1, type=14, cpp_type=8, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='reason', full_name='monorail.v1.UserSettings.SiteAccess.reason', index=1,
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
    _USERSETTINGS_SITEACCESS_STATUS,
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=730,
  serialized_end=882,
)

_USERSETTINGS = _descriptor.Descriptor(
  name='UserSettings',
  full_name='monorail.v1.UserSettings',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='name', full_name='monorail.v1.UserSettings.name', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=_b('\372A\034\n\032api.crbug.com/UserSettings'), file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='site_role', full_name='monorail.v1.UserSettings.site_role', index=1,
      number=2, type=14, cpp_type=8, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=_b('\340A\003'), file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='linked_secondary_users', full_name='monorail.v1.UserSettings.linked_secondary_users', index=2,
      number=3, type=9, cpp_type=9, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=_b('\372A\024\n\022api.crbug.com/User\340A\003'), file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='site_access', full_name='monorail.v1.UserSettings.site_access', index=3,
      number=4, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=_b('\340A\003'), file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='notification_traits', full_name='monorail.v1.UserSettings.notification_traits', index=4,
      number=5, type=14, cpp_type=8, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='privacy_traits', full_name='monorail.v1.UserSettings.privacy_traits', index=5,
      number=6, type=14, cpp_type=8, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='site_interaction_traits', full_name='monorail.v1.UserSettings.site_interaction_traits', index=6,
      number=7, type=14, cpp_type=8, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[_USERSETTINGS_SITEACCESS, ],
  enum_types=[
    _USERSETTINGS_SITEROLE,
    _USERSETTINGS_NOTIFICATIONTRAITS,
    _USERSETTINGS_PRIVACYTRAITS,
    _USERSETTINGS_SITEINTERACTIONTRAITS,
  ],
  serialized_options=_b('\352A4\n\032api.crbug.com/UserSettings\022\026usersettings/{user_id}'),
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=260,
  serialized_end=1438,
)


_USERSAVEDQUERY = _descriptor.Descriptor(
  name='UserSavedQuery',
  full_name='monorail.v1.UserSavedQuery',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='name', full_name='monorail.v1.UserSavedQuery.name', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='display_name', full_name='monorail.v1.UserSavedQuery.display_name', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='query', full_name='monorail.v1.UserSavedQuery.query', index=2,
      number=3, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='projects', full_name='monorail.v1.UserSavedQuery.projects', index=3,
      number=4, type=9, cpp_type=9, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=_b('\372A\027\n\025api.crbug.com/Project'), file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='subscription_mode', full_name='monorail.v1.UserSavedQuery.subscription_mode', index=4,
      number=5, type=14, cpp_type=8, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
    _USERSAVEDQUERY_SUBSCRIPTIONMODE,
  ],
  serialized_options=_b('\352AI\n\034api.crbug.com/UserSavedQuery\022)users/{user_id}/savedQueries/{savedQuery}'),
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=1441,
  serialized_end=1809,
)


_PROJECTSTAR = _descriptor.Descriptor(
  name='ProjectStar',
  full_name='monorail.v1.ProjectStar',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='name', full_name='monorail.v1.ProjectStar.name', index=0,
      number=1, type=9, cpp_type=9, label=1,
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
  serialized_options=_b('\352AH\n\031api.crbug.com/ProjectStar\022+users/{user_id}/projectStars/{project_name}'),
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=1811,
  serialized_end=1915,
)

_USERSETTINGS_SITEACCESS.fields_by_name['status'].enum_type = _USERSETTINGS_SITEACCESS_STATUS
_USERSETTINGS_SITEACCESS.containing_type = _USERSETTINGS
_USERSETTINGS_SITEACCESS_STATUS.containing_type = _USERSETTINGS_SITEACCESS
_USERSETTINGS.fields_by_name['site_role'].enum_type = _USERSETTINGS_SITEROLE
_USERSETTINGS.fields_by_name['site_access'].message_type = _USERSETTINGS_SITEACCESS
_USERSETTINGS.fields_by_name['notification_traits'].enum_type = _USERSETTINGS_NOTIFICATIONTRAITS
_USERSETTINGS.fields_by_name['privacy_traits'].enum_type = _USERSETTINGS_PRIVACYTRAITS
_USERSETTINGS.fields_by_name['site_interaction_traits'].enum_type = _USERSETTINGS_SITEINTERACTIONTRAITS
_USERSETTINGS_SITEROLE.containing_type = _USERSETTINGS
_USERSETTINGS_NOTIFICATIONTRAITS.containing_type = _USERSETTINGS
_USERSETTINGS_PRIVACYTRAITS.containing_type = _USERSETTINGS
_USERSETTINGS_SITEINTERACTIONTRAITS.containing_type = _USERSETTINGS
_USERSAVEDQUERY.fields_by_name['subscription_mode'].enum_type = _USERSAVEDQUERY_SUBSCRIPTIONMODE
_USERSAVEDQUERY_SUBSCRIPTIONMODE.containing_type = _USERSAVEDQUERY
DESCRIPTOR.message_types_by_name['User'] = _USER
DESCRIPTOR.message_types_by_name['UserSettings'] = _USERSETTINGS
DESCRIPTOR.message_types_by_name['UserSavedQuery'] = _USERSAVEDQUERY
DESCRIPTOR.message_types_by_name['ProjectStar'] = _PROJECTSTAR
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

User = _reflection.GeneratedProtocolMessageType('User', (_message.Message,), dict(
  DESCRIPTOR = _USER,
  __module__ = 'api.v1.api_proto.user_objects_pb2'
  # @@protoc_insertion_point(class_scope:monorail.v1.User)
  ))
_sym_db.RegisterMessage(User)

UserSettings = _reflection.GeneratedProtocolMessageType('UserSettings', (_message.Message,), dict(

  SiteAccess = _reflection.GeneratedProtocolMessageType('SiteAccess', (_message.Message,), dict(
    DESCRIPTOR = _USERSETTINGS_SITEACCESS,
    __module__ = 'api.v1.api_proto.user_objects_pb2'
    # @@protoc_insertion_point(class_scope:monorail.v1.UserSettings.SiteAccess)
    ))
  ,
  DESCRIPTOR = _USERSETTINGS,
  __module__ = 'api.v1.api_proto.user_objects_pb2'
  # @@protoc_insertion_point(class_scope:monorail.v1.UserSettings)
  ))
_sym_db.RegisterMessage(UserSettings)
_sym_db.RegisterMessage(UserSettings.SiteAccess)

UserSavedQuery = _reflection.GeneratedProtocolMessageType('UserSavedQuery', (_message.Message,), dict(
  DESCRIPTOR = _USERSAVEDQUERY,
  __module__ = 'api.v1.api_proto.user_objects_pb2'
  # @@protoc_insertion_point(class_scope:monorail.v1.UserSavedQuery)
  ))
_sym_db.RegisterMessage(UserSavedQuery)

ProjectStar = _reflection.GeneratedProtocolMessageType('ProjectStar', (_message.Message,), dict(
  DESCRIPTOR = _PROJECTSTAR,
  __module__ = 'api.v1.api_proto.user_objects_pb2'
  # @@protoc_insertion_point(class_scope:monorail.v1.ProjectStar)
  ))
_sym_db.RegisterMessage(ProjectStar)


_USER.fields_by_name['display_name']._options = None
_USER._options = None
_USERSETTINGS.fields_by_name['name']._options = None
_USERSETTINGS.fields_by_name['site_role']._options = None
_USERSETTINGS.fields_by_name['linked_secondary_users']._options = None
_USERSETTINGS.fields_by_name['site_access']._options = None
_USERSETTINGS._options = None
_USERSAVEDQUERY.fields_by_name['projects']._options = None
_USERSAVEDQUERY._options = None
_PROJECTSTAR._options = None
# @@protoc_insertion_point(module_scope)
