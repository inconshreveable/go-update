package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"runtime"

	"github.com/fynelabs/selfupdate/cmd/selfupdatectl/internal/cloud"
	"github.com/urfave/cli/v2"
)

type awsConfig struct {
	endpoint   string
	region     string
	bucket     string
	akid       string
	secret     string
	baseS3Path string
}

func s3upload() *cli.Command {
	var exeTemplate string
	var goos string
	var goarch string

	a := &application{}
	config := &awsConfig{}

	return &cli.Command{
		Name:        "s3upload",
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
				Destination: &config.endpoint,
			},
			&cli.StringFlag{
				Name:        "aws-region",
				Aliases:     []string{"r"},
				Usage:       "AWS region to connect to",
				EnvVars:     []string{"AWS_S3_REGION"},
				Destination: &config.region,
			},
			&cli.StringFlag{
				Name:        "aws-bucket",
				Aliases:     []string{"b"},
				Usage:       "AWS bucket to store data into",
				EnvVars:     []string{"AWS_S3_BUCKET"},
				Destination: &config.bucket,
			},
			&cli.StringFlag{
				Name:        "aws-secret",
				Aliases:     []string{"s"},
				Usage:       "AWS secret to use to establish S3 connection",
				Destination: &config.secret,
			},
			&cli.StringFlag{
				Name:        "aws-AKID",
				Aliases:     []string{"a"},
				Usage:       "AWS Access Key ID to use to establish S3 connection",
				Destination: &config.akid,
			},
			&cli.StringFlag{
				Name:        "aws-base-s3-path",
				Aliases:     []string{"path", "p"},
				Usage:       "Specify the sub path in which the executable will be uploaded",
				Destination: &config.baseS3Path,
			},
			&cli.StringFlag{
				Name:        "template",
				Aliases:     []string{"t"},
				Usage:       "Specify the pattern to use for the executable once uploaded",
				Destination: &exeTemplate,
				Value:       "{{.Executable}}-{{.OS}}-{{.Arch}}{{.Ext}}",
			},
			&cli.StringFlag{
				Name:        "executable-os",
				Aliases:     []string{"os"},
				Usage:       "Allow to override the runtime OS to ease cross compilation use case",
				Destination: &goos,
				Value:       runtime.GOOS,
			},
			&cli.StringFlag{
				Name:        "executable-arch",
				Aliases:     []string{"arch"},
				Usage:       "Allow to override the runtime ARCH to ease cross compilation use case",
				Destination: &goarch,
				Value:       runtime.GOARCH,
			},
		},
		Action: func(ctx *cli.Context) error {
			if ctx.Args().Len() != 1 {
				return fmt.Errorf("only one executable should be specified to be uploaded")
			}

			log.Println("Connecting to AWS")
			aws, err := cloud.NewAWSSession(config.akid, config.secret, config.endpoint, config.region, config.bucket)
			if err != nil {
				return err
			}

			exe := ctx.Args().First()

			targetExe, err := buildTargetExecutable(exeTemplate, exe, goos, goarch)
			if err != nil {
				return err
			}

			s3path := buildS3Path(config.baseS3Path, targetExe)

			err = a.s3upload(aws, exe, s3path)
			if err != nil {
				return err
			}
			return nil
		},
	}
}

func s3uploads() *cli.Command {
	a := &application{}
	config := &awsConfig{}

	return &cli.Command{
		Name:        "s3uploads",
		Usage:       "Upload multiple executable which has been properly signed to a S3 path",
		Description: "The executable specified will get their signature generated and checked before being uploaded to a S3 bucket location specified as the last arguments.",
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
				Destination: &config.endpoint,
			},
			&cli.StringFlag{
				Name:        "aws-region",
				Aliases:     []string{"r"},
				Usage:       "AWS region to connect to",
				EnvVars:     []string{"AWS_S3_REGION"},
				Destination: &config.region,
			},
			&cli.StringFlag{
				Name:        "aws-bucket",
				Aliases:     []string{"b"},
				Usage:       "AWS bucket to store data into",
				EnvVars:     []string{"AWS_S3_BUCKET"},
				Destination: &config.bucket,
			},
			&cli.StringFlag{
				Name:        "aws-secret",
				Aliases:     []string{"s"},
				Usage:       "AWS secret to use to establish S3 connection",
				Destination: &config.secret,
			},
			&cli.StringFlag{
				Name:        "aws-AKID",
				Aliases:     []string{"a"},
				Usage:       "AWS Access Key ID to use to establish S3 connection",
				Destination: &config.akid,
			},
		},
		Action: func(ctx *cli.Context) error {
			if ctx.Args().Len() > 2 {
				return fmt.Errorf("at least one executable and a S3 target path need to be specified")
			}

			log.Println("Connecting to AWS")
			aws, err := cloud.NewAWSSession(config.akid, config.secret, config.endpoint, config.region, config.bucket)
			if err != nil {
				return err
			}

			config.baseS3Path = ctx.Args().Slice()[ctx.Args().Len()-1]

			for _, exe := range ctx.Args().Slice()[:ctx.Args().Len()-1] {
				s3path := buildS3Path(config.baseS3Path, exe)

				err = a.s3upload(aws, exe, s3path)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func (a *application) s3upload(aws *cloud.AWSSession, executable string, destination string) error {
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

type platform struct {
	OS         string
	Arch       string
	Ext        string
	Executable string
}

func newPlatform(exe string, goos string, goarch string) *platform {
	ext := ""
	if goos == "windows" {
		ext = ".exe"
	}

	executable := exe
	if goos == "windows" {
		executable = exe[:len(exe)-len(".exe")]
	}

	return &platform{
		OS:         goos,
		Arch:       goarch,
		Ext:        ext,
		Executable: executable,
	}
}

func buildS3Path(baseS3Path string, exe string) string {
	s3path := ""
	if baseS3Path != "" {
		s3path = baseS3Path
		if baseS3Path[len(baseS3Path)-1] != '/' {
			s3path += "/"
		}
	}
	s3path += exe

	return s3path
}

func buildTargetExecutable(exeTemplate string, exe string, goos string, goarch string) (string, error) {
	var t *template.Template

	if exeTemplate != "" {
		var err error

		t, err = template.New("platform").Parse(exeTemplate)
		if err != nil {
			return "", err
		}
	}

	p := newPlatform(exe, goos, goarch)

	targetExe := exe
	if exeTemplate != "" {
		buf := &bytes.Buffer{}
		err := t.Execute(buf, p)
		if err != nil {
			return "", err
		}
		targetExe = buf.String()
	}

	return targetExe, nil
}
