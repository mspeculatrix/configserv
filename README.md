# configserv

`go get github.com/mspeculatrix/configserv`

Intended to run on a Raspberry Pi on board a robot. It's used to tell the RPi the location/configuration of a remote server used to, for example, log telemetry or other data. The idea is that, after starting up the robot, the user will employ a web app running on the remote server (see [robot_server](https://github.com/mspeculatrix/robot_server)) to send key info telling the RPi where to connect, and then the RPi will configure its settings accordingly.

Might be expandable to handle other configurations and even act as some kind of REST API for the robot.

It receives a GET request via HTTP and converts the query string to a map before saving the map as key/value pairs in a config file, using the format k=v.

The `up` Bash script is there to compile the binary for the Arm architecture and then push it to a Raspberry Pi on the local network. Configure as appropriate for your setup.

To understand what the hell is going on here, [read my blog](https://mansfield-devine.com/speculatrix/ "Machina Speculatrix").
