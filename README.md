# GoMap

A comprehensive network scanning tool written in Go with both GUI and command-line interfaces.

> **Note**: This project is currently under active development and is not yet fully functional. Many features are still being implemented and may not work as expected.

## Features

- Network scanning and host discovery
- Port scanning with customizable ranges
- Modern and classic GUI interfaces
- Banner grabbing capabilities
- Vulnerability scanning

## Installation

```bash
git clone https://github.com/adehlbom/GoMap.git
cd GoMap
go mod download
```

## Usage

### GUI Mode (Default)

```bash
go run main.go
```

By default, GoMap launches with the modern UI interface. To use the classic UI:

```bash
go run main.go --classic
```
or
```bash
go run main.go -c
```

### Command-line Mode

Command-line functionality is currently being implemented.

## Building

To build a standalone executable:

```bash
go build -o gomap
```

## License

GoMap is released under the GNU General Public License v3.0 (GPL-3.0).

This means:
- You are free to use, modify, and distribute this software.
- If you distribute modified versions, you must distribute them under the same license (GPL-3.0).
- You must disclose your source code when you distribute this software.
- This license includes a strong copyleft provision, meaning all derivative works must also be open-sourced under GPL-3.0.

While this license does not strictly prohibit commercial use, it requires that any commercial distributions must also be open source and share their modifications, which protects against proprietary commercial exploitation.

For the full license text, see the LICENSE file in this repository or visit https://www.gnu.org/licenses/gpl-3.0.html
