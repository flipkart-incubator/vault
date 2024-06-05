// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package cockroachdb

// sqlKeywords is a reference of all of the keywords that we do not allow for use as the table name
// Referenced from:
// https://www.cockroachlabs.com/docs/stable/keywords-and-identifiers.html#identifiers
// -> https://www.cockroachlabs.com/docs/stable/keywords-and-identifiers.html#keywords
// -> https://www.cockroachlabs.com/docs/stable/sql-grammar.html
var sqlKeywords = map[string]bool{
	// reserved_keyword
	// https://www.cockroachlabs.com/docs/stable/sql-grammar.html#reserved_keyword
	"ALL":               true,
	"ANALYSE":           true,
	"ANALYZE":           true,
	"AND":               true,
	"ANY":               true,
	"ARRAY":             true,
	"AS":                true,
	"ASC":               true,
	"ASYMMETRIC":        true,
	"BOTH":              true,
	"CASE":              true,
	"CAST":              true,
	"CHECK":             true,
	"COLLATE":           true,
	"COLUMN":            true,
	"CONCURRENTLY":      true,
	"CONSTRAINT":        true,
	"CREATE":            true,
	"CURRENT_CATALOG":   true,
	"CURRENT_DATE":      true,
	"CURRENT_ROLE":      true,
	"CURRENT_SCHEMA":    true,
	"CURRENT_TIME":      true,
	"CURRENT_TIMESTAMP": true,
	"CURRENT_USER":      true,
	"DEFAULT":           true,
	"DEFERRABLE":        true,
	"DESC":              true,
	"DISTINCT":          true,
	"DO":                true,
	"ELSE":              true,
	"END":               true,
	"EXCEPT":            true,
	"FALSE":             true,
	"FETCH":             true,
	"FOR":               true,
	"FOREIGN":           true,
	"FROM":              true,
	"GRANT":             true,
	"GROUP":             true,
	"HAVING":            true,
	"IN":                true,
	"INITIALLY":         true,
	"INTERSECT":         true,
	"INTO":              true,
	"LATERAL":           true,
	"LEADING":           true,
	"LIMIT":             true,
	"LOCALTIME":         true,
	"LOCALTIMESTAMP":    true,
	"NOT":               true,
	"NULL":              true,
	"OFFSET":            true,
	"ON":                true,
	"ONLY":              true,
	"OR":                true,
	"ORDER":             true,
	"PLACING":           true,
	"PRIMARY":           true,
	"REFERENCES":        true,
	"RETURNING":         true,
	"SELECT":            true,
	"SESSION_USER":      true,
	"SOME":              true,
	"SYMMETRIC":         true,
	"TABLE":             true,
	"THEN":              true,
	"TO":                true,
	"TRAILING":          true,
	"TRUE":              true,
	"UNION":             true,
	"UNIQUE":            true,
	"USER":              true,
	"USING":             true,
	"VARIADIC":          true,
	"WHEN":              true,
	"WHERE":             true,
	"WINDOW":            true,
	"WITH":              true,

	// cockroachdb_extra_reserved_keyword
	// https://www.cockroachlabs.com/docs/stable/sql-grammar.html#cockroachdb_extra_reserved_keyword
	"INDEX":   true,
	"NOTHING": true,

	// type_func_name_keyword
	// https://www.cockroachlabs.com/docs/stable/sql-grammar.html#type_func_name_keyword
	"COLLATION": true,
	"CROSS":     true,
	"FULL":      true,
	"INNER":     true,
	"ILIKE":     true,
	"IS":        true,
	"ISNULL":    true,
	"JOIN":      true,
	"LEFT":      true,
	"LIKE":      true,
	"NATURAL":   true,
	"NONE":      true,
	"NOTNULL":   true,
	"OUTER":     true,
	"OVERLAPS":  true,
	"RIGHT":     true,
	"SIMILAR":   true,
	"FAMILY":    true,

	// col_name_keyword
	// https://www.cockroachlabs.com/docs/stable/sql-grammar.html#col_name_keyword
	"ANNOTATE_TYPE":    true,
	"BETWEEN":          true,
	"BIGINT":           true,
	"BIT":              true,
	"BOOLEAN":          true,
	"CHAR":             true,
	"CHARACTER":        true,
	"CHARACTERISTICS":  true,
	"COALESCE":         true,
	"DEC":              true,
	"DECIMAL":          true,
	"EXISTS":           true,
	"EXTRACT":          true,
	"EXTRACT_DURATION": true,
	"FLOAT":            true,
	"GREATEST":         true,
	"GROUPING":         true,
	"IF":               true,
	"IFERROR":          true,
	"IFNULL":           true,
	"INT":              true,
	"INTEGER":          true,
	"INTERVAL":         true,
	"ISERROR":          true,
	"LEAST":            true,
	"NULLIF":           true,
	"NUMERIC":          true,
	"OUT":              true,
	"OVERLAY":          true,
	"POSITION":         true,
	"PRECISION":        true,
	"REAL":             true,
	"ROW":              true,
	"SMALLINT":         true,
	"SUBSTRING":        true,
	"TIME":             true,
	"TIMETZ":           true,
	"TIMESTAMP":        true,
	"TIMESTAMPTZ":      true,
	"TREAT":            true,
	"TRIM":             true,
	"VALUES":           true,
	"VARBIT":           true,
	"VARCHAR":          true,
	"VIRTUAL":          true,
	"WORK":             true,

	// unreserved_keyword
	// https://www.cockroachlabs.com/docs/stable/sql-grammar.html#unreserved_keyword
	"ABORT":                     true,
	"ACTION":                    true,
	"ADD":                       true,
	"ADMIN":                     true,
	"AGGREGATE":                 true,
	"ALTER":                     true,
	"AT":                        true,
	"AUTOMATIC":                 true,
	"AUTHORIZATION":             true,
	"BACKUP":                    true,
	"BEGIN":                     true,
	"BIGSERIAL":                 true,
	"BLOB":                      true,
	"BOOL":                      true,
	"BUCKET_COUNT":              true,
	"BUNDLE":                    true,
	"BY":                        true,
	"BYTEA":                     true,
	"BYTES":                     true,
	"CACHE":                     true,
	"CANCEL":                    true,
	"CASCADE":                   true,
	"CHANGEFEED":                true,
	"CLUSTER":                   true,
	"COLUMNS":                   true,
	"COMMENT":                   true,
	"COMMIT":                    true,
	"COMMITTED":                 true,
	"COMPACT":                   true,
	"COMPLETE":                  true,
	"CONFLICT":                  true,
	"CONFIGURATION":             true,
	"CONFIGURATIONS":            true,
	"CONFIGURE":                 true,
	"CONSTRAINTS":               true,
	"CONVERSION":                true,
	"COPY":                      true,
	"COVERING":                  true,
	"CREATEROLE":                true,
	"CUBE":                      true,
	"CURRENT":                   true,
	"CYCLE":                     true,
	"DATA":                      true,
	"DATABASE":                  true,
	"DATABASES":                 true,
	"DATE":                      true,
	"DAY":                       true,
	"DEALLOCATE":                true,
	"DELETE":                    true,
	"DEFERRED":                  true,
	"DISCARD":                   true,
	"DOMAIN":                    true,
	"DOUBLE":                    true,
	"DROP":                      true,
	"ENCODING":                  true,
	"ENUM":                      true,
	"ESCAPE":                    true,
	"EXCLUDE":                   true,
	"EXECUTE":                   true,
	"EXPERIMENTAL":              true,
	"EXPERIMENTAL_AUDIT":        true,
	"EXPERIMENTAL_FINGERPRINTS": true,
	"EXPERIMENTAL_RELOCATE":     true,
	"EXPERIMENTAL_REPLICA":      true,
	"EXPIRATION":                true,
	"EXPLAIN":                   true,
	"EXPORT":                    true,
	"EXTENSION":                 true,
	"FILES":                     true,
	"FILTER":                    true,
	"FIRST":                     true,
	"FLOAT4":                    true,
	"FLOAT8":                    true,
	"FOLLOWING":                 true,
	"FORCE_INDEX":               true,
	"FUNCTION":                  true,
	"GLOBAL":                    true,
	"GRANTS":                    true,
	"GROUPS":                    true,
	"HASH":                      true,
	"HIGH":                      true,
	"HISTOGRAM":                 true,
	"HOUR":                      true,
	"IMMEDIATE":                 true,
	"IMPORT":                    true,
	"INCLUDE":                   true,
	"INCREMENT":                 true,
	"INCREMENTAL":               true,
	"INDEXES":                   true,
	"INET":                      true,
	"INJECT":                    true,
	"INSERT":                    true,
	"INT2":                      true,
	"INT2VECTOR":                true,
	"INT4":                      true,
	"INT8":                      true,
	"INT64":                     true,
	"INTERLEAVE":                true,
	"INVERTED":                  true,
	"ISOLATION":                 true,
	"JOB":                       true,
	"JOBS":                      true,
	"JSON":                      true,
	"JSONB":                     true,
	"KEY":                       true,
	"KEYS":                      true,
	"KV":                        true,
	"LANGUAGE":                  true,
	"LAST":                      true,
	"LC_COLLATE":                true,
	"LC_CTYPE":                  true,
	"LEASE":                     true,
	"LESS":                      true,
	"LEVEL":                     true,
	"LIST":                      true,
	"LOCAL":                     true,
	"LOCKED":                    true,
	"LOGIN":                     true,
	"LOOKUP":                    true,
	"LOW":                       true,
	"MATCH":                     true,
	"MATERIALIZED":              true,
	"MAXVALUE":                  true,
	"MERGE":                     true,
	"MINUTE":                    true,
	"MINVALUE":                  true,
	"MONTH":                     true,
	"NAMES":                     true,
	"NAN":                       true,
	"NAME":                      true,
	"NEXT":                      true,
	"NO":                        true,
	"NORMAL":                    true,
	"NO_INDEX_JOIN":             true,
	"NOCREATEROLE":              true,
	"NOLOGIN":                   true,
	"NOWAIT":                    true,
	"NULLS":                     true,
	"IGNORE_FOREIGN_KEYS":       true,
	"OF":                        true,
	"OFF":                       true,
	"OID":                       true,
	"OIDS":                      true,
	"OIDVECTOR":                 true,
	"OPERATOR":                  true,
	"OPT":                       true,
	"OPTION":                    true,
	"OPTIONS":                   true,
	"ORDINALITY":                true,
	"OTHERS":                    true,
	"OVER":                      true,
	"OWNED":                     true,
	"PARENT":                    true,
	"PARTIAL":                   true,
	"PARTITION":                 true,
	"PARTITIONS":                true,
	"PASSWORD":                  true,
	"PAUSE":                     true,
	"PHYSICAL":                  true,
	"PLAN":                      true,
	"PLANS":                     true,
	"PRECEDING":                 true,
	"PREPARE":                   true,
	"PRESERVE":                  true,
	"PRIORITY":                  true,
	"PUBLIC":                    true,
	"PUBLICATION":               true,
	"QUERIES":                   true,
	"QUERY":                     true,
	"RANGE":                     true,
	"RANGES":                    true,
	"READ":                      true,
	"RECURSIVE":                 true,
	"REF":                       true,
	"REGCLASS":                  true,
	"REGPROC":                   true,
	"REGPROCEDURE":              true,
	"REGNAMESPACE":              true,
	"REGTYPE":                   true,
	"REINDEX":                   true,
	"RELEASE":                   true,
	"RENAME":                    true,
	"REPEATABLE":                true,
	"REPLACE":                   true,
	"RESET":                     true,
	"RESTORE":                   true,
	"RESTRICT":                  true,
	"RESUME":                    true,
	"REVOKE":                    true,
	"ROLE":                      true,
	"ROLES":                     true,
	"ROLLBACK":                  true,
	"ROLLUP":                    true,
	"ROWS":                      true,
	"RULE":                      true,
	"SETTING":                   true,
	"SETTINGS":                  true,
	"STATUS":                    true,
	"SAVEPOINT":                 true,
	"SCATTER":                   true,
	"SCHEMA":                    true,
	"SCHEMAS":                   true,
	"SCRUB":                     true,
	"SEARCH":                    true,
	"SECOND":                    true,
	"SERIAL":                    true,
	"SERIALIZABLE":              true,
	"SERIAL2":                   true,
	"SERIAL4":                   true,
	"SERIAL8":                   true,
	"SEQUENCE":                  true,
	"SEQUENCES":                 true,
	"SERVER":                    true,
	"SESSION":                   true,
	"SESSIONS":                  true,
	"SET":                       true,
	"SHARE":                     true,
	"SHOW":                      true,
	"SIMPLE":                    true,
	"SKIP":                      true,
	"SMALLSERIAL":               true,
	"SNAPSHOT":                  true,
	"SPLIT":                     true,
	"SQL":                       true,
	"START":                     true,
	"STATISTICS":                true,
	"STDIN":                     true,
	"STORE":                     true,
	"STORED":                    true,
	"STORING":                   true,
	"STRICT":                    true,
	"STRING":                    true,
	"SUBSCRIPTION":              true,
	"SYNTAX":                    true,
	"SYSTEM":                    true,
	"TABLES":                    true,
	"TEMP":                      true,
	"TEMPLATE":                  true,
	"TEMPORARY":                 true,
	"TESTING_RELOCATE":          true,
	"TEXT":                      true,
	"TIES":                      true,
	"TRACE":                     true,
	"TRANSACTION":               true,
	"TRIGGER":                   true,
	"TRUNCATE":                  true,
	"TRUSTED":                   true,
	"TYPE":                      true,
	"THROTTLING":                true,
	"UNBOUNDED":                 true,
	"UNCOMMITTED":               true,
	"UNKNOWN":                   true,
	"UNLOGGED":                  true,
	"UNSPLIT":                   true,
	"UNTIL":                     true,
	"UPDATE":                    true,
	"UPSERT":                    true,
	"UUID":                      true,
	"USE":                       true,
	"USERS":                     true,
	"VALID":                     true,
	"VALIDATE":                  true,
	"VALUE":                     true,
	"VARYING":                   true,
	"VIEW":                      true,
	"WITHIN":                    true,
	"WITHOUT":                   true,
	"WRITE":                     true,
	"YEAR":                      true,
	"ZONE":                      true,
}
