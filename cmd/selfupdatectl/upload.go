package main

import (
	"bytes"
	"fmt"
	"log"
	"runtime"

	"github.com/fynelabs/selfupdate/cmd/selfupdatectl/internal/cloud"
	"github.com/urfave/cli/v2"
)

func upload() *cli.Command {
	var endpoint string
	var region string
	var bucket string
	var akid string
	var secret string
	var baseS3Path string
	var template string

	a := &application{}

	return &cli.Command{
		Name:        "upload",
		Usage:       "Upload an executable which has been properly signed to S3",
		Description: "The executable specified will get its signature generated and checked before being uploaded to a S3 bucket than can be optionally specified.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "private-key",
				Aliases:     []string{"priv"},
				Usage:       "The private key file to store the new key in.",
				Destination: &a.privateKey,
				Value:       "ed25519.key",
			},
			&cli.StringFlag{
				Name:        "public-key",
				Aliases:     []string{"pub"},
				Usage:       "The public key file to store the new key in.",
				Destination: &a.publicKey,
				Value:       "ed25519.pem",
			},
			&cli.StringFlag{
				Name:        "aws-endpoint",
				Aliases:     []string{"e"},
				Usage:       "AWS endpoint to connect to (can be used to connect to non AWS S3 services)",
				EnvVars:     []string{"AWS_S3_ENDPOINT"},
				Destination: &endpoint,
			},
			&cli.StringFlag{
				Name:        "aws-region",
				Aliases:     []string{"r"},
				Usage:       "AWS region to connect to",
				EnvVars:     []string{"AWS_S3_REGION"},
				Destination: &region,
			},
			&cli.StringFlag{
				Name:        "aws-bucket",
				Aliases:     []string{"b"},
				Usage:       "AWS bucket to store data into",
				EnvVars:     []string{"AWS_S3_BUCKET"},
				Destination: &bucket,
			},
			&cli.StringFlag{
				Name:        "aws-secret",
				Aliases:     []string{"s"},
				Usage:       "AWS secret to use to establish S3 connection",
				Destination: &secret,
			},
			&cli.StringFlag{
				Name:        "aws-AKID",
				Aliases:     []string{"a"},
				Usage:       "AWS Access Key ID to use to establish S3 connection",
				Destination: &akid,
			},
			&cli.StringFlag{
				Name:        "aws-base-s3-path",
				Aliases:     []string{"path", "p"},
				Usage:       "Specify the sub path in which the executable will be uploaded",
				Destination: &baseS3Path,
			},
			&cli.StringFlag{
				Name:        "template",
				Aliases:     []string{"t"},
				Usage:       "Specify the pattern to use for the executable once uploaded",
				Destination: &template,
				Value:       "{{.Executable}}-{{.OS}}-{{.Arch}}{{.Ext}}",
			},
		},
		Action: func(ctx *cli.Context) error {
			if ctx.Args().Len() == 0 {
				return fmt.Errorf("at least one executable to upload")
			}

			log.Println("Connecting to AWS")
			aws, err := cloud.NewAWSSession(akid, secret, endpoint, region, bucket)
			if err != nil {
				return err
			}

			ext := ""
			if runtime.GOOS == "windows" {
				ext = ".exe"
			}

			p := platform{
				OS:   runtime.GOOS,
				Arch: runtime.GOARCH,
				Ext:  ext,
			}

			var t *template.Template

			if binaryPattern != "" {
				t, err = template.New("platform").Parse(binaryPattern)
				if err != nil {
					return err
				}
			}

			for _, exe := range ctx.Args().Slice() {
				if runtime.GOOS == "windows" {
					p.Executable = exe[:len(exe)-len(".exe")]
				} else {
					p.Executable = exe
				}

				s3path := ""
				if baseS3Path != "" {
					s3path = baseS3Path + "/"
				}

				if binaryPattern != "" {
					buf := &bytes.Buffer{}
					err = t.Execute(buf, p)
					if err != nil {
						return err
					}
					s3path += buf.String()
				} else {
					s3path += exe
				}

				err = a.upload(aws, exe, s3path)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
}

type platform struct {
	OS         string
	Arch       string
	Ext        string
	Executable string
}

func (a *application) upload(aws *cloud.AWSSession, executable string, destination string) error {
	if a.check(executable) != nil {
		if err := a.sign(executable); err != nil {
			return err
		}
		if err := a.check(executable); err != nil {
			return err
		}
	}

	err := aws.UploadFile(executable, destination)
	if err != nil {
		return err
	}
	fmt.Println()

	defer fmt.Println()
	return aws.UploadFile(executable+".ed25519", destination+".ed25519")
}
