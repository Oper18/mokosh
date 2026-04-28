package config

// AWS returns the AWS/S3 configuration built from the current options.
func (c *Config) AWS() AWSConfiguration {
	return AWSConfiguration{
		AWSBaseEndpoint:    c.options.AWSBaseEndpoint,
		AWSRegion:          c.options.AWSRegion,
		AWSAccessKeyID:     c.options.AWSAccessKeyID,
		AWSSecretAccessKey: c.options.AWSSecretAccessKey,
		S3FilesBucketName:  c.options.S3FilesBucketName,
	}
}

// S3Enabled reports whether S3 storage is configured — i.e. a bucket name
// and access credentials are all non-empty.
func (c *Config) S3Enabled() bool {
	o := c.options
	return o.S3FilesBucketName != "" && o.AWSAccessKeyID != "" && o.AWSSecretAccessKey != ""
}
