# mnm

Ⓜ️NⓂ️  is a process monitor which will automatically send notifications to the configured groups in Mixin Messenger.

1. Get Mixin Messenger from https://messenger.mixin.one.
2. Search `mnm` and add it as a contact.
3. Add `mnm` to the groups you wish to receive notifications.
4. Open https://mnm.sh from your desktop browser.

## 🎡 Typical Usages

Have a long running task already? Get notified when it's done:

```bash
mnm monitor PID
```

Make a large tarball? Let mnm monitor that and notify you when it's done:

```
mnm run 'tar jcvf snapshots.tar.bz2 snapshots'
```

Download a large file? Let mnm monitor that and notify you when it's done:

```
mnm run 'wget https://some.large/file.zip'
```

Even if you already have a passive external monitor for your services, you may use mnm as a proactive monitor for your service:

```
[Unit]
Description=Mixin Network Kernel Daemon
After=network.target

[Service]
User=one
Type=simple
ExecStart=/usr/bin/mnm run '/usr/bin/mixin kernel -dir /data/mixin -port 7239'
Restart=on-failure
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
```

Endless possibilities, and yet convenient to go.
