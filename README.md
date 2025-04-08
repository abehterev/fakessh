# Fake SSH Server

## Motivation
This project was born out of frustration with constant password-scanning attacks on SSH servers. After being tired of seeing numerous brute-force attempts in server logs, I decided to create a fake SSH server to study these attacks more closely. By logging the credentials used in these attacks, it became possible to:

1. Analyze what usernames and passwords are most commonly targeted
2. Study the patterns and techniques used by attackers
3. Track the origin of these attacks
4. Gather intelligence on current brute-force trends without exposing real systems

The server deliberately mimics real OpenSSH behavior but always rejects authentication, making it perfect for a honeypot to collect data on SSH attacks.

## Project Description
This project is a simulation of an SSH server operating on port 2222. The server emulates OpenSSH behavior from the protocol perspective but always rejects authentication attempts. The main goal is to record credentials used in login attempts.

## Architecture

### Main Components:
1. **SSH Server** — handles incoming SSH connections
   - Uses the `golang.org/x/crypto/ssh` library
   - Configured to support only password authentication
   - Mimics OpenSSH behavior and responses
   - Always returns an authentication error

2. **Credentials Logger** — saves authentication attempt data
   - Uses the `zerolog` library for structured logging
   - Records IP address, username, and password
   - **Supports both file logging and direct console output (stdout)**
   - Provides JSON and human-readable logging formats

3. **Configuration** — manages server settings
   - Port (default 2222)
   - Logging settings (format, file path)
   - SSH greeting banner
   - SSH server version
   - SSH key settings

### Architecture Layers:
- **Transport Layer** — TCP connection handling, SSH protocol
- **Authentication Layer** — processing login attempts
- **Logging Layer** — recording and storing credentials
- **Configuration Layer** — application behavior settings

## Technologies
- Programming Language: Go
- SSH Library: golang.org/x/crypto/ssh
- Logging: zerolog
- Configuration: viper + cobra

## Features
- Precise OpenSSH server emulation (version, banner, error messages)
- Structured credential logging in JSON or human-readable format
- Support for persistent SSH key (built-in, from file, or generating new ones)
- Flexible configuration through file, command-line flags, or environment variables

## Testing
The project is fully covered with unit tests using the standard Go testing package.

### Integration Testing
Integration tests can be run with the standard test command:
```bash
make test
```

### Docker Integration Testing
Docker integration tests verify that environment variables are properly applied in Docker containers. These tests are only run when specifically requested with the `docker` tag:
```bash
make test-docker
```

These tests build and run a Docker container with custom environment variables and verify that they are correctly applied to the running server.

## Building and Running

### Requirements
- Go 1.23 or higher (updated from 1.18 due to requirements of new dependencies such as crypto/ecdh, log/slog, and slices)
- Make (optional)

### Building from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/fakessh.git
cd fakessh

# Build the project
make build

# Run tests
make test

# Build and run
make run
```

### Command-line Parameters

```
Usage:
  fakessh [flags]

Flags:
      --banner string         SSH banner (version part) (default "Ubuntu-4ubuntu0.5")
      --config string         path to configuration file
      --generate-key          generate a new SSH key on each start (default true)
      --help                  help for command
      --key string            path to SSH private key (if not specified, built-in or newly generated will be used)
      --log string            path to credentials log file (use "stdout" for console output) (default "credentials.log")
      --log-format string     log format (json, pretty or text) (default "json")
      --port int              SSH server port (default 2222)
      --server-version string SSH server version (default "OpenSSH_8.2p1")
```

### Usage Examples

#### Running with Default Parameters (logs to file)
```bash
./build/fakessh
```

#### Running with Console Output (logs to stdout)
```bash
./build/fakessh --log stdout
```

#### Running with Console Output in Human-readable Format
```bash
./build/fakessh --log stdout --log-format pretty
```

#### Running with Specified Port and Log File
```bash
./build/fakessh --port 2222 --log /var/log/fakessh/credentials.log
```

#### Running with Custom SSH Key
```bash
./build/fakessh --key /path/to/ssh_host_key --generate-key=false
```

#### Running with Configuration File
```bash
./build/fakessh --config config.yaml
```

### Connecting to the Server

Since the server uses a fixed or generated key, it's recommended to disable known_hosts checking for test connections:

```bash
# Option 1: Disable host key checking
ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -p 2222 test@localhost

# Option 2: Temporarily ignore known_hosts file
ssh -o UserKnownHostsFile=/dev/null -p 2222 test@localhost
```

For production use, it's recommended to maintain a persistent server key using the `--key` option and not use `--generate-key`.

### Environment Variables
All settings can also be configured through environment variables with the `FAKESSH_` prefix:

```bash
FAKESSH_PORT=2222 FAKESSH_LOG_FILE=stdout FAKESSH_LOG_FORMAT=pretty ./build/fakessh
```

### Using Docker

#### Building Docker Image
```bash
make docker
```

#### Running with Docker
```bash
# Run with default settings
docker run -p 2222:2222 fakessh

# Run with custom port mapping
docker run -p 3333:2222 fakessh

# Run with custom parameters
docker run -p 2222:2222 fakessh --log stdout --log-format pretty --port 2222
```

#### Environment Variables in Docker
All settings can be configured through environment variables with the `FAKESSH_` prefix:

```bash
# Run with environment variables
docker run -p 2222:2222 \
  -e FAKESSH_LOG_FORMAT=pretty \
  -e FAKESSH_BANNER="Custom Banner" \
  -e FAKESSH_SERVER_VERSION="OpenSSH_7.4p1" \
  fakessh
```

These environment variables have been verified with integration tests. To run the Docker integration tests:

```bash
make test-docker
```

#### Available Environment Variables

| Environment Variable | Default Value | Description |
|----------------------|---------------|-------------|
| FAKESSH_PORT | 2222 | SSH server port |
| FAKESSH_LOG_FILE | stdout | Path to log file (stdout for console output) |
| FAKESSH_LOG_FORMAT | json | Log format (json, pretty, text) |
| FAKESSH_BANNER | Ubuntu-4ubuntu0.5 | SSH banner (version part) |
| FAKESSH_SERVER_VERSION | OpenSSH_8.2p1 | SSH server version |
| FAKESSH_GENERATE_KEY | false | Whether to generate a new SSH key on each start |
| FAKESSH_KEY | | Path to private key file inside container |

#### Persisting Logs and Custom Keys
You can mount volumes to persist logs and use custom keys:

```bash
# Mount a volume for logs and custom keys
docker run -p 2222:2222 \
  -v /host/path/to/logs:/app/logs \
  -v /host/path/to/keys:/app/keys \
  -e FAKESSH_LOG_FILE=/app/logs/credentials.log \
  -e FAKESSH_KEY=/app/keys/ssh_host_key \
  fakessh
```

### systemd Integration

The project includes a systemd unit file for running the server via Docker with journald log integration:

#### Installing systemd Service
```bash
# Copy unit file
sudo cp fakessh.service /etc/systemd/system/

# Reload systemd configuration
sudo systemctl daemon-reload

# Enable autostart
sudo systemctl enable fakessh.service

# Start service
sudo systemctl start fakessh.service
```

#### Viewing Logs through journald
```bash
# View all service logs
sudo journalctl -u fakessh

# View logs in real-time
sudo journalctl -u fakessh -f
```

## Log Formats and Destinations

### Log Destinations
The server can write logs to:
- **File** (default: credentials.log)
- **Console (stdout)** - ideal for Docker containers and systemd integration

### JSON Format (Default)
```json
{"level":"info","component":"auth","time":"2022-04-15T10:30:45Z","remote_addr":"192.168.1.100:54321","username":"admin","password":"password123","event":"auth_attempt","message":"authentication attempt"}
```

### Human-readable Format (pretty)
```
10:30:45 INF component=auth remote_addr=192.168.1.100:54321 username=admin password=password123 event=auth_attempt authentication attempt
``` 

## Practical Usage as a Honeypot

### Setting Up for Attack Monitoring
To effectively use this tool as a honeypot for monitoring SSH attacks:

1. **Deploy on a public server** with a real IP address (cloud VPS works well)
2. **Configure your firewall** to allow inbound connections on port 2222 (or map standard port 22 to 2222)
3. **Set up persistent logging** to a secure location:
   ```bash
   ./build/fakessh --log /var/log/ssh-attacks.log --log-format json
   ```
4. **Run as a service** using the provided systemd unit file to ensure continuous operation
5. **Regularly analyze the logs** to identify attack patterns

### Security Considerations
While this tool is designed to be secure, please keep the following in mind:

- This is a **honeypot** - do not deploy it on production servers
- Always run it as an unprivileged user
- Ensure logs are stored securely as they may contain sensitive information
- Regularly review the logs to detect if attackers are trying to exploit the tool itself
- Consider adding a banner explicitly stating this is a honeypot (depending on your goals)

### Example Insights from Real-world Deployment
After deploying this tool on a public IP for a week, common observations include:

- Most common targeted usernames: root, admin, user, ubuntu, postgres
- Many attacks follow predictable patterns (sequential attempts with common passwords)
- Significant number of attacks originate from specific regions
- Many attackers use the same toolkits, detectable by their attack patterns

This data can be valuable for understanding current threats and improving security on legitimate systems.

## License

This project is licensed under the [GNU General Public License v2.0](LICENSE) - see the LICENSE file for details.

## Repository

The official repository for this project is available at: [github.com/abehterev/fakessh](https://github.com/abehterev/fakessh) 