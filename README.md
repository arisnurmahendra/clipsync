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

## Google Apps Script (GAS) Setup

This application relies on Google Apps Script to store and retrieve clipboard content remotely. Follow these steps to set up your own GAS endpoint:

1. Go to [Google Apps Script](https://script.google.com/)
2. Click "New Project" to create a new script
3. Delete the default `Code.gs` content and replace it with the content from the `Code.js` file in this repository (rename it to `Code.gs` in the Google Apps Script editor)
4. In the left sidebar, click the "+" icon and select "HTML" to create an HTML file
5. Name the file "index" and copy the content from the `index.html` file in this repository
6. Save the project (Ctrl+S)
7. Click "Deploy" in the top menu and select "New deployment"
8. Select "Web app" as the type
9. Fill in the required fields:
   - "I want to deploy from": "Head"
   - "Execute as": "Me"
   - "Who has access": "Anyone" (or "Anyone with Google account" depending on your preference)
10. Click "Deploy" and copy the provided web app URL
11. Take the URL and append `?mode=api` to the end (e.g., `https://script.google.com/macros/s/YOUR_SCRIPT_ID/exec?mode=api`)
12. Update the `gas_url` field in `config.json` with this URL

## Configuration

On first run, the application will generate a `config.json` file with default settings. Edit this file to customize:

- `gas_url`: Your Google Apps Script URL for remote clipboard sync (with `?mode=api` parameter)
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
- Access the web interface by opening the GAS URL in a browser (without the `?mode=api` parameter) to manually manage clipboard content

## Security Note

Note that this application directly calls Windows API to simulate keyboard input and clipboard operations, which may be flagged by security software.

## License

This is an open source project licensed under the MIT license.