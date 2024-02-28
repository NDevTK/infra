# Copyright 2024 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""create indices for device columns

Revision ID: c10375863126
Revises: 8b7c9cfc4c56
Create Date: 2024-01-18 23:29:33.182801

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa

# revision identifiers, used by Alembic.
revision: str = 'c10375863126'
down_revision: Union[str, None] = '8b7c9cfc4c56'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
  # Devices indices
  op.create_index(
      'Devices_device_type', 'Devices', ['device_type'], unique=False)

  # DeviceLeaseRecords indices
  op.create_index(
      'DeviceLeaseRecords_device_id',
      'DeviceLeaseRecords', ['device_id'],
      unique=False)
  op.create_index(
      'DeviceLeaseRecords_device_type',
      'DeviceLeaseRecords', ['device_type'],
      unique=False)


def downgrade() -> None:
  # Devices indices
  op.drop_index('Devices_device_type', 'Devices')

  # DeviceLeaseRecords indices
  op.drop_index('DeviceLeaseRecords_device_id', 'DeviceLeaseRecords')
  op.drop_index('DeviceLeaseRecords_device_type', 'DeviceLeaseRecords')
