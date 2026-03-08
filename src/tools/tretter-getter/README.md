# tretter-getter

`tretter-getter` is a daemon that monitors the TVP VOD schedule and automatically records new episodes of *Na Dobre i Na Złe* (and potentially other shows).

It features:
- **Automatic Scheduling**: Checks TVP API for upcoming episodes.
- **Docker Worker**: Spawns ephemeral `yt-dlp` containers to record the stream.
- **Web Dashboard**: A beautiful, Material You (Beer CSS) dashboard to see what's recording and view history.
- **SQLite Database**: Tracks recorded episodes locally.

## Prerequsites

- **Go 1.22+**
- **Docker**
- **sqlc** (for generating DB code)

## Architecture

1.  **Daemon**: The main process runs indefinitely.
2.  **Scheduler**: `robfig/cron` triggers a check every minute.
3.  **Web Server**: `Echo` serves the API and UI.
4.  **Database**: `modernc.org/sqlite` stores episode metadata (pure Go driver).
5.  **Workers**: `yt-dlp` containers are spawned via Docker API.

## Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `TARGET_STATION_ID` | `1998766` | Station ID to monitor |
| `DATA_DIR` | `/data/app/tretter-getter` | DB and state directory |
| `DOWNLOADS_DIR` | `/downloads` | Video output directory |
| `STREAM_URL` | `https://vod.tvp.pl...` | Stream URL |
| `BUFFER_MINUTES` | `1` | Minutes before/after episode to record |
| `LOG_LEVEL` | `info` | Logging verbosity (debug, info, warn, error) |
| `DRY_RUN` | `false` | If true, don't actually spawn workers |
| `PORT` | `1945` | Web dashboard port |

## Makefile

Available `make` commands:

```console
$ make help
build                                     Build the binary
run                                       Run the tool
run-dry                                   Run the tool in dry-run mode
run-debug                                 Run the tool with debug logging
docker-build                              Build Docker image
test                                      Run unit tests
check                                     Run linters
format                                    Format code
tidy                                      Tidy dependencies
sqlc                                      Generate SQLC code
help                                      Show help message
```

## Authors
Developed and maintained by [Paweł Adamczak][pawelad].

Released under [Mozilla Public License 2.0][license].

[license]: ../../../LICENSE
[pawelad]: https://pawelad.me/
