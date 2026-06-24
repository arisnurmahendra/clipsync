# ClipSync - Remote Clipboard Daemon

ClipSync is a lightweight Windows clipboard synchronization and automation tool built with Go. It focuses on providing a seamless GUI desktop experience without console windows, featuring system tray integration, hotkey triggers, and clipboard content processing.

## Features

- **Silent GUI Launch**: Built with `-H windowsgui` to hide console windows, showing only a system tray icon
- **System Tray Integration**: Based on `github.com/getlantern/systray` for menu, icon, and exit controls
- **Global Hotkey Support**: Register shortcuts (like Ctrl+Shift+V) to trigger custom actions (such as simulated paste)
- **Clipboard Monitoring and Injection**: Read/write text clipboard content with support for automatic typing or formatting
- **Configuration Driven Behavior**: Load parameters like GAS URL through `config.json`, automatically generates default configuration on first run
- **Windows Native Interaction**: Encapsulates Windows API calls to implement low-level operations
- **Human-like Typing Simulation**: Simulates natural typing behavior with configurable delays, burst lengths, and occasional typos

## Installation

1. Make sure you have Go 1.26.4 installed
2. Clone the repository
3. Install dependencies:
   ```bash
   go mod tidy
   ```
4. Place an `icon.ico` file in the project root (you can download free .ico files from https://icon-icons.com or convert PNG to ICO at https://convertio.co)
5. Build the application:
   ```bash
   go build -ldflags="-H windowsgui" -o clipsync.exe .
   ```

## Configuration

On first run, the application will generate a `config.json` file with default settings. Edit this file to customize:

- `gas_url`: Your Google Apps Script URL for remote clipboard sync
- `delay_preset`: Typing speed presets (slow, normal, fast, custom)
- `human_like_typing`: Enable/disable human-like typing simulation
- `burst_length_min/max`: Minimum and maximum characters typed in bursts
- `long_pause_freq`: Frequency of longer pauses during typing
- `error_rate`: Percentage chance of typos
- `correction_rate`: Percentage chance of correcting typos

## Building with Metadata

To include metadata in the executable (icon, version, company, etc.):

1. Install goversioninfo:
   ```bash
   go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
   ```

2. Generate the resource file:
   ```bash
   goversioninfo -64 -o resource_windows_amd64.syso versioninfo.json
   ```

3. Build the application normally:
   ```bash
   go build -ldflags="-H windowsgui" -o clipsync.exe .
   ```

## Usage

- Run the application and it will appear in the system tray
- Right-click the tray icon to access the menu
- Configure your Google Apps Script URL in `config.json`
- Use the hotkey (INS key by default) to fetch content from your GAS endpoint and simulate typing
- Adjust typing parameters through the tray menu

## Security Note

Note that this application directly calls Windows API to simulate keyboard input and clipboard operations, which may be flagged by security software.

## License

This is an open source project licensed under the MIT license.