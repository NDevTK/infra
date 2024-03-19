# Alembic DB Management

[TOC]

## Prerequisites

Set up the Python virtualenv and install Alembic:

```bash
python -m venv dev-env

source dev-env/bin/activate

pip install -r requirements.txt
```

## Development

See main README for development DB setup.

## Staging / Production

To manage the Staging and Production AlloyDB instances, you must have GCP
permissions to access the GCE instance used to manage the cluster. Currently,
only Device Manager admins will have access. The following instructions cover
staging management. You can replace the host URIs to manage production.

### 1. Tunnel to AlloyDB Bastion

If granted permission to the bastion, create a tunnel to staging with this
command in a separate window. No `dev-env` is required:

```bash
ssh -L localhost:5433:10.40.128.2:5432 \
  -i ~/.ssh/google_compute_engine \
  ${USERNAME}_google_com@nic0.alloydb-bastion.us-central1-a.c.fleet-device-manager-dev.internal.gcpnode.com
```

This creates a tunnel to the bastion and maps the DB port `5432` to the your
`localhost:5433` to avoid collisions with the development setup.

### 2. Run Alembic Migrations

To run Alembic against the staging environment, change the values of the
`alloydb.dev` section in `database/db_config.ini`. In a different window with
`dev-env` activated, run the following to check if a connection can be
established:

```bash
# check current alembic version on db
(dev-env) $ ALEMBIC_ENV=alloydb.dev alembic current

INFO  [alembic.runtime.migration] Context impl PostgresqlImpl.
INFO  [alembic.runtime.migration] Will assume transactional DDL.
095e5749b6e6 (head)  # this hash points to an alembic revision
```

Once confirmed, you can run the migrations as needed:

```bash
# upgrading schema to current version
(dev-env) $ ALEMBIC_ENV=alloydb.dev alembic upgrade head

INFO  [alembic.runtime.migration] Context impl PostgresqlImpl.
INFO  [alembic.runtime.migration] Will assume transactional DDL.
INFO  [alembic.runtime.migration] Running upgrade  -> 3303697cfdfd, create Devices table
INFO  [alembic.runtime.migration] Running upgrade 3303697cfdfd -> 8b7c9cfc4c56, create DeviceLeaseRecords table
INFO  [alembic.runtime.migration] Running upgrade 8b7c9cfc4c56 -> c10375863126, create indices for device columns
INFO  [alembic.runtime.migration] Running upgrade c10375863126 -> 095e5749b6e6, create ExtendLeaseRequests table
```

### 3. Verify with `psql`

Verify by connecting via `psql` and running queries.

```bash
psql -h localhost -U postgres -p 5433
```

## Warnings

**Note: There will be hardening eventually that should prevent these things
from happening.**

1. Do NOT run migrations at any branch not ToT. This may cause
inconsistencies in data that will be difficult to rollback.
2. Do NOT modify any data directly.
3. Do NOT check in a modified `db_config.ini` file.
