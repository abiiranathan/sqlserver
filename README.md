# sqlserver

Communicate to an sqlite3 database over the network using a custom efficient binary protocol based og GOB encoding.

Has a dedicated `sqlserver` and `sqlclient` usable as CLIs. Both packages expose APIs for library use.

### Install sqlserver:

```console
go install github.com/abiiranathan/sqlserver@latest
```

### Install sqlclient:

```console
go install github.com/abiiranathan/sqlserver/cmd/sqlclient@latest
```

#### Server Usage:

```
sqlserver -host 0.0.0.0 -port 9999 -db db.sqlite3
```

#### Client Usage:

```
sqlclient -host localhost -port 9999
```

If you want pre-compiled binaries, check the Releases page.
