#!/bin/bash

set -e

BASE_DIR=$PWD
INTERVAL=1
ENV_ARG=""

for arg in "$@"; do
  case "$arg" in
    --env=*)
      ENV_ARG="$arg"
      ;;
  esac
done

ACTION="$1"
SERVER="xserver"
ENTRY="./main.go"

function ensure_env_arg() {
  if [ -z "$ENV_ARG" ]; then
    echo "failed, miss env args, etc: sh server.sh $ACTION --env=dev|test|prod"
    exit 1
  fi
}

function build() {
  go build -o "$SERVER" "$ENTRY"
  echo "$SERVER build success"
}

function process_id() {
  pgrep -f "$BASE_DIR/$SERVER" -u "$UID" || true
}

function start() {
  ensure_env_arg

  if [ "$(process_id)" != "" ]; then
    echo "$SERVER already running"
    exit 1
  fi

  echo "$BASE_DIR/$SERVER $ENV_ARG"
  nohup "$BASE_DIR/$SERVER" "$ENV_ARG" >> /dev/null 2>&1 &
  echo "sleeping..." && sleep $INTERVAL

  if [ "$(process_id)" == "" ]; then
    echo "$SERVER start failed"
    exit 1
  else
    echo "start success"
  fi
}

function status() {
  if [ "$(process_id)" != "" ]; then
    echo "$SERVER is running"
  else
    echo "$SERVER is not running"
  fi
}

function stop() {
  PID=$(process_id)
  if [ "$PID" != "" ]; then
    kill $PID
  fi

  echo "sleeping..." && sleep $INTERVAL

  if [ "$(process_id)" != "" ]; then
    echo "$SERVER stop failed"
    exit 1
  else
    echo "stop success"
  fi
}

function version() {
  "$BASE_DIR/$SERVER" version
}

case "$ACTION" in
  build)
    build
    ;;
  run)
    build
    stop
    start
    ;;
  start)
    start
    ;;
  stop)
    stop
    ;;
  status)
    status
    ;;
  restart)
    stop
    start
    ;;
  version)
    version
    ;;
  *)
    echo "usage: $0 {build|run|start|stop|restart|status|version} --env=dev|test|prod"
    exit 1
    ;;
esac
