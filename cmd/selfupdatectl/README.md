# selfupdatectl: Manage deployment of self updating Fyne programs [![godoc reference](https://godoc.org/github.com/fynelabs/self-update?status.png)](https://godoc.org/github.com/fynelabs/self-update)

_selfupdatectl_ provide the ability to generate private and public key that can be used to sign and verify Fyne program to finally upload them to a S3 bucket.

To install _selfupdatectl_ you should do:
```
go install github.com/fynelabs/selfupdate/cmd/selfupdatectl
```

## _selfupdatectl create-keys_ and _selfupdatectl print-key_

Calling `selfupdatectl create-keys` will result in the generation of a brand new private key `ed25519.key` and public key `ed25519.pem`. It is recommended to absolutely not store your private key (`ed25519.key`) in your git repository. If you are using a password manager, you should use it to store this private key. This keys are not password protected at the moment and you should be careful with them.

As you need to specify your public key to use selfupdate managed API, the easiest way is by using:
```
$ selfupdatectl print-key
publicKey := ed25519.PublicKey{178, 103, 83, 57, 61, 138, 18, 249, 244, 80, 163, 162, 24, 251, 190, 241, 11, 168, 179, 41, 245, 27, 166, 70, 220, 254, 118, 169, 101, 26, 199, 129}
```

And copy&paste the resulting line in your source code where you are instantiating **selfupdate.Manage**.

## _selfupdatectl sign myprogram ..._

To sign an executable, you first need to create the private key necessary for that as described in the previous session. Once you do have the private key, you can do to sign your binary **myprogram** the following command: `selfupdatectl sign myprogram`.

This will generate a file named **myprogram.ed25519** of size 64 bytes that contain the signature of your binary.

## _selfupdatectl check myprogram ..._

To verify that your binary was properly signed, just call `selfupdatectl check myprogram`. It will error if there is a problem with your signature.

## _selfupdatectl upload myprogram ..._

Finaly, `selfupdatectl upload myprogram` will sign **myprogram** if necessary and upload it along with its signature to a S3 bucket. If no parameter are specified, it will try to read AWS information from configuration file and environment variable. Usually you would need to set *$AWS_S3_REGION* and *$AWS_S3_BUCKET* to match your need.

We also by default will upload and rename **myprogram** following the template `{{.Executable}}-{{.OS}}-{{.Arch}}{{.Ext}}` to make sure that your setup is cross platform. The same template can be picked up by the selfupdate.Manage API. If you pass an empty string as a template, no change to **myprogram** name will be applied.

