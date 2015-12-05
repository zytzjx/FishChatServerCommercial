nohup ./msg_server -conf_file=./msg_server.19000.json 1>&2 3>&2 4>&2 5>&2 6>&2 7>&2 8>&2 2>/var/log/msg_server.19000.log &
nohup ./msg_server -conf_file=./msg_server.19001.json 1>&2 3>&2 4>&2 5>&2 6>&2 7>&2 8>&2 2>/var/log/msg_server.19001.log &

ps aux | grep msg_server
