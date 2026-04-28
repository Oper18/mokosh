package config

import "github.com/urfave/cli/v2"

func init() {
	Flags = append(Flags,
		CliFlag{Flag: &cli.StringFlag{
			Name:    "aws-base-endpoint",
			Usage:   "S3 base `ENDPOINT` URL; leave empty for native AWS S3, e.g. https://fsn1.your-objectstorage.com for Hetzner",
			EnvVars: EnvVars("AWS_BASE_ENDPOINT"),
		}},
		CliFlag{Flag: &cli.StringFlag{
			Name:    "aws-region",
			Usage:   "S3 bucket `REGION`, e.g. us-east-1",
			EnvVars: EnvVars("AWS_REGION"),
		}},
		CliFlag{Flag: &cli.StringFlag{
			Name:    "aws-access-key-id",
			Usage:   "S3 access key `ID`",
			EnvVars: EnvVars("AWS_ACCESS_KEY_ID"),
		}},
		CliFlag{Flag: &cli.StringFlag{
			Name:    "aws-secret-access-key",
			Usage:   "S3 secret access `KEY`",
			EnvVars: EnvVars("AWS_SECRET_ACCESS_KEY"),
			Hidden:  true,
		}},
		CliFlag{Flag: &cli.StringFlag{
			Name:    "s3-files-bucket-name",
			Usage:   "S3 bucket `NAME` for storing originals; leave empty to use local storage",
			EnvVars: EnvVars("S3_FILES_BUCKET_NAME"),
		}},
	)
}
