# ipreg
ipreg is a standalone process that periodically scans a list of
configured subnets to determine which addresses are in use.
This information is made available through a web interface
that also allows users to 'claim' an IP address for use. 
This is useful in small research envrionments where
subnets are shared and IP address usage must be de-conflicted
by claiming which addresses are in use.

##Installation
* Install Go
* Set $GOPATH 
* Run go get github.com/nealjc/ipreg
* Run go install github.com/nealjc/ipreg
* ipreg will now be located at $GOPATH/bin. Add $GOPATH/bin to your
  path.

ipreg requires a configuration file at /etc/ipreg.conf. 
There is a sample configuration file, config.txt, in
$GOPATH/src/github.com/nealjc/ipreg that can be used
as a basis for /etc/ipreg.conf.

### Configure ipreg for your network
Edit the following fields in/etc/ipreg.conf

* Database: ipreg will create its database in this directory. It must exist and ipreg must have write access to it.
* Html: Specifies the directory ipreg will use to serve the main web page. 
* The remainder of the settings in /etc/ipreg.conf are optional.

Copy the $GOPATH/src/github.com/nealjc/ipreg/index.html file to the directory specified in the Html field above.

Next, edit the index.html file for the following:

* Set the Javascript variable ipregServer to be the IP address of the
  server running ipreg.
* Set the Javascript variable ipregServerPort to match ListenPort in
    the /etc/ipreg.conf file.


## Running


*TODO*
