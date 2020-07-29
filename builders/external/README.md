# External service chaincode builder

This builder enables the chaincode build and launch lifecycle to be managed independently of peers.

_Status: should work!_

## Package contents

This builder requires a `.tgz` package file with the following contents.

### `metadata.json` file

```
{
    "type": "prebuilt",
    "label": "..."
}
```

where `label` is suitable string to identify your chaincode.

### `code.tar.gz` archive

The `code.tar.gz` file should contain a `connection.json` file to allow the peer to connect to the chaincode, for example:

```
{
    "address": "chaincode.example.com:9999",
    "dial_timeout": "10s",
    "tls_required": "true",
    "client_auth_required": "true",
    "client_key": "-----BEGIN EC PRIVATE KEY----- ... -----END EC PRIVATE KEY-----",
    "client_cert": "-----BEGIN CERTIFICATE----- ... -----END CERTIFICATE-----",
    "root_cert": "-----BEGIN CERTIFICATE---- ... -----END CERTIFICATE-----"
}
```
