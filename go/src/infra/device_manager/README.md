# Device Manager

go/device-manager-design

[TOC]

## Development

The Device Manager service uses containers for deployment. To replicate the
experience, the local development uses Docker containers.

The stack contains two containers:

1.  Postgres DB
2.  Device Lease Service

### Set Up

You must have the following installed:

*   Docker -
    [Instructions](https://g3doc.corp.google.com/cloud/containers/g3doc/glinux-docker/install.md?cl=head#installation).
    This is needed to run any containers locally.
*   Docker Compose Plugin -
    [Instructions](https://docs.docker.com/compose/install/linux/#install-the-plugin-manually).
    The dev setup is defined with a `docker-compose.dev.yml` file. It
    streamlines commands and makes configurations more maintainable.
*   grpcurl - Should already be installed. A useful tool to send requests and
    test the gRPC service.

#### 0. Bring Up the Containers

Once everything is installed, in separate windows, run the following:

```bash
# This brings up the Postgres container.
#
# If you already have Postgres running, you may have to modify the port in
# docker-compose.dev.yml.
make docker-db

# This brings up the Device Lease service container.
make docker-service
```

Your containers should be producing Docker logs at this point.

#### 1. Set Up Virtual Env

The database is blank when the container first spins up. To set up the database,
first create a virtualenv and install Alembic database migration tool:

```bash
python -m venv dev-env

pip install -r requirements.txt

source dev-env/bin/activate
```

The necessary tools to manage the Postgres or AlloyDB database are installed.

#### 2. Connect to Postgres DB

The `database/db_config.ini` file contains the sections for defining the
database configs for different environments. The default values for the local
setup are provided in the `alloydb.local` section.

Next, with `dev-env` activated, try connecting to the Docker Postgres database:

```bash
ALEMBIC_ENV=alloydb.local alembic history
```

The command should show that no migrations have been run yet. Run the migrations
to bring the database schema up to date:

```bash
ALEMBIC_ENV=alloydb.local alembic upgrade head

# Output should look something like this
INFO  [alembic.runtime.migration] Context impl PostgresqlImpl.
INFO  [alembic.runtime.migration] Will assume transactional DDL.
INFO  [alembic.runtime.migration] Running upgrade  -> 3303697cfdfd, create Devices table
INFO  [alembic.runtime.migration] Running upgrade 3303697cfdfd -> 8b7c9cfc4c56, create DeviceLeaseRecords table
...
```

#### 3. Connect to the Device Lease Service

With the Postgres database set up, try making a request to the service via
`grpcurl`:

```bash
grpcurl -plaintext -H "Authorization: Bearer $(gcloud auth print-identity-token)" localhost:50051 list chromiumos.test.api.DeviceLeaseService

# Output should look like this
chromiumos.test.api.DeviceLeaseService.ExtendLease
chromiumos.test.api.DeviceLeaseService.GetDevice
chromiumos.test.api.DeviceLeaseService.LeaseDevice
chromiumos.test.api.DeviceLeaseService.ListDevices
chromiumos.test.api.DeviceLeaseService.ReleaseDevice
```