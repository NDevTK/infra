# Monorail API v1

Monorail API v1 aims to provide nearly identical interface to Google Code issue tracker's API for existing clients' smooth transition. You can get a high-level overview from the documents below.

* [One-page getting started](https://docs.google.com/document/d/1Loz3NIqpTrKWLR5rV3tNGLl-7Rn9ZyrmG9MFOft35Pc)
* [Code example in python](query_issues.py)
* [Design doc](https://docs.google.com/document/d/1FcmDVP5PwlMHi3ozi98lgK1E-ZU7WWrLYN8g9KIwnbU)


In details, API provides the following methods to read/write user/issue/comment objects in Monorail:

[TOC]

## monorail.groups.create

* Description: Create a new user group.
* Permission: The requester has permission to create groups.
* Parameters:
  * groupName(required, string): The name of the group to create.
  * who_can_view_members(required, string): The visibility setting of the group. Available options are 'ANYONE', 'MEMBERS' and 'OWNERS'.
  * ext_group_type(optional, string): The type of the source group if the new group is imported from the source. Available options are 'BAGGINS', 'CHROME_INFRA_AUTH' and 'MDB'.
* Return message:
  * groupID(int): The ID of the newly created group. 
* Error code:
  * 403: The requester has no permission to create a group.

## monorail.groups.get

* Description: Get a group's settings and users.
* Permission: The requester has permission to vuew this group.
* Parameters:
  - groupName(required, string): The name of the group to view.
* Return message:
  - groupID(int): The ID of the newly created group.
  - groupSettings(dict): 
* Error code:
  - 403: The requester has no permission to view this group.