# ipreg
ipreg is a standalone process that periodically scans a list of
configured subnets to determine which addresses are in use.
This information is made available through a web interface
that also allows users to 'claim' an IP address for use. 
This is useful in small research envrionments where
subnets are shared and IP address usage must be de-conflicted
by claiming which addresses are in use.

## Installation

### Pre-requisites
* ipreg requires sqlite3 libraries.

### Building from source

* Install Go 1.2+
* Set $GOPATH. This will be where ipreg will be downloaded to and
  built from. 
* Run go get github.com/nealjc/ipreg
* ipreg will now be located at $GOPATH/bin. Copy $GOPATH/bin/ipreg to /usr/local/bin


### Configure ipreg for your network

ipreg requires a configuration file at /etc/ipreg.conf. 
There is a sample configuration file, config.txt, in
$GOPATH/src/github.com/nealjc/ipreg that can be used
as a basis for /etc/ipreg.conf.

Edit the following fields in/etc/ipreg.conf

* DatabaseDir: ipreg will create its database in this directory. It must exist and ipreg must have write access to it.
* HtmlDir: Specifies the directory ipreg will use to serve the main web page.
* Add subnets. DO NOT include any '/'s in Subnet names.
* The remainder of the settings in /etc/ipreg.conf are optional. 

Copy the $GOPATH/src/github.com/nealjc/ipreg/index.html file to the directory specified in the Html field above.

Next, edit the index.html file for the following:

* Set the Javascript variable ipregServer to be the IP address of the
  server running ipreg.
* Set the Javascript variable ipregServerPort to match ListenPort in
    the /etc/ipreg.conf file.


## Running

### Ubuntu
The scripts/ directory contains a script that can be used with Ubuntu
upstart. Move it to /etc/init/ directory and ipreg can be
started/stopped with the start and stop commands in Ubuntu, e.g., sudo
start ipreg.



