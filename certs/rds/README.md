# AWS RDS CA Certificates

These certificates are downloaded from https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/UsingWithRDS.SSL.html#UsingWithRDS.SSL.CertificatesDownload to establish TLS connectivity with RDS databases.
To support new regions, you need to download the region's CA cert and add it to the `rds-<region>-bundle.pem` file.