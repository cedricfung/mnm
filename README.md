# mnm

Ⓜ️NⓂ️  is a process monitor which will automatically send notifications to the configured groups in Mixin Messenger.

Add 7000103800 as a contact, then add it to the groups you wish to receive notifications. Afterwards, refresh this page.

## 🎡 Typical Usages

Make a large tarball? Let mnm monitor that and notify you when it's done.

```
mnm run 'tar jcvf snapshots.tar.bz2 snapshots'
```

Download a large file? Let mnm monitor that and notify you when it's done.

```
mnm run 'wget https://some.large/file.zip'
```

Even if you already have a passive external monitor for your services, you may use mnm as a proactive monitor for your service.

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

Endless possibilities, and yet convenient to go
