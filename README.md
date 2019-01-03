[![Build Status](https://img.shields.io/travis/function61/ubackup.svg?style=for-the-badge)](https://travis-ci.org/function61/ubackup)
[![Download](https://img.shields.io/bintray/v/function61/ubackup/main.svg?style=for-the-badge&label=Download)](https://bintray.com/function61/ubackup/main/_latestVersion#files)

What
----

µbackup takes backups from your Docker containers 100 % automatically, properly encrypts
(more on this in this README) and uploads them to S3.

Stateful containers are gross, but there are use cases where you need them.

```
+------------+     +-----------------------------+      +------------------------------+      +--------------+
|            |     |                             |      |                              |      |              |
| Once a day +-----> For each container:         +------> Compress all containers'     +------> Upload to S3 |
|            |     | - if BACKUP_COMMAND defined |      | backups into .zip archive    |      |              |
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

`BACKUP_COMMAND` is a container's ENV variable (specified during
`$ docker run -e "BACKUP_COMMAND=..."` for example) that contains the command used to take
a backup of the important state inside the container.

If you need to backup a single file inside a container, use: `BACKUP_COMMAND=cat /yourfile.db`

For PostgreSQL, you could use: `BACKUP_COMMAND=pg_dump -U postgres nameOfYourDatabase`

For a directory, you could use: `BACKUP_COMMAND=tar -cC /yourdirectory -f - .` (`-` means
`$ tar` will write the archive to `stdout`, `.` just means to process all files in the
selected directory)

This simple approach is suprprisingly flexible and its streaming approach is more efficient
than having to write temporary files. µbackup shoves all containers' backups in a single
compressed .zip file.

You don't have to compress the file inside the container, since that is taken care of for you.


Security & encryption
---------------------

The backups are encrypted with per-backup 256-bit AES (in CTR mode with proper IV)
key, which itself is asymmetrically encrypted with 4096-bit RSA public key. This means
that immediately after the backup is complete, µbackup forgets/loses access to the actual
encryption key, and only the user holding the private key will be able to decrypt the
backup. This way the servers nor Amazon can ever access your backups.

If you are serious about security, with this design you could even store the private key
in a [YubiKey](https://www.yubico.com/) (or some other form of HSM).


How to use
----------

```
# download the program

$ mkdir ~/ubackup && cd ~/ubackup/
$ VERSION_TO_DOWNLOAD="..." # find this from Bintray. Looks like: 20180828_1449_b9d7759cf80f0b4a
$ curl --location --fail --output ubackup "https://dl.bintray.com/function61/ubackup/$VERSION_TO_DOWNLOAD/ubackup_linux-amd64" && sudo chmod +x ubackup

# create encryption & decryption keys
# (for security you should not actually ever store the decryption key on the same machine
#  that takes the backups, but this is provided for demonstration purposes)

$ ./ubackup decryption-key-generate > backups.key
$ ./ubackup decryption-key-to-encryption-key < backups.key > backups.pub

# create configuration file stub (and embed encryption key in the config)

$ ./ubackup print-default-config --pubkey-file backups.pub > config.json

# edit the configuration further (specify your S3 bucket details)

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
but that is not implemented yet. See [#6](https://github.com/function61/ubackup/issues/6)


Restoring from backup
---------------------

Download the `<HOSTNAME>/backup-<TIMESTAMP>.zip.aes` file from your S3 bucket.

The `decrypt` verb of µbackup requires path to your decryption key, reads the encrypted
backup file from stdin and outputs the decrypted file to stdout.

```
./ubackup decrypt backups.key < backup-2019-01-03_1340.zip.aes > backup-2019-01-03_1340.zip
```

You could even pipe output directly to unzip and exclude decompression of files that you
are not interested in, so the backup restore is faster because unnecessary files don't
have to be written to disk!


IAM policy
----------

You should minimize attack surface by only allowing the backup program to put stuff into
the bucket, since read access is not required. If you want the bucket to automatically delete
old backups, the backup program should not do it but you should use
[S3 lifecycle policies](https://docs.aws.amazon.com/AmazonS3/latest/dev/object-lifecycle-mgmt.html)
instead to make AWS remove your old backups.

You should also consider enabling bucket versioning so that if an attacker gained
credentials used by this backup program, she cannot permantently destroy the backups by
overwriting old backups with empty content (`s3:PutObject` can do that). Versioning would
allow you to recover these tampered files in this described scenario.

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
