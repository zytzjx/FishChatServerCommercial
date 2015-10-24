#!/bin/bash

pkill gateway
pkill msg_server
pkill router

./gateway/gateway -conf_file ./gateway/gateway.json &
./msg_server/msg_server -conf_file ./msg_server/msg_server.19000.json &
./msg_server/msg_server -conf_file ./msg_server/msg_server.19001.json &
./router/router -conf_file ./router/router.json &
