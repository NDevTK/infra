# Copyright 2024 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""add columns for notification times

Revision ID: b56ba79423b3
Revises: 0034a4e9d75d
Create Date: 2024-05-28 14:34:33.779920

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision: str = 'b56ba79423b3'
down_revision: Union[str, None] = '0034a4e9d75d'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    op.add_column(
        "Devices",
        sa.Column("last_notification_time", sa.DateTime(), server_default=None),
    )


def downgrade() -> None:
    op.drop_column(
        "Devices",
        "last_notification_time",
    )
