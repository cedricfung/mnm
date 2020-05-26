# pgn

process guard &amp; notifier

## Usage

Get [Mixin Messenger](https://mixin.one/messenger) and find the Webhook bot `7000000012`, then get your token.

```
curl https://github.com/cedricfung/pgn/raw/master/pgn-linux-x64 -o pgn
chmod +x pgn
./pgn -run 'ls -l /tmp' -token ACCESS-TOKEN
```
