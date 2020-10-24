#!/bin/bash

pkill -e -f "^go run . -n g001-scr-.* -bot"

NUM_BOTS=$1

for ((i=1; i <= $NUM_BOTS; i++))
do
  echo "Welcome $i bot"
  screen -dmS g001-scr-$i bash -c "go run . -n g001-scr-$i -bot"
  sleep 0.025
done

sleep 1

ps -ef | grep g001-scr- | grep -v -e SCREEN -e grep
echo "started: $(ps -ef | grep g001-scr- | grep -c -v -e SCREEN -e grep)"
