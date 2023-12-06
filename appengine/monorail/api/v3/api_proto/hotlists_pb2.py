# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: api/v3/api_proto/hotlists.proto
"""Generated protocol buffer code."""
from google.protobuf.internal import builder as _builder
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from api.v3.api_proto import feature_objects_pb2 as api_dot_v3_dot_api__proto_dot_feature__objects__pb2
from google.protobuf import field_mask_pb2 as google_dot_protobuf_dot_field__mask__pb2
from google.protobuf import empty_pb2 as google_dot_protobuf_dot_empty__pb2
from google.api import field_behavior_pb2 as google_dot_api_dot_field__behavior__pb2
from google.api import resource_pb2 as google_dot_api_dot_resource__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x1f\x61pi/v3/api_proto/hotlists.proto\x12\x0bmonorail.v3\x1a&api/v3/api_proto/feature_objects.proto\x1a google/protobuf/field_mask.proto\x1a\x1bgoogle/protobuf/empty.proto\x1a\x1fgoogle/api/field_behavior.proto\x1a\x19google/api/resource.proto\"B\n\x14\x43reateHotlistRequest\x12*\n\x07hotlist\x18\x01 \x01(\x0b\x32\x14.monorail.v3.HotlistB\x03\xe0\x41\x02\"@\n\x11GetHotlistRequest\x12+\n\x04name\x18\x01 \x01(\tB\x1d\xe0\x41\x02\xfa\x41\x17\n\x15\x61pi.crbug.com/Hotlist\"\x92\x01\n\x14UpdateHotlistRequest\x12\x44\n\x07hotlist\x18\x01 \x01(\x0b\x32\x14.monorail.v3.HotlistB\x1d\xe0\x41\x02\xfa\x41\x17\n\x15\x61pi.crbug.com/Hotlist\x12\x34\n\x0bupdate_mask\x18\x02 \x01(\x0b\x32\x1a.google.protobuf.FieldMaskB\x03\xe0\x41\x02\"\x81\x01\n\x17ListHotlistItemsRequest\x12-\n\x06parent\x18\x01 \x01(\tB\x1d\xe0\x41\x02\xfa\x41\x17\n\x15\x61pi.crbug.com/Hotlist\x12\x11\n\tpage_size\x18\x02 \x01(\x05\x12\x10\n\x08order_by\x18\x03 \x01(\t\x12\x12\n\npage_token\x18\x04 \x01(\t\"\\\n\x18ListHotlistItemsResponse\x12\'\n\x05items\x18\x01 \x03(\x0b\x32\x18.monorail.v3.HotlistItem\x12\x17\n\x0fnext_page_token\x18\x02 \x01(\t\"\xa0\x01\n\x19RerankHotlistItemsRequest\x12+\n\x04name\x18\x01 \x01(\tB\x1d\xfa\x41\x17\n\x15\x61pi.crbug.com/Hotlist\xe0\x41\x02\x12\x38\n\rhotlist_items\x18\x02 \x03(\tB!\xfa\x41\x1b\n\x19\x61pi.crbug.com/HotlistItem\xe0\x41\x02\x12\x1c\n\x0ftarget_position\x18\x03 \x01(\rB\x03\xe0\x41\x02\"\x8d\x01\n\x16\x41\x64\x64HotlistItemsRequest\x12-\n\x06parent\x18\x01 \x01(\tB\x1d\xe0\x41\x02\xfa\x41\x17\n\x15\x61pi.crbug.com/Hotlist\x12+\n\x06issues\x18\x02 \x03(\tB\x1b\xe0\x41\x02\xfa\x41\x15\n\x13\x61pi.crbug.com/Issue\x12\x17\n\x0ftarget_position\x18\x03 \x01(\r\"w\n\x19RemoveHotlistItemsRequest\x12-\n\x06parent\x18\x01 \x01(\tB\x1d\xe0\x41\x02\xfa\x41\x17\n\x15\x61pi.crbug.com/Hotlist\x12+\n\x06issues\x18\x02 \x03(\tB\x1b\xe0\x41\x02\xfa\x41\x15\n\x13\x61pi.crbug.com/Issue\"w\n\x1bRemoveHotlistEditorsRequest\x12+\n\x04name\x18\x01 \x01(\tB\x1d\xe0\x41\x02\xfa\x41\x17\n\x15\x61pi.crbug.com/Hotlist\x12+\n\x07\x65\x64itors\x18\x02 \x03(\tB\x1a\xe0\x41\x02\xfa\x41\x14\n\x12\x61pi.crbug.com/User\"H\n\x1cGatherHotlistsForUserRequest\x12(\n\x04user\x18\x01 \x01(\tB\x1a\xe0\x41\x02\xfa\x41\x14\n\x12\x61pi.crbug.com/User\"G\n\x1dGatherHotlistsForUserResponse\x12&\n\x08hotlists\x18\x01 \x03(\x0b\x32\x14.monorail.v3.Hotlist2\xe6\x06\n\x08Hotlists\x12J\n\rCreateHotlist\x12!.monorail.v3.CreateHotlistRequest\x1a\x14.monorail.v3.Hotlist\"\x00\x12\x44\n\nGetHotlist\x12\x1e.monorail.v3.GetHotlistRequest\x1a\x14.monorail.v3.Hotlist\"\x00\x12J\n\rUpdateHotlist\x12!.monorail.v3.UpdateHotlistRequest\x1a\x14.monorail.v3.Hotlist\"\x00\x12I\n\rDeleteHotlist\x12\x1e.monorail.v3.GetHotlistRequest\x1a\x16.google.protobuf.Empty\"\x00\x12\x61\n\x10ListHotlistItems\x12$.monorail.v3.ListHotlistItemsRequest\x1a%.monorail.v3.ListHotlistItemsResponse\"\x00\x12V\n\x12RerankHotlistItems\x12&.monorail.v3.RerankHotlistItemsRequest\x1a\x16.google.protobuf.Empty\"\x00\x12P\n\x0f\x41\x64\x64HotlistItems\x12#.monorail.v3.AddHotlistItemsRequest\x1a\x16.google.protobuf.Empty\"\x00\x12V\n\x12RemoveHotlistItems\x12&.monorail.v3.RemoveHotlistItemsRequest\x1a\x16.google.protobuf.Empty\"\x00\x12Z\n\x14RemoveHotlistEditors\x12(.monorail.v3.RemoveHotlistEditorsRequest\x1a\x16.google.protobuf.Empty\"\x00\x12p\n\x15GatherHotlistsForUser\x12).monorail.v3.GatherHotlistsForUserRequest\x1a*.monorail.v3.GatherHotlistsForUserResponse\"\x00\x42#Z!infra/monorailv2/api/v3/api_protob\x06proto3')

_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, globals())
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'api.v3.api_proto.hotlists_pb2', globals())
if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'Z!infra/monorailv2/api/v3/api_proto'
  _CREATEHOTLISTREQUEST.fields_by_name['hotlist']._options = None
  _CREATEHOTLISTREQUEST.fields_by_name['hotlist']._serialized_options = b'\340A\002'
  _GETHOTLISTREQUEST.fields_by_name['name']._options = None
  _GETHOTLISTREQUEST.fields_by_name['name']._serialized_options = b'\340A\002\372A\027\n\025api.crbug.com/Hotlist'
  _UPDATEHOTLISTREQUEST.fields_by_name['hotlist']._options = None
  _UPDATEHOTLISTREQUEST.fields_by_name['hotlist']._serialized_options = b'\340A\002\372A\027\n\025api.crbug.com/Hotlist'
  _UPDATEHOTLISTREQUEST.fields_by_name['update_mask']._options = None
  _UPDATEHOTLISTREQUEST.fields_by_name['update_mask']._serialized_options = b'\340A\002'
  _LISTHOTLISTITEMSREQUEST.fields_by_name['parent']._options = None
  _LISTHOTLISTITEMSREQUEST.fields_by_name['parent']._serialized_options = b'\340A\002\372A\027\n\025api.crbug.com/Hotlist'
  _RERANKHOTLISTITEMSREQUEST.fields_by_name['name']._options = None
  _RERANKHOTLISTITEMSREQUEST.fields_by_name['name']._serialized_options = b'\372A\027\n\025api.crbug.com/Hotlist\340A\002'
  _RERANKHOTLISTITEMSREQUEST.fields_by_name['hotlist_items']._options = None
  _RERANKHOTLISTITEMSREQUEST.fields_by_name['hotlist_items']._serialized_options = b'\372A\033\n\031api.crbug.com/HotlistItem\340A\002'
  _RERANKHOTLISTITEMSREQUEST.fields_by_name['target_position']._options = None
  _RERANKHOTLISTITEMSREQUEST.fields_by_name['target_position']._serialized_options = b'\340A\002'
  _ADDHOTLISTITEMSREQUEST.fields_by_name['parent']._options = None
  _ADDHOTLISTITEMSREQUEST.fields_by_name['parent']._serialized_options = b'\340A\002\372A\027\n\025api.crbug.com/Hotlist'
  _ADDHOTLISTITEMSREQUEST.fields_by_name['issues']._options = None
  _ADDHOTLISTITEMSREQUEST.fields_by_name['issues']._serialized_options = b'\340A\002\372A\025\n\023api.crbug.com/Issue'
  _REMOVEHOTLISTITEMSREQUEST.fields_by_name['parent']._options = None
  _REMOVEHOTLISTITEMSREQUEST.fields_by_name['parent']._serialized_options = b'\340A\002\372A\027\n\025api.crbug.com/Hotlist'
  _REMOVEHOTLISTITEMSREQUEST.fields_by_name['issues']._options = None
  _REMOVEHOTLISTITEMSREQUEST.fields_by_name['issues']._serialized_options = b'\340A\002\372A\025\n\023api.crbug.com/Issue'
  _REMOVEHOTLISTEDITORSREQUEST.fields_by_name['name']._options = None
  _REMOVEHOTLISTEDITORSREQUEST.fields_by_name['name']._serialized_options = b'\340A\002\372A\027\n\025api.crbug.com/Hotlist'
  _REMOVEHOTLISTEDITORSREQUEST.fields_by_name['editors']._options = None
  _REMOVEHOTLISTEDITORSREQUEST.fields_by_name['editors']._serialized_options = b'\340A\002\372A\024\n\022api.crbug.com/User'
  _GATHERHOTLISTSFORUSERREQUEST.fields_by_name['user']._options = None
  _GATHERHOTLISTSFORUSERREQUEST.fields_by_name['user']._serialized_options = b'\340A\002\372A\024\n\022api.crbug.com/User'
  _CREATEHOTLISTREQUEST._serialized_start=211
  _CREATEHOTLISTREQUEST._serialized_end=277
  _GETHOTLISTREQUEST._serialized_start=279
  _GETHOTLISTREQUEST._serialized_end=343
  _UPDATEHOTLISTREQUEST._serialized_start=346
  _UPDATEHOTLISTREQUEST._serialized_end=492
  _LISTHOTLISTITEMSREQUEST._serialized_start=495
  _LISTHOTLISTITEMSREQUEST._serialized_end=624
  _LISTHOTLISTITEMSRESPONSE._serialized_start=626
  _LISTHOTLISTITEMSRESPONSE._serialized_end=718
  _RERANKHOTLISTITEMSREQUEST._serialized_start=721
  _RERANKHOTLISTITEMSREQUEST._serialized_end=881
  _ADDHOTLISTITEMSREQUEST._serialized_start=884
  _ADDHOTLISTITEMSREQUEST._serialized_end=1025
  _REMOVEHOTLISTITEMSREQUEST._serialized_start=1027
  _REMOVEHOTLISTITEMSREQUEST._serialized_end=1146
  _REMOVEHOTLISTEDITORSREQUEST._serialized_start=1148
  _REMOVEHOTLISTEDITORSREQUEST._serialized_end=1267
  _GATHERHOTLISTSFORUSERREQUEST._serialized_start=1269
  _GATHERHOTLISTSFORUSERREQUEST._serialized_end=1341
  _GATHERHOTLISTSFORUSERRESPONSE._serialized_start=1343
  _GATHERHOTLISTSFORUSERRESPONSE._serialized_end=1414
  _HOTLISTS._serialized_start=1417
  _HOTLISTS._serialized_end=2287
# @@protoc_insertion_point(module_scope)
