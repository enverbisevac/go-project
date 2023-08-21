#!/bin/bash

if [[ "$OSTYPE" == "win32" ]]; then
    tskill sample
else
    pkill sample
fi

# never add here -d because it will be deleted by app and no need to check
# because app will run without that parameter
pid=$(pgrep -f $(which sample) --port=5555 --dsn="./tests/app.db" --log-dir=./tests/logs)

rm ./tests/logs/app.log
rm ./tests/app.db

if [ -z "$pid" ]
then
    sample server --port=5555 --dsn="./tests/app.db" -d --log-dir=./tests/logs
    while [[ "$(curl -s -o /dev/null -w ''%{http_code}'' localhost:5555/status)" != "200" ]]
    do 
        echo "Waiting for server...";
        sleep 5;
    done    
else
    echo "app is running pid: $pid"
fi