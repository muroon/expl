
# expl

MySQL explain checker.
This tool runs multiple SQL explain from SQL log.

- Read Log or DB(mysql.general_log table) and explain multiple SQL
- Content filtering, Display explain result with 'SELECT_TYPE', 'TABLE', 'TYPE', 'EXTRA' specified

# simple usage

```
$expl explain simple "select * from memo" --database database1 --host localhost --user root --pass ""

  DataBase:  database1
  SQL:       select * from memo
+----+------------+-------+------------+------+--------------+-----+--------+-----+------+----------+-------+
| ID | SELECTTYPE | TABLE | PARTITIONS | TYPE | POSSIBLEKEYS | KEY | KEYLEN | REF | ROWS | FILTERED | EXTRA |
+----+------------+-------+------------+------+--------------+-----+--------+-----+------+----------+-------+
|  1 | SIMPLE     | memo  |          0 | ALL  |              |     |      0 |     |   39 | 100.0000 |       |
+----+------------+-------+------------+------+--------------+-----+--------+-----+------+----------+-------+
```

# the case with logs of multiple databases in one file

## 1. Create config file ("conf" sub command)

Make config file in YAML format.
This includes relationships between tables and databases.

- add setting in config file

```
# expl conf add host database user pass -c config_file_path

$expl conf add localhost database1 root "" -c config.yaml
$expl conf add localhost database2 root "" -c config.yaml
```

## 2. Execute Explian ("explain" sub command)

Execute explain multipule SQL

1. make mapping relationships between tables and databases in config YAML file
2. execute explain SQL using using table-database mapping

This has advantages such as using the "Combine SQL" (see below)

```
#expl explain mode -c config_file_path --format format_type --log sql_log_file_path

$expl explain log -c config.yaml --format simple --log simple.yaml
```

# explain sub command

### mode parameter

2nd Parameter

- simple : SQL direct input. The third parameter is sql.
- log : Getting SQL from log file. (official generate log or custom log file)
- log-db : Getting SQL from database. (mysql.general_log table)

```
# simple mode
$expl explain simple "select * from memo" --database database1 --host localhost --user root --pass ""

# log mode
$expl explain log --conf config.yaml --format official --log sql.log

# log-db mode
$expl explain log-db --conf config.yaml --format official
```

### conf option

This is Config file path. The config file includes host, database, user, password and table-database mapping.


### log option

This is Log file path.

### format option

format of one line in SQL log file.

- simple : Raw SQL
- official : Same log format of MySQL general_log. https://dev.mysql.com/doc/refman/5.6/en/query-log.html
- command : Edit by OS command

```
# simple format
$expl explain simple "select * from memo" --database database1 --host localhost --user root --pass ""

# official format
$expl explain log --conf config.yaml --format official --log /var/lib/mysql/general_sql.log

# command format
$expl explain log --conf config.yaml --log custom_sql.log --format command --format-cmd "cut -c 21-"
```

#### format-cmd option

Using only "command" format.
OS command for edit line of log to raw SQL.

```
$expl explain log --conf config.yaml --log custom_sql.log --format command --format-cmd "cut -c 21-"

# same (using pipe simple mode)
$cut -c 21- custom_sql.log | xargs -I$ expl explain simple "$" --conf config.yaml --format command --format-cmd "cut -c 21-"
```

### Combine SQL option

Display the same type of SQL results in one view.
The two SQL statements below are identical to the explain result. Thus, SQLs of the same type are displayed together in one

```
# sql1
select * from memo where id = 1;

# sql2
select * from memo where id = 100;
```

```
$expl explain log --conf config.yaml --format official --log /var/lib/mysql/general_sql.log --combine-sql
```

### filter option

Filtering the explain results

| option | meaning |
----|----
| filter-select-type | Show only results included in the specified "Select Type" of explain |
| filter-no-select-type | Show only results that are not included in the specified "Select Type" of explain |
| filter-table | Show only results that contain the specified table |
| filter-no-table | Show only results that do not contain the specified table |
| filter-type | Show only results included in the specified "Type" of explain |
| filter-no-type | Show only results that are not included in the specified "Type" of explain |
| filter-extra | Show only results included in the specified "Extra" of explain |
| filter-no-extra | Show only results that are not included in the specified "Extra" of explain |

```
# view only results where includes "ALL" in TYPE column.

$expl explain log --conf config.yaml --format official --log /var/lib/mysql/general_sql.log --filter-type ALL

  DataBase:  memo_sample
  SQL:       select tag.* from tag, tag_memo where tag.id = tag_memo.tag_id
+----+------------+----------+------------+------+--------------+---------+--------+--------------------+------+----------+-------------+
| ID | SELECTTYPE |  TABLE   | PARTITIONS | TYPE | POSSIBLEKEYS |   KEY   | KEYLEN |        REF         | ROWS | FILTERED |    EXTRA    |
+----+------------+----------+------------+------+--------------+---------+--------+--------------------+------+----------+-------------+
|  1 | SIMPLE     | tag      |          0 | ALL  | PRIMARY      |         |      0 |                    |   22 | 100.0000 |             |
|  1 | SIMPLE     | tag_memo |          0 | ref  | PRIMARY      | PRIMARY |      4 | memo_sample.tag.id |    1 | 100.0000 | Using index |
+----+------------+----------+------------+------+--------------+---------+--------+--------------------+------+----------+-------------+
```

### ignore error option

This is to ignore the "Explain SQL Error" or "SQL Parse Error".

```
$expl explain log --conf config.yaml --format official --log /var/lib/mysql/general_sql.log --ignore-error
```
With SQL parse or explain SQL errors, let's try use this option.

- If table "user" exists in "database1" and "database2", this tool will try to explain both "database1" and "database2". In this case, this option is useful to ignore the error and do the following processing.
- If the log contains only one unparsable string, using this option will not stop the execution of subsequent correct SQL statement lines.

### option file option

Using file for option settings
You can use YAML files instead of directly specifying options in the command

If there are duplicate definitions, priority is given in the following order
1. Command
2. env
3. option file

```
$expl explain log --option-file ./option.yaml --filter-extra "using where"
```

### verbose output option

Display the value of the option just before execution


