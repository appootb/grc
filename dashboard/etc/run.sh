#!/bin/sh

# start nginx in background
/usr/sbin/nginx

# run dashboard
/data/dashboard/bin/dashboard -cfg=/etc/config.yaml