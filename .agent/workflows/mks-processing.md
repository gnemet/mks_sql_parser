---
description: how to use mks sql parser and execute in database
---

# Create a MKS SQL

- $1 means jsonb object
  select 1 as id
  , $1 ? 'key' as key_exists
  , $1->>'key' as key_value
  , 'JSON KEY EXISTS' as info --#key

# Create jsonb input

- this jsonb will be used in SQL query
  {
  "key" : "json value"
  }

# MKS parsing

- parsing MSK SQL keep or delete lines/block depends of MKS rules
- add a sql text

# SQL execute

## COPY mode

- replace sql's $1 with jsonb input with cast
- COPY ( sql ) OUT STDOUT CSV WITH HEADING
  select 1 as id
  , jsonb '{
  "key" : "json value"
  }' ? 'key' as key_exists
  , jsonb '{
  "key" : "json value"
  }'->>'key' as key_value
  , 'JSON KEY EXISTS' as info --#key
  -- run the modified sql, get data CSV format
  id,key_exists,key_value,info
  1,true,json value,JSON KEY EXISTS

## EXECUTE mode

- replace sql's $1 with $1::jsonb
- Formula: `SELECT * FROM ( [SQL] ) AS mks_wrapper LIMIT $2`
- Create execute command: `execute ( SELECT * FROM ( SQL ) AS mks_wrapper LIMIT 100 ) USING jsonb '...'`
- Print execute command to SQL Log:
  ```sql
  EXECUTE (
    SELECT * FROM (
      SELECT 1 as id, $1::jsonb ? 'key' as key_exists, $1::jsonb->>'key' as key_value, 'JSON KEPT' as info
    ) AS mks_wrapper
    LIMIT 100
  ) USING jsonb '{"key": "val"}'
  ```

# display sql result

- show data in table using header
