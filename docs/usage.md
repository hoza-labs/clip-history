# Usage guide

## Commands

- `ch config` shows all effective configuration parameters as TOML, including paths, log retention, log time zone, and sensitive-item storage, after defaults and normalization are applied.
- `ch start` starts background clipboard monitoring.
- `ch stop` stops background clipboard monitoring.
- `ch status` shows the clipboard history directory and whether monitoring is running.
- `ch version` shows the current target release version.
- `ch --help` shows the root help text.

`ch monitor` is an internal worker command used by `ch start` and is intentionally omitted from normal help output.

While monitoring, each clipboard update logs the available format IDs and names,
followed by a format count. Standard formats use names such as `CF_UNICODETEXT`;
registered formats use their Windows names, such as `HTML Format`. Unnamed formats
are logged as `unknown` with their numeric ID. If clipboard access or enumeration
fails, the failure is logged and monitoring continues with subsequent events.
Access-denied errors when opening the clipboard are retried up to 10 times with
25 ms between attempts (225 ms of total retry delays). Recovery is logged with
the attempt count; persistent failures are logged after the retry limit.
This stage does not read clip contents or save clips to `HistoryDir`.

## Configuration

The application stores its TOML config, logs, and monitor runtime state in a directory named `.clip-history` beneath the user's home directory. Logs are written as `clip-history-{yyyy-mm-dd}.log`, rotate automatically when the configured local date changes, and matching old log files are removed after `logRetentionDays`. Log timestamps and file dates use `logTimezone`, which defaults to `local`; other values must be valid Go time zone names such as `America/Los_Angeles` or `UTC`.
