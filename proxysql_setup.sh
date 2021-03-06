#!/bin/bash

while [ ! -f "/home/trove/.guestagent.prepare.end" ]; do

sleep 60

done

sed '10s/3306/9999/' /etc/mysql/my.cnf >> /home/trove/my_bak.cnf && sudo mv /home/trove/my_bak.cnf /etc/mysql/my.cnf
sudo systemctl restart mariadb.service

sudo dpkg -i /etc/proxysql_modified.deb

sudo service proxysql restart
sleep 60

sudo mysql -uadmin -padmin -P 6032 -e "INSERT INTO mysql_servers(hostgroup_id,hostname,port) VALUES (0,'127.0.0.1',9999);"
sudo mysql -uadmin -padmin -P 6032 -e "LOAD MYSQL SERVERS TO RUN;"
sudo mysql -uadmin -padmin -P 6032 -e "SAVE MYSQL SERVERS TO DISK;"
sudo mysql -uadmin -padmin -P 6032 -e "INSERT INTO mysql_query_rules (rule_id,active,username,match_digest,apply) VALUES (100,1,'os_admin','.',1);"
sudo mysql -uadmin -padmin -P 6032 -e "INSERT INTO mysql_query_rules (rule_id,active,username,schemaname,match_digest,apply) VALUES (200,1,'os_admin','mysql','.',1);"
sudo mysql -uadmin -padmin -P 6032 -e "INSERT INTO mysql_query_rules (rule_id,active,match_digest,error_msg,apply) VALUES (1000,1,'^[(select)|(insert)|(update)|(delete)|(create)|(drop)|(reload)|(process)|(references)|(index)|(alter)|(show databases)|(create temporary tables)|(lock tables)|(execute)|(replication slave)|(replication client)|(create view)|(show view)|(create routine)|(alter routine)|(create user)|(event)|(trigger)].* mysql\..*','Use of system database is forbidden !',1);"
sudo mysql -uadmin -padmin -P 6032 -e "INSERT INTO mysql_query_rules (rule_id,active,schemaname,match_digest,error_msg,apply) VALUES (2000,1,'mysql','.','Use of system database is forbidden !',1);"
sudo mysql -uadmin -padmin -P 6032 -e "LOAD MYSQL QUERY RULES TO RUN;"
sudo mysql -uadmin -padmin -P 6032 -e "SAVE MYSQL QUERY RULES TO DISK;"

#Setup cron task
echo "* * * * * /etc/proxysql_modified.sh" >> /etc/proxysqlcron
/usr/bin/crontab /etc/proxysqlcron
