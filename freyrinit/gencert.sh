#!/bin/bash

openssl req -x509 -newkey rsa:4096 -keyout privkey.pem -out fullchain.cert.pem -days 365
openssl rsa -in privkey.pem -out privkey.pem
mv {privkey,fullchain.cert}.pem ../nginx/conf/