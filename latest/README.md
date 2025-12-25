# MKS Parser PostgreSQL Extension

This extension integrates the MKS SQL Parser (Go shared library) into PostgreSQL.

## Prerequisites
- PostgreSQL development headers (`postgresql-server-dev-wa.b.` or `postgresql-devel`)
- GCC or Clang compiler
- Go compiler (to build the shared library)

## Build Instructions

1. **Build the Go Shared Library**
   Navigate to the `mks_so` directory and build the shared object:
   ```bash
   cd ../mks_so
   go build -buildmode=c-shared -o mks_sql.so .
   ```

2. **Build and Install the Extension**
   Navigate to this directory (`pg_extension`) and run:
   ```bash
   make
   sudo make install
   ```

3. **Register in PostgreSQL**
   Log into your database and run:
   ```sql
   CREATE EXTENSION mks_parser;
   SELECT mks_parser('select * from t where id = $1', '{"id": 123}');
   ```

## Troubleshooting
- If `make` fails with `pg_config: command not found`, ensure PostgreSQL bin directory is in your PATH.
- If loading fails with "cannot open shared object file", ensure `mks_sql.so` is in a location accessible by the PostgreSQL server or configured in `LD_LIBRARY_PATH`.
