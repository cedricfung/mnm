# pgn

process guard &amp; notifier

## Usage

1. Get [Mixin Messenger](https://mixin.one/messenger).
2. Add Webhook bot `7000000012` as contact.
3. Obtain your access token.

```
curl -L https://github.com/cedricfung/pgn/raw/master/pgn-linux-x64 -o pgn
chmod +x pgn
./pgn -run 'ls -l /tmp' -token ACCESS-TOKEN
```
