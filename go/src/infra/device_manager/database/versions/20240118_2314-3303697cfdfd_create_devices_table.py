# Copyright 2024 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""create Devices table

Revision ID: 3303697cfdfd
Revises:
Create Date: 2024-01-18 23:14:02.705448

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa
from sqlalchemy.dialects.postgresql import JSONB

# revision identifiers, used by Alembic.
revision: str = '3303697cfdfd'
down_revision: Union[str, None] = None
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
  op.create_table(
      'Devices',
      sa.Column('id', sa.String, primary_key=True),
      sa.Column('device_address', sa.String, nullable=False),
      sa.Column('device_type', sa.String, nullable=False),
      sa.Column('device_state', sa.String, nullable=False),
      sa.Column('schedulable_labels', JSONB),
  )


def downgrade() -> None:
  op.drop_table('Devices')
