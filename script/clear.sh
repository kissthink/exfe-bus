#!/bin/sh

DIR=`dirname $0`
cd ${DIR}

DB=`./config.py '["db"]["db_name"]'`
DB_NAME=`./config.py '["db"]["username"]'`
DB_PASS=`./config.py '["db"]["password"]'`

PASS_ARG="-p${DB_PASS}"
if [ "${DB_PASS}" = "" ]
then
	PASS_ARG="-u${DB_NAME}"
fi

echo "DELETE FROM tokens WHERE expire_at < DATE_ADD(NOW(), INTERVAL \"-7\" DAY);" | mysql -u"${DB_NAME}" "${PASS_ARG}" ${DB}
