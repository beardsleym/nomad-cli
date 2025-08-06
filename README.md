# Nomad CLI

Nomad CLI is a versatile command-line tool designed for the modern digital nomad. It provides quick access to essential information like currency exchange rates, weather forecasts, timezone conversions, and network speed tests.

## Features

- **Currency Conversion:** Get the latest exchange rates and convert between different currencies.
- **Weather Information:** Check the current weather for any location, with auto-detection based on your IP address.
- **Timezone Lookup:** Find the current time in any city or timezone.
- **Network Speed Test:** Test your internet connection's download/upload speed, latency, and jitter.

## Installation

You can install `nomad-cli` directly using `go install`:

```bash
go install github.com/beardsleym/nomad-cli@latest
```

After installation, you can run the tool using:

```bash
nomad-cli --help
```

If you get a "command not found" error and you are using Zsh, you may need to add the Go bin directory to your `PATH`. You can do this by running the following commands:

```bash
echo 'export PATH=$(go env GOPATH)/bin:$PATH' >> ~/.zshrc
source ~/.zshrc
```

Alternatively, you can build from source:

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/your-username/nomad-cli.git
    cd nomad-cli
    ```

2.  **Build the executable:**
    ```bash
    go build .
    ```

3.  **(Optional) Add to your PATH:**
    For easy access, move the `nomad-cli` executable to a directory in your system's PATH (e.g., `/usr/local/bin`).

## Usage

### Currency Conversion

```bash
nomad cv <amount> <from_currency> <to_currency>
```

**Example:**

```bash
nomad cv 1000 thb aud
```

### Weather

```bash
nomad w [city]
```

If no city is provided, it will automatically detect your location.

**Examples:**

```bash
nomad w
nomad w "New York"
```

### Time

```bash
nomad t <city or address>
```

**Example:**

```bash
nomad t Tokyo
```

### Speed Test

```bash
nomad s
```

This command will run a comprehensive speed test and provide a quality assessment for streaming, gaming, and web chat.

## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue.
