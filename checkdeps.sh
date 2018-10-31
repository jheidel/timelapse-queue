#!/bin/sh

# Need yarn for building.
if ! [ -x "$(command -v yarn)" ]; then
  echo "Missing yarn. https://yarnpkg.com/lang/en/docs/install"
  exit 1
fi

# Need polymer for frontend compilation.
if ! [ -x "$(command -v polymer)" ]; then
  echo "Missing polymer CLI. https://www.polymer-project.org/3.0/start/install-3-0"
  exit 1
fi

# Need go-bindata for asset packaging.
if ! [ -x "$(command -v go-bindata)" ]; then
  echo "Missing go-bindata."
  echo "On ubuntu: sudo apt-get install go-bindata"
  exit 1
fi

# Check for libjpeg-turbo (dependency of pixiv/go-libjpeg)
ldconfig -p | grep libjpeg > /dev/null
if [ $? -ne 0 ]; then
  echo "Missing libjpeg."
  echo "libjpeg-turbo recommended for better performance."
  echo "On ubuntu: sudo apt-get install libjpeg-turbo8-dev"
  exit 1
fi

# Need go for building.
if ! [ -x "$(command -v go)" ]; then
  echo "Missing go. https://golang.org/doc/install"
  echo "On ubuntu: sudo apt-get install golang-go"
  exit 1
fi

echo "All dependencies ok."
exit 0
