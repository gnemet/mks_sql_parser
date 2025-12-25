-- complain if script is sourced in psql, rather than via CREATE EXTENSION
\echo Use "CREATE EXTENSION mks_parser" to load this file. \quit

CREATE OR REPLACE FUNCTION mks_parser(text, text)
RETURNS text
AS 'MODULE_PATHNAME'
LANGUAGE C STRICT;
