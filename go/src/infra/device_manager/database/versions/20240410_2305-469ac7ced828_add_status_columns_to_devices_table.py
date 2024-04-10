# Copyright 2024 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""add status columns to Devices table

Revision ID: 469ac7ced828
Revises: 095e5749b6e6
Create Date: 2024-04-10 23:05:23.341893

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa

# revision identifiers, used by Alembic.
revision: str = '469ac7ced828'
down_revision: Union[str, None] = '095e5749b6e6'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
  op.add_column("Devices", sa.Column("last_update_time", sa.DateTime()))
  op.add_column("Devices", sa.Column("is_active", sa.Boolean))


def downgrade() -> None:
  op.drop_column("Devices", "last_update_time")
  op.drop_column("Devices", "is_active")
