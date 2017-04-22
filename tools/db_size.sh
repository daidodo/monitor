#!/bin/bash

mysql -h db -u root -pmysql -e "select table_name, table_rows, round((data_length+index_length)/1024,2) as 'size(KB)' from information_schema.tables where information_schema.tables.table_schema='overlord'"
