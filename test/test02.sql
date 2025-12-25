with ut as
        ( select  mutd.id                       as "id"
                , mutd.full_name                as "fullName"
--<version:4
                , $1->>'version'                as version
--version>
                , mutd.email                    as "email"
                , '%value%'                     as "replaceValue"
    --<type:list  -- type = list
                , $1->>'type'                   as "typeCheck" -- "type=list"
    --type>
    --<!type:list   -- type <> list
                , $1->>'type'                   as  "typeCheck" -- "type!=list"
    --type>
    --<valueLength:collect  -- short name
                , $1->>'valueLength'                   as "nameLength1" -- "valueLength=collect"
    --valueLength>
    --<valueLength:collected  -- long name
                , $1->>'valueLength'                   as  "nameLength2" -- "valueLength=collected"
    --valueLength>
    --<keyLong -- short key
                , 'blockKeyLong'                    as "keyLong1"
    --keyLong>
    --<keyLonger  -- longer key
                , 'blockKeyLonger'                  as "keyLonger1"
    --keyLonger>
                , $1->>'keyLong'                    as "keyLong2"
                , $1->>'keyLonger'                  as "keyLonger2"
                , 'keyLong'                         as "keyLong3"    -- #keyLong
                , 'keyLonger'                       as "keyLonger3"  -- #keyLonger
                , mutd.phone_number                 as "phoneNumber"
                , '{a,a,a}'::text[]                 as "castCheck"
                , 'this line must be here'          as "joinEavCheck"
$joinEav
--              select jsonb_agg( jsonb_build_object( 'hdr'   , jsonb_strip_nulls( to_jsonb( rslt ) - '{tableName}'::text[] ) )
--                  || jsonb_build_object( 'record', mkd_record( rslt."tableName", rslt."tableId" )->'data' ) -- #record
--                  order by rslt."scorePrcnt" desc, rslt."tableId" asc  ) as results
                , mutd.birth_date               as "birthDate"
    --<mode:full
                , mutd.birth_place              as "birthPlace"
                , mutd.mother_name              as "motherName"
                , mutd.gender                   as "gender"
                , mutd.address                  as "address"
                , 'shortKey'                    as "keyCheck"           -- #kulcs
                , 'longKey'                     as "keyCheck"           -- #kulcsok
                , $1->>'kulcs'                  as "keyCheckInput"
                , $1->>'kulcsok'                as "keyCheckInput"
                , 'longKey'                     as "keyCheck"           -- #kulcsok
                , true                          as "existsAddressTrue"  --#existsAddress
                , false                         as "existsAddressFalse"  --#!existsAddress
                , mutd.address                  as "address"
                , $2->>'constantValue'          as "constantValue"
                , :offset       as offset_colon
                , $offset       as offset_dollar
    --mode>
--<mode:include
                , ( @inc1 )     as current_include
                , ( @inc2 )     as other_include
                , ( @inc3 )     as current_include_1
                , ( @inc4 )     as other_include_2
                , ( @inc5 )     as other_include_3
                , ( @inc6 )     as other_include_4
--mode>
    --<!banana
                , 'banana not exists'           as "banana"    -- block key is not exists
    --banana>
    --<banana
                , 'banana exists'               as "banana"    -- block key is exists
    --banana>
--<{ $.krumpli == "ok" || $.krumpli == "nice" }
    , 'krumpli "ok" exists'                     as "krumpli"
-->
--<{ $.krumpli == "wrong" }
    , 'krumpli "wrong" exists'                  as "krumpli"    -- block key is exists
-->
    , 'burgonya "ok" exists'                    as "burgonya"   -- #{ $.burgonya == "ok" }
    , 'burgonya "wrong" exists'                 as "burgonya"   -- #{ $.burgonya == "wrong" }
--<{ exists($.key1) || exists($.key2) }
    , 'key1 or key2' as "isExistOrBlock"
-->
--<{ exists($.key1) && exists($.key2) }
    , 'key1 and key2' as "isExistAndBlock"
-->
    , 'key1 or key2' as "isExistOrLine"                 -- #{ exists($.key1) || exists($.key2) }
    , 'key1 and key2' as "isExistAndLine"               -- #{ exists($.key1) && exists($.key2) }
            from mks_unit_test as mutd
           where true
             and mutd.id           = ( $1->^'id'         )
             and mutd.full_name    = ( $1->>'fullName'   )
             and mutd.email        = ( $1->>'email'      )
             and mutd.phone_number = ( $1->>'phoneNumber')
             and mutd.birth_date   = ( $1->@'birthDate'  )
             and mutd.address      = ( $1->>'address'    )  -- #checkAddress:false
             and mutd.address      = ( $1->>'address'    )  -- #checkAddress
             and mutd.address      = ( $1->>'address'    )  -- #!checkAddress
             and mutd.address      = ( $1->>'address'    )  -- #checkAddress:true
             and mutd.address      = ( $1->>'address'    )  -- #checkAddress:!true
             and mutd.mother_name  = ( $1->>'motherName' )  -- #
             and mutd.gender       = ( $1->>'gender'     )  -- #
             and mutd.birth_place  = ( $1->>'birthPlace' )  -- #
           limit :limit
          offset :offset
        )
    select jsonb_agg ( ut )
      from ut
--recursive>