EXTENSION = mks_parser
DATA = mks_parser--1.0.sql
MODULE_big = mks_parser
OBJS = mks_parser.o

# Helper to locate mks_so directory
MKS_SO_DIR = ../mks_so

PG_CONFIG = pg_config
PGXS := $(shell $(PG_CONFIG) --pgxs)
include $(PGXS)

# Add inclusion path for mks_sql.h
PKG_CPPFLAGS += -I$(MKS_SO_DIR)

# Link against the Go shared library
# We assume mks_sql.so is in MKS_SO_DIR
SHLIB_LINK += -L$(MKS_SO_DIR) -lmks_sql -Wl,-rpath,$(MKS_SO_DIR)
