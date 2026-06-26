FILE="$1"

jq -n \
  --arg filename "$(basename "$FILE")" \
  --arg md5 "$(openssl md5 -binary "$FILE" | base64)" \
  --argjson size "$(stat -c%s "$FILE")" \
  '{
      file_name: $filename,
      size: $size,
      file_hash: $md5
   }'
