# sprobe

## Overview

`sprobe` is a systemd service health checker. It monitors services based on configured probes (Exec, HTTP, TCP) and ensures their health by automatically restarting failed services when needed. Additionally, it exposes Prometheus metrics on port `2112` to track service health.

## Features

- Supports multiple probe types:
  - **Exec**: Runs a command to check service health.
  - **HTTPGet**: Sends an HTTP GET request and evaluates the response.
  - **TCPSocket**: Checks if a TCP connection can be established.
- Configurable health check parameters:
  - Initial delay, period, timeout, failure threshold, success threshold.
- Automatic service restart on failure.
- Prometheus metrics exposure for monitoring.

## Installation

### Prerequisites
- Golang 1.18+
- Systemd
- Prometheus (optional, for metrics collection)

### Clone the Repository
```sh
$ git clone https://github.com/glendsoza/sprobe.git
$ cd sprobe
```

### Build
```sh
$ go build -o sprobe ./cmd/sprobe
```

### Install
```sh
$ sudo mv sprobe /usr/local/bin/
```

## Usage

### Configuration
Create a YAML configuration file with the following structure:

```yaml
- serviceName: "nginx.service"
  httpGet:
    path: "http://localhost"
    port: 80
  initialDelaySeconds: 1
  periodSeconds: 5
  timeoutSeconds: 10
  failureThreshold: 1
  successThreshold: 1
  autoRestart: false
```

### Configuration Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `serviceName` | string | Name of the systemd service being monitored. |
| `exec.command` | string | Command to execute for probing service health. |
| `httpGet.url` | string | URL to send an HTTP GET request to check service health. |
| `tcpSocket.port` | int | TCP port to probe for service availability. |
| `initialDelaySeconds` | int | Delay before the first probe is executed (in seconds). |
| `periodSeconds` | int | Time interval between consecutive probes (in seconds). |
| `timeoutSeconds` | int | Timeout for each probe attempt (in seconds). |
| `failureThreshold` | int | Number of consecutive failures before marking the service as unhealthy. |
| `successThreshold` | int | Number of consecutive successes before marking the service as healthy. |
| `autoRestart` | bool | Whether to automatically restart the service if it becomes unhealthy. |

### Running `sprobe`
```sh
$ sprobe start --config /path/to/config.yaml
```

### Prometheus Metrics
`sprobe` exposes service health metrics on port `2112`.

Example metric:
```sh
$ curl http://localhost:2112/metrics
```
Output:
```
sprobe_service_health{service_name="my-service"} 0
```
(0 = Unhealthy, 1 = Healthy, -1 = Unknown)

## Contributing

1. Fork the repository.
2. Create a new feature branch.
3. Commit changes.
4. Push the branch and open a pull request.

## License

This project is licensed under the MIT License. See the LICENSE file for details.

## Contact

For any questions or issues, please open an issue in the GitHub repository.

