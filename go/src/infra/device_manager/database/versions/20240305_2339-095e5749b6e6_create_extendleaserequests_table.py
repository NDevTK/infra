# Copyright 2024 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""create ExtendLeaseRequests table

Revision ID: 095e5749b6e6
Revises: c10375863126
Create Date: 2024-03-05 23:39:38.191497

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision: str = '095e5749b6e6'
down_revision: Union[str, None] = 'c10375863126'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
  op.create_table(
      'ExtendLeaseRequests',
      sa.Column('id', sa.UUID(as_uuid=True), primary_key=True),
      sa.Column('lease_id', sa.UUID(as_uuid=True), nullable=False),
      sa.Column('idempotency_key', sa.UUID(as_uuid=True), nullable=False),
      sa.Column('extend_duration', sa.Integer, nullable=False),
      sa.Column('request_time', sa.DateTime()),
      sa.Column('expiration_time', sa.DateTime()),
  )


def downgrade() -> None:
  op.drop_table('ExtendLeaseRequests')
