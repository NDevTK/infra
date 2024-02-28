# Copyright 2024 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""create DeviceLeaseRecords table

Revision ID: 8b7c9cfc4c56
Revises: 3303697cfdfd
Create Date: 2024-01-18 23:15:57.591408

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa
from sqlalchemy.dialects.postgresql import JSONB

# revision identifiers, used by Alembic.
revision: str = '8b7c9cfc4c56'
down_revision: Union[str, None] = '3303697cfdfd'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
  op.create_table(
      'DeviceLeaseRecords',
      sa.Column('id', sa.UUID(as_uuid=True), primary_key=True),
      sa.Column('idempotency_key', sa.UUID(as_uuid=True), nullable=False),
      sa.Column('device_id', sa.String, nullable=False),
      sa.Column('device_address', sa.String, nullable=False),
      sa.Column('device_type', sa.String, nullable=False),
      sa.Column('owner_id', sa.String, nullable=False),
      sa.Column('leased_time', sa.DateTime(), nullable=False),
      sa.Column('released_time', sa.DateTime()),
      sa.Column('expiration_time', sa.DateTime()),
      sa.Column('last_updated_time', sa.DateTime(), nullable=False),
      sa.Column('request_parameters', JSONB),
  )


def downgrade() -> None:
  op.drop_table('DeviceLeaseRecords')
