
# wscli

A Go command-line tool for interacting with WebSocket servers.  It's designed for testing and scripting, offering similar functionality to `wscat` with some enhancements.  This tool is currently under active development.

## Features

* **Native Binaries:**  Distributable and easy to install.
* **Piped Input:** Use the `--stdin` flag to pipe input to the WebSocket server.  Note: Interactive terminal features are not available when using piped input.
* **Multiple Messages on Connect:** Send a series of messages immediately upon establishing the connection.
* **Background Execution:**
    * Run in the background using `nohup` (redirect output to `nohup.out` and use `-w` to wait for messages). Example: `nohup wscli -c ws://localhost/ws -w 1s > nohup.out 2>&1 &`
    * Redirect output and run in the background. Example: `wscli -c ws://localhost/ws >> output.txt & 2>&1`
* **History Persistence:**  Maintain a history of commands for easy reuse.
* **Command Execution on Connect:** Use the `-x` flag (multiple times for multiple commands) to execute commands immediately after connection.  An interactive terminal will open after the commands are executed.
* **JSON Pretty Printing:** Format server responses as nicely formatted JSON using the `--jspp` flag.
* **Terminal Shortcuts:** Utilize standard terminal shortcuts like Ctrl+W (delete word) and Ctrl+R (reverse search).  A full list of available readline shortcuts can be found [here](https://github.com/chzyer/readline/blob/master/doc/shortcut.md).

## Available Flags

| Flag | Shorthand | Description |
|---|---|---|
| `--auth` |  | HTTP Basic Authentication credentials (e.g., `username:password`). |
| `--ca` |  | Path to the CA certificate file (optional). |
| `--cert` |  | Path to the client certificate file (optional). |
| `--connect` | `-c` | WebSocket connection URL. |
| `--execute` | `-x` | Command to execute after connection (can be used multiple times). |
| `--header` | `-H` | Custom header in `key:value` format (can be used multiple times, commas are also supported for multiple values). |
| `--help` | `-h` | Display help information. |
| `--jspp` |  | Enable JSON pretty printing for responses. |
| `--key` |  | Path to the certificate key file (optional). |
| `--no-check` | `-n` | Disable certificate verification. |
| `--no-color` |  | Disable colored output. |
| `--origin` | `-o` | Origin for the WebSocket connection (optional). |
| `--proxy` |  | Proxy URL. |
| `--response` | `-r` | Display HTTP response headers from the server. |
| `--show-ping-pong` | `-P` | Show ping/pong messages. |
| `--slash` |  | Enable slash commands (currently under development). |
| `--stdin` | `-i` | Read input from stdin. |
| `--sub-protocol` | `-s` | Sub-protocol for the WebSocket connection (optional, can be used multiple times). |
| `--verbose` | `-v` | Enable debug logging. |
| `--version` | `-V` | Display version information. |
| `--wait` | `-w` | Wait time after command execution (e.g., `1s`, `1m`, `1h`). |

## TODO

* **Enhanced Slash Commands:** Implement additional slash commands, such as reading binary files from a path and sending them to the server.
* **Basic Load Generation:** Add basic load generation capabilities from interactive mode using slash commands.