#!/bin/bash
set -e

# Go to shared lib dir and build
echo "Building mks_sql.so..."
cd ../mks_so
go build -buildmode=c-shared -o mks_sql.so .
cp mks_sql.h ../test/
cp mks_sql.so ../test/

# Go back to test dir
cd ../test

# Compile test app
echo "Compiling test_app..."
# We need to set rpath so it finds the .so at runtime in the current dir
gcc -o test_app test_app.c ./mks_sql.so -Wl,-rpath,.

# Run test
echo "Running test..."
./test_app test01.sql test01.json
