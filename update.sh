#!/bin/bash

git pull &&
go build -o gologd &&
./gologd rest &&
systemctl status goLogD
