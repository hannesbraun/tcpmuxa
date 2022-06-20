# tcpmuxa

A TCPMUX implementation conforming to [RFC 1078](https://datatracker.ietf.org/doc/html/rfc1078)

## Building

Make sure you have Go (1.18 or higher) installed. Then run: `make build`.
You'll find the executable in the `bin` directory.

## Usage

You have to create a configuration file first in order to use tcpmuxa. Otherwise no service will be available.

### Configuration file format

* One line represents one service.
* Fields are seperated trough whitespace.
* Configuration variables can be specified using the following syntax: `$key=value`.
  Only the `port` can optionally be specified with this syntax right now.
* Every service starts with its name. The name is case-insensitive.
* A network service continues with `net`, the hostname or IP followed by the port.
* A local service continues with `local`, the path to the executable followed by its arguments.
* Leading and trailing whitespace is ignored.
* Whitespace can't be escaped.
* Lines starting with a `#` are considered comments
* Empty lines are allowed and do nothing.

Example:

```
$port=4242

# Local service: date
date            local   /bin/date

# Network service: rcssmonitor3d
rcssmonitor3d   net     127.0.0.1 3200
```

### Running tcpmuxa

Place the configuration file named `tcpmuxa.conf` either in the same directory you're running tcpmuxa from (option 1)
or provide a path to the configuration file as an argument for the executable (option 2).

```sh
# Option 1:
tcpmuxa

# Option 2:
tcpmuxa /etc/tcpmuxa.conf
```

## License

This project is licensed under the GNU General Public License 3. See [LICENSE](LICENSE) for more details.
