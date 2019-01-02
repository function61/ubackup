[![Build Status](https://img.shields.io/travis/function61/ubackup.svg?style=for-the-badge)](https://travis-ci.org/function61/ubackup)
[![Download](https://img.shields.io/bintray/v/function61/ubackup/main.svg?style=for-the-badge&label=Download)](https://bintray.com/function61/ubackup/main/_latestVersion#files)

What
----

µbackup takes backups from your Docker containers 100 % automatically and uploads them to S3.

Stateful containers are gross, but there are use cases where you need them.

```
+------------+     +-----------------------------+      +------------------------------+      +--------------+
|            |     |                             |      |                              |      |              |
| Once a day +-----> For each container:         +------> Compress all containers'     +------> Upload to S3 |
|            |     | - if BACKUP_COMMAND defined |      | backups into .tar.gz archive |      |              |
+------------+     |                             |      |                              |      +--------------+
                   +-----------+-----^-----------+      +------------------------------+
                               |     |
                               |     |
                               |     |
                 +-------------v-----+---------------+
                 |                                   |
                 |  Execute BACKUP_COMMAND inside    |
                 |  the container, taking its stdout |
                 |  as the backup stream             |
                 |                                   |
                 +-----------------------------------+
```

`BACKUP_COMMAND` is an ENV variable.

This simple approach is suprprisingly flexible and its streaming approach is more efficient
than having to write temporary files.

If you need to backup a single file inside a container, use: `BACKUP_COMMAND=cat /yourfile.db`

For PostgreSQL, you could use: `BACKUP_COMMAND=pg_dump -U postgres f61`

For a directory, you could use `BACKUP_COMMAND=tar -cC /yourdirectory -f - .`


How to use
----------

```
$ mkdir ~/ubackup && cd ~/ubackup/
$ VERSION_TO_DOWNLOAD="..." # find this from Bintray. Looks like: 20180828_1449_b9d7759cf80f0b4a
$ sudo curl --location --fail --output ubackup "https://dl.bintray.com/function61/ubackup/$VERSION_TO_DOWNLOAD/ubackup_linux-amd64" && sudo chmod +x ubackup
$ ./ubackup print-default-config > config.json
$ vim config.json
$ ./ubackup scheduler install-systemd-service-file
Wrote unit file to /etc/systemd/system/ubackup.service
Run to enable on boot & to start now:
        $ systemctl enable ubackup
        $ systemctl start ubackup
        $ systemctl status ubackup
```

Currently this is offered as a binary that you'll pluck into your server nodes. It would
not be hard to distribute this as a system-level Docker service (= runs on every node),
but that is not implemented yet.


IAM policy
----------

You should minimize attack surface by only allowing the backup program to put stuff into
the bucket. Read access is not required. If you want the bucket to automatically delete
old backups, the backup program should not do it but you should use
[S3 lifecycle policies](https://docs.aws.amazon.com/AmazonS3/latest/dev/object-lifecycle-mgmt.html)
instead to make AWS remove your old backups.

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:PutObject",
                "s3:PutObjectAcl"
            ],
            "Resource": [
                "arn:aws:s3:::YOURBUCKET/*"
            ]
        }
    ]
}
```
