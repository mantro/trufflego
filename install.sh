#!/bin/bash

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)

if [[ -f "/usr/local/bin/trufflego" ]]; then
  echo "/usr/local/bin/trufflego already exists, removing.."
  sudo rm /usr/local/bin/trufflego

fi

echo sudo ln -s "$SCRIPT_DIR/trufflego-stub.sh" "/usr/local/bin/trufflego"
sudo ln -s "$SCRIPT_DIR/trufflego-stub.sh" "/usr/local/bin/trufflego"
