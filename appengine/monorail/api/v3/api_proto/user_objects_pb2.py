# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: api/v3/api_proto/user_objects.proto
"""Generated protocol buffer code."""
from google.protobuf.internal import builder as _builder
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google.api import resource_pb2 as google_dot_api_dot_resource__pb2
from google.api import field_behavior_pb2 as google_dot_api_dot_field__behavior__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n#api/v3/api_proto/user_objects.proto\x12\x0bmonorail.v3\x1a\x19google/api/resource.proto\x1a\x1fgoogle/api/field_behavior.proto\"\xa4\x01\n\x04User\x12\x0c\n\x04name\x18\x01 \x01(\t\x12\x14\n\x0c\x64isplay_name\x18\x02 \x01(\t\x12\x12\n\x05\x65mail\x18\x04 \x01(\tB\x03\xe0\x41\x03\x12\x1c\n\x14\x61vailability_message\x18\x03 \x01(\t\x12\x1c\n\x14last_visit_timestamp\x18\x05 \x01(\x05:(\xea\x41%\n\x12\x61pi.crbug.com/User\x12\x0fusers/{user_id}\"\x9a\t\n\x0cUserSettings\x12-\n\x04name\x18\x01 \x01(\tB\x1f\xfa\x41\x1c\n\x1a\x61pi.crbug.com/UserSettings\x12:\n\tsite_role\x18\x02 \x01(\x0e\x32\".monorail.v3.UserSettings.SiteRoleB\x03\xe0\x41\x03\x12:\n\x16linked_secondary_users\x18\x03 \x03(\tB\x1a\xfa\x41\x14\n\x12\x61pi.crbug.com/User\xe0\x41\x03\x12>\n\x0bsite_access\x18\x04 \x01(\x0b\x32$.monorail.v3.UserSettings.SiteAccessB\x03\xe0\x41\x03\x12I\n\x13notification_traits\x18\x05 \x03(\x0e\x32,.monorail.v3.UserSettings.NotificationTraits\x12?\n\x0eprivacy_traits\x18\x06 \x03(\x0e\x32\'.monorail.v3.UserSettings.PrivacyTraits\x12P\n\x17site_interaction_traits\x18\x07 \x03(\x0e\x32/.monorail.v3.UserSettings.SiteInteractionTraits\x1a\x98\x01\n\nSiteAccess\x12;\n\x06status\x18\x01 \x01(\x0e\x32+.monorail.v3.UserSettings.SiteAccess.Status\x12\x0e\n\x06reason\x18\x02 \x01(\t\"=\n\x06Status\x12\x16\n\x12STATUS_UNSPECIFIED\x10\x00\x12\x0f\n\x0b\x46ULL_ACCESS\x10\x01\x12\n\n\x06\x42\x41NNED\x10\x02\"<\n\x08SiteRole\x12\x19\n\x15SITE_ROLE_UNSPECIFIED\x10\x00\x12\n\n\x06NORMAL\x10\x01\x12\t\n\x05\x41\x44MIN\x10\x02\"\xea\x01\n\x12NotificationTraits\x12#\n\x1fNOTIFICATION_TRAITS_UNSPECIFIED\x10\x00\x12\'\n#NOTIFY_ON_OWNED_OR_CC_ISSUE_CHANGES\x10\x01\x12#\n\x1fNOTIFY_ON_STARRED_ISSUE_CHANGES\x10\x02\x12\"\n\x1eNOTIFY_ON_STARRED_NOTIFY_DATES\x10\x03\x12\x18\n\x14\x43OMPACT_SUBJECT_LINE\x10\x04\x12#\n\x1fGMAIL_INCLUDE_ISSUE_LINK_BUTTON\x10\x05\"B\n\rPrivacyTraits\x12\x1e\n\x1aPRIVACY_TRAITS_UNSPECIFIED\x10\x00\x12\x11\n\rOBSCURE_EMAIL\x10\x01\"\x81\x01\n\x15SiteInteractionTraits\x12\'\n#SITE_INTERACTION_TRAITS_UNSPECIFIED\x10\x00\x12&\n\"REPORT_RESTRICT_VIEW_GOOGLE_ISSUES\x10\x01\x12\x17\n\x13PUBLIC_ISSUE_BANNER\x10\x02:7\xea\x41\x34\n\x1a\x61pi.crbug.com/UserSettings\x12\x16usersettings/{user_id}\"\xf4\x02\n\x0eUserSavedQuery\x12\x0c\n\x04name\x18\x01 \x01(\t\x12\x14\n\x0c\x64isplay_name\x18\x02 \x01(\t\x12\r\n\x05query\x18\x03 \x01(\t\x12,\n\x08projects\x18\x04 \x03(\tB\x1a\xfa\x41\x17\n\x15\x61pi.crbug.com/Project\x12G\n\x11subscription_mode\x18\x05 \x01(\x0e\x32,.monorail.v3.UserSavedQuery.SubscriptionMode\"f\n\x10SubscriptionMode\x12!\n\x1dSUBSCRIPTION_MODE_UNSPECIFIED\x10\x00\x12\x13\n\x0fNO_NOTIFICATION\x10\x01\x12\x1a\n\x16IMMEDIATE_NOTIFICATION\x10\x02:P\xea\x41M\n\x1c\x61pi.crbug.com/UserSavedQuery\x12-users/{user_id}/savedQueries/{saved_query_id}\"h\n\x0bProjectStar\x12\x0c\n\x04name\x18\x01 \x01(\t:K\xea\x41H\n\x19\x61pi.crbug.com/ProjectStar\x12+users/{user_id}/projectStars/{project_name}B#Z!infra/monorailv2/api/v3/api_protob\x06proto3')

_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, globals())
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'api.v3.api_proto.user_objects_pb2', globals())
if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'Z!infra/monorailv2/api/v3/api_proto'
  _USER.fields_by_name['email']._options = None
  _USER.fields_by_name['email']._serialized_options = b'\340A\003'
  _USER._options = None
  _USER._serialized_options = b'\352A%\n\022api.crbug.com/User\022\017users/{user_id}'
  _USERSETTINGS.fields_by_name['name']._options = None
  _USERSETTINGS.fields_by_name['name']._serialized_options = b'\372A\034\n\032api.crbug.com/UserSettings'
  _USERSETTINGS.fields_by_name['site_role']._options = None
  _USERSETTINGS.fields_by_name['site_role']._serialized_options = b'\340A\003'
  _USERSETTINGS.fields_by_name['linked_secondary_users']._options = None
  _USERSETTINGS.fields_by_name['linked_secondary_users']._serialized_options = b'\372A\024\n\022api.crbug.com/User\340A\003'
  _USERSETTINGS.fields_by_name['site_access']._options = None
  _USERSETTINGS.fields_by_name['site_access']._serialized_options = b'\340A\003'
  _USERSETTINGS._options = None
  _USERSETTINGS._serialized_options = b'\352A4\n\032api.crbug.com/UserSettings\022\026usersettings/{user_id}'
  _USERSAVEDQUERY.fields_by_name['projects']._options = None
  _USERSAVEDQUERY.fields_by_name['projects']._serialized_options = b'\372A\027\n\025api.crbug.com/Project'
  _USERSAVEDQUERY._options = None
  _USERSAVEDQUERY._serialized_options = b'\352AM\n\034api.crbug.com/UserSavedQuery\022-users/{user_id}/savedQueries/{saved_query_id}'
  _PROJECTSTAR._options = None
  _PROJECTSTAR._serialized_options = b'\352AH\n\031api.crbug.com/ProjectStar\022+users/{user_id}/projectStars/{project_name}'
  _USER._serialized_start=113
  _USER._serialized_end=277
  _USERSETTINGS._serialized_start=280
  _USERSETTINGS._serialized_end=1458
  _USERSETTINGS_SITEACCESS._serialized_start=750
  _USERSETTINGS_SITEACCESS._serialized_end=902
  _USERSETTINGS_SITEACCESS_STATUS._serialized_start=841
  _USERSETTINGS_SITEACCESS_STATUS._serialized_end=902
  _USERSETTINGS_SITEROLE._serialized_start=904
  _USERSETTINGS_SITEROLE._serialized_end=964
  _USERSETTINGS_NOTIFICATIONTRAITS._serialized_start=967
  _USERSETTINGS_NOTIFICATIONTRAITS._serialized_end=1201
  _USERSETTINGS_PRIVACYTRAITS._serialized_start=1203
  _USERSETTINGS_PRIVACYTRAITS._serialized_end=1269
  _USERSETTINGS_SITEINTERACTIONTRAITS._serialized_start=1272
  _USERSETTINGS_SITEINTERACTIONTRAITS._serialized_end=1401
  _USERSAVEDQUERY._serialized_start=1461
  _USERSAVEDQUERY._serialized_end=1833
  _USERSAVEDQUERY_SUBSCRIPTIONMODE._serialized_start=1649
  _USERSAVEDQUERY_SUBSCRIPTIONMODE._serialized_end=1751
  _PROJECTSTAR._serialized_start=1835
  _PROJECTSTAR._serialized_end=1939
# @@protoc_insertion_point(module_scope)
