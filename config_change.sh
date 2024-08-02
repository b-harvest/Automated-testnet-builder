#!/bin/bash

usage() {
    echo "Usage: $0 <validator home folder>"
    exit 1
}

if [ -z "$1" ]; then
    usage
fi

FILE_PATH="./peer-node/config/app.toml"

sed -i '' 's#address = "tcp://localhost:1317"#address = "tcp://localhost:11317"#g' "$FILE_PATH"
sed -i '' 's#address = "localhost:9090"#address = "localhost:19090"#g' "$FILE_PATH"
sed -i '' 's#address = "127.0.0.1:8545"#address = "127.0.0.1:18545"#g' "$FILE_PATH"
sed -i '' 's#ws-address = "127.0.0.1:8546"#ws-address = "127.0.0.1:18546"#g' "$FILE_PATH"
sed -i '' 's#metrics-address = "127.0.0.1:6065"#metrics-address = "127.0.0.1:16065"#g' "$FILE_PATH"


FILE_PATH="./peer-node/config/config.toml"

sed -i '' 's#proxy_app = "tcp://127.0.0.1:26658"#proxy_app = "tcp://127.0.0.1:36658"#g' "$FILE_PATH"
sed -i '' 's#laddr = "tcp://127.0.0.1:26657"#laddr = "tcp://127.0.0.1:36657"#g' "$FILE_PATH"
sed -i '' 's#pprof_laddr = "localhost:6060"#pprof_laddr = "localhost:16060"#g' "$FILE_PATH"
sed -i '' 's#laddr = "tcp://0.0.0.0:26656"#laddr = "tcp://0.0.0.0:36656"#g' "$FILE_PATH"


PEERID=$(cantod tendermint  show-node-id  --home $1)
sed -i '' "s#persistent_peers = \"\"#persistent_peers = \"$PEERID\@localhost:26656\"#g" "$FILE_PATH"


section_found=false

TEMP_FILE=$(mktemp)

while IFS= read -r line; do
  if [[ $line == "[grpc-web]" ]]; then
    section_found=true
  fi

  if $section_found && [[ $line != "[grpc-web]" && $line == "["* ]]; then
    echo 'address = "localhost:19091"' >> "$TEMP_FILE"
    section_found=false
  fi

  # 라인 쓰기
  echo "$line" >> "$TEMP_FILE"
done < "$FILE_PATH"

if $section_found; then
  echo 'address = "localhost:19091"' >> "$TEMP_FILE"
fi

mv "$FILE_PATH" "$FILE_PATH.bak"
mv "$TEMP_FILE" "$FILE_PATH"