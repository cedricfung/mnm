#!/bin/sh

# mnm is a process monitor & notifier
# notifications delivered to Mixin Messenger
# curl -sSL https://mnm.sh/in/MM-WEBHOOK-TOKEN.sh | bash

TOKEN=MM-WEBHOOK-TOKEN
BIN=/usr/bin/mnm
ETC=/etc/default/mnm

case "$(uname -m)" in
    x86_64)     arch="x64"  ;;
    aarch64)    arch="arm64";;
esac

case "$(uname -s)" in
    Linux*)     bin="linux-${arch}";;
    Darwin*)    bin="macos-${arch}";;
    *)          bin="linux-${arch}";;
esac

RAW="https://raw.githubusercontent.com/cedricfung/mnm/master/bin/pgn-${bin}"

echo "🚀 curl -L ${RAW} -o /tmp/mnm"
echo
curl -L "${RAW}" -o /tmp/mnm
echo

if [ -f "${BIN}" ]
then
    echo "🌝 ${BIN} INSTALLED"
    exit 0
fi

if [ -f "${ETC}" ]
then
    echo "🌝 ${ETC} CONFIGURED"
    exit 0
fi

if file /tmp/mnm | grep -q "executable"
then
    echo "🧭 chmod +x /tmp/mnm"
    chmod +x /tmp/mnm

    echo "🧭 sudo mv /tmp/mnm /usr/bin/"
    sudo mv /tmp/mnm /usr/bin/

    echo "🧭 echo ${TOKEN} | sudo tee ${ETC}"
    echo "${TOKEN}" | sudo tee "${ETC}"
else
    echo "🌚 OOPS"
fi

echo "🌞 OK"
echo

echo "🚨 run a test notification"
echo "🚨 mnm run 'ls -l /tmp'"
mnm run 'ls -l /tmp'
echo

echo "🌞 OK"
echo
