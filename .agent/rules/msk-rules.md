---
trigger: always_on
---

# MKSQL Parser Rules

$1 means the first argument and jsonb object in postgresql function

## json operators:
    scalar operators:
        json ->, #>
        text ->>, #>>
        boolean ->&, #&
        numeric ->#, #=
        integer ->^, #^
        timestamp ->@, #@
    array operators:
        text array ->>>, #>>> 
        boolean array ->&&, #&&
        numeric array ->##, #>##
        integer array ->^^, #^^
        timestamp array ->@@, #@@

## json key format:
    string key: operator start with "-" + other operators + 'key'
    for example: $1->'key' or $1->>'key' or $1 #>> '{key1,key2}' or $1 ->@'key' or $1 ->^'key' or $1 ->#'key' or $1 ->&'key' or $1 ->@@'key' or $1 ->&&'key' or $1 ->##'key' or $1 ->^^'key' 
    nested key: operator start with "#" + other operators + '{key1,key2}'   
    for example: $1#>> '{key1,key2}' or $1 #>^ '{key1,key2}' 

## Key existence in json expression:
    $1{operator}'key' 
        for example: $1->'key' or $1->>'key' or $1 #>> '{key1,key2}' or $1 ->@'key' or $1 ->^'key' or $1 ->#'key' or $1 ->&'key' or $1 ->@@'key' or $1 ->&&'key' or $1 ->##'key' or $1 ->^^'key' 
    if key exists, it keep the line 
    if key does not exist, it delete the line
    for example: select  $1->'key1' as jsonb
                        ,$1->>'key2' as text
                        ,$1#>> '{key1,key2}'as "text"
                        ,$1->@'key3' as timestamp
                        ,$1->^'key4' as integer
                        ,$1-># 'key5' as numeric
                        ,$1->& 'key6' as boolean
                        ,$1->@@'key7' as array_of_timestamp
                        ,$1->&&'key8' as array_of_boolean
                        ,$1->##'key9' as array_of_numeric
                        ,$1->^^'key10' as array_of_integer
Key existence in line comment:
    --#key
    if key exists, it keep the line 
    if key does not exist, it delete the line
    for example
        select 'EXISTS' as "exists" --#key
Key not existence in line comment:
    --#!key
    if key exists, it delete the line 
    if key does not exist, it keep the line
    for example
        select 'EXISTS' as "exists" --#!key    
Key value in line comment:
    --#key:value
    if key exists and value is equal, it keep the line 
    if key does not exist or value is not equal, it delete the line
    for example
        select 'EXISTS' as "exists" --#key:value    
Key not value in line comment:
    --#!key:value or --#key!:value
    if key exists and value is equal, it delete the line 
    if key does not exist or value is not equal, it keep the line
    for example
        select 'EXISTS' as "exists" --#!key:value
        select 'EXISTS' as "exists" --#key!:value    
Block key existence logic:
    --<key
    if key exists, it keep the block 
    if key does not exist, it delete the block
    for example:
        --<key
        select 'EXISTS' as "exists" 
        -->key
Block key not existence logic:
    --<!key
    if key exists, it delete the block 
    if key does not exist, it keep the block
    for example:
        --<!key
        select 'EXISTS' as "exists" 
        -->!key
Block key value logic:
    --<key:value
    if key exists and value is equal, it keep the block 
    if key does not exist or value is not equal, it delete the block
    for example:
        --<key:value
        select 'EXISTS' as "exists" 
        --key>
Block key not value logic:
    --<!key:value or --<key!:value
    if key exists and value is equal, it delete the block 
    if key does not exist or value is not equal, it keep the block
    for example:
        --<!key:value
        select 'EXISTS' as "exists" 
        --key>
        --<key!:value
        select 'EXISTS' as "exists" 
        --key>
jsonpath logic in line comment:
    --#{ EXPRESSION }
    if expression is true, it keep the line 
    if expression is false, it delete the line
    for example:
        select 'EXISTS' as "exists" --#{ $.key == "value" }
        select 'EXISTS' as "exists" --#{ $.key != "value" }
        select 'EXISTS' as "exists" --#{ $.key == "value" && $.key2 == "value2" }
        select 'EXISTS' as "exists" --#{ $.key == "value" || $.key2 == "value2" }
        select 'EXISTS' as "exists" --#{ $.key == "value" || $.key2 == "value2" || $.key3 == "value3" }
jsonpath logic in block:
    --<{ EXPRESSION }
    if expression is true, it keep the block 
    if expression is false, it delete the block
    for example:
        --<{ $.key == "value" }
        select 'EXISTS' as "exists" 
        -->     

replace or empty:
    %key%
    if key exists, it replace the tag    
    if key does not exist, replace the tag to empty string
    for example:
        select %key% as key_value_or_empty

 replace or delete:
    :key, $key
    if key exists, it replace the tag 
    if key does not exist, it delete the full line  
    for example:
        select :key as key_value_or_delete
        select $key as key_value_or_delete
    
    
