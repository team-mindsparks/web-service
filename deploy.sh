#! /bin/bash

# build
GOOS=linux go build -o /tmp/ws

# push to server
scp /tmp/ws root@188.226.156.181:/var/www/ws.new

# stop service
ssh root@188.226.156.181 'service mindsparks_service stop'

# mv binary
ssh root@188.226.156.181 'mv /var/www/ws.new /var/www/ws'

# restart service
ssh root@188.226.156.181 'service mindsparks_service start'
