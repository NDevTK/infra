# Copyright 2024 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""add created_time column to Devices table

Revision ID: 0034a4e9d75d
Revises: 469ac7ced828
Create Date: 2024-05-10 19:09:11.682234

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa
from sqlalchemy.sql import func


# revision identifiers, used by Alembic.
revision: str = '0034a4e9d75d'
down_revision: Union[str, None] = '469ac7ced828'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
  op.add_column(
    "Devices",
    sa.Column("created_time", sa.DateTime(), server_default=func.now())
  )


def downgrade() -> None:
  op.drop_column("Devices", "created_time")
