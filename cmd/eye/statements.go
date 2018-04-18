/*
 * Copyright (c) 2016, 1&1 Internet SE
 * Written by Jörg Pernfuß <joerg.pernfuss@1und1.de>
 * All rights reserved.
 */

package main

const stmtGetLookupIDForItem = `
SELECT lookup_id
FROM   eye.configuration_items
WHERE  configuration_item_id = $1::uuid;`

const stmtGetItemCountForLookupID = `
SELECT COUNT(1)::integer
FROM   eye.configuration_items
WHERE  lookup_id = $1::varchar;`

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
