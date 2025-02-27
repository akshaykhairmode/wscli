# wscli

A lightweight and powerful Go command-line tool for interacting with WebSocket servers. Designed for testing, debugging, and scripting, `wscli` provides functionality similar to `wscat` but with additional features. This tool is actively developed.

## 🚀 Installation

### Using `Docker`

```sh
docker run -it akshaykhairmode/wscli:latest -c "ws://example.com/ws"
```

### Using `go install`
```sh
go install github.com/akshaykhairmode/wscli@latest
```

### Download Prebuilt Binaries
If you don’t have Go installed, download the latest binaries from the [Releases Page](https://github.com/akshaykhairmode/wscli/releases).

## 🔧 Usage

To quickly connect to a WebSocket server:
```sh
wscli -c ws://localhost:8080/ws
```

## ✨ Features

- **🔹 Native Binaries:** Distributable and easy to install across systems.
- **📤 Piped Input:** Send piped input by using pipe `|`. _(Note: Interactive terminal features are disabled when using this mode.)_
- **📨 Multiple Messages on Connect:** Send multiple messages immediately after connecting.
- **🎭 Background Execution:**
  - Run `wscli` in the background using `nohup`:
    ```sh
    nohup wscli -c ws://localhost/ws -w 1s > nohup.out 2>&1 &
    ```
  - Redirect output and run in the background:
    ```sh
    wscli -c ws://localhost/ws >> output.txt & 2>&1
    ```
- **📜 History Persistence:** Maintain a command history for quick reuse.
- **⚡ Command Execution on Connect:** Use `-x` to execute commands automatically after connection.
- **📌 JSON Pretty Printing:** Format JSON responses with the `--jspp` flag.
- **⌨️ Terminal Shortcuts:** Supports readline shortcuts like `Ctrl+W` (delete word) and `Ctrl+R` (reverse search). [See full list](https://github.com/chzyer/readline/blob/master/doc/shortcut.md).

## 🛠 Available Flags

| Flag             | Shorthand | Description |
|-----------------|----------|-------------|
| `--auth`       |          | HTTP Basic Authentication credentials (e.g., `username:password`). |
| `--binary`     | `-b`     | Send hex encoded data to server |
| `--ca`         |          | Path to the CA certificate file (optional). |
| `--cert`       |          | Path to the client certificate file (optional). |
| `--connect`    | `-c`     | WebSocket connection URL. |
| `--execute`    | `-x`     | Execute a command after connecting (use multiple times for multiple commands). |
| `--gzipr`      |          | Enable gzip decoding if server messages are gzip-encoded. _(Note: Server must send messages as binary.)_ |
| `--header`     | `-H`     | Custom headers (`key:value`, can be used multiple times). |
| `--help`       | `-h`     | Display help information. |
| `--jspp`       |          | Enable JSON pretty printing for responses. |
| `--key`        |          | Path to the certificate key file (optional). |
| `--no-check`   | `-n`     | Disable TLS certificate verification. |
| `--no-color`   |          | Disable colored output. |
| `--origin`     | `-o`     | Specify origin for the WebSocket connection (optional). |
| `--proxy`      |          | Use a proxy URL. |
| `--response`   | `-r`     | Display HTTP response headers from the server. |
| `--show-ping-pong` | `-P` | Show ping/pong messages. |
| `--slash`      |          | Enable slash commands _(Experimental)_. |
| `--sub-protocol` | `-s`   | Specify a sub-protocol for the WebSocket connection (optional, can be used multiple times). |
| `--verbose`    | `-v`     | Enable debug logging. |
| `--version`    | `-V`     | Display version information. |
| `--wait`       | `-w`     | Wait time after command execution (`1s`, `1m`, `1h`). |

## 🚧 Upcoming Features (TODO)

- **Enhanced Slash Commands:** Ability to read binary files and send them to the server.
- **Basic Load Generation:** Support for load testing via interactive mode.
- **WebSocket Listener:** Implement a feature to start a WebSocket server.
