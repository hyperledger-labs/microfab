
* When using the client SDKs, do not use the 'asLocalhost' option set to true. It must be false. Without this clients will connect to microfab passing 'localhost' as a host name. Microfab needs the hostname to work out where to route the connection. Which it won't be able to do if 'asLocalhost' is set
