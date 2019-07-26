[![Build Status](https://img.shields.io/travis/function61/ubackup.svg?style=for-the-badge)](https://travis-ci.org/function61/ubackup)
[![Download](https://img.shields.io/bintray/v/function61/dl/ubackup.svg?style=for-the-badge&label=Download)](https://bintray.com/function61/dl/ubackup/_latestVersion#files)
[![Download](https://img.shields.io/docker/pulls/fn61/ubackup.svg?style=for-the-badge)](https://hub.docker.com/r/fn61/ubackup/)

What
----

µbackup takes backups from your Docker containers 100 % automatically, properly encrypts
(more on this in this README) and uploads them to S3. µbackup is also an embeddable library
for taking & storing backups from your application.

Stateful containers are gross, but there are use cases where you need them.

```
+------------+     +-----------------------------+      +------------------------------+      +--------------+
|            |     |                             |      |                              |      |              |
| Once a day +-----> For each container:         +------> Compress & encrypt stdout    +------> Upload to S3 |
|            |     | - if BACKUP_COMMAND defined |      | of BACKUP_COMMAND            |      |              |
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

This simple approach is surprisingly flexible and its streaming approach is more efficient
than having to write temporary files.

You don't have to compress the file inside the container, since that is taken care of for you.


Security & encryption
---------------------

The backups are encrypted with per-backup 256-bit AES (in CTR mode with random IV)
key, which itself is asymmetrically encrypted with 4096-bit RSA public key. This means
that immediately after the backup is complete, µbackup forgets/loses access to the actual
encryption key, and only the user holding the private key will be able to decrypt the
backup. This way the servers nor Amazon can ever access your backups (if you store the
private key somewhere else).

If you are serious about security, with this design you could even store the private key
in a [YubiKey](https://www.yubico.com/) (or some other form of HSM).


How to use, as Docker service
-----------------------------

First, do the configuration steps from below section "How to use, binary installation".

You only need to do them to create correct `config.json` file. Now, convert that file for
embedding as ENV variable:

```
$ cat config.json | base64 -w 0
```

That value will be your `UBACKUP_CONF` ENV var.

Now, just deploy Docker service as global service in the cluster (= runs on every node):

```
$ docker service create --name ubackup --mode global \
	--mount type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock \
	-e UBACKUP_CONF=... \
	fn61/ubackup:VERSION
```

Note: `VERSION` is the same as you would find for the binary installation.


How to use, binary installation
-------------------------------

Download:

```
$ mkdir ~/ubackup && cd ~/ubackup/
$ VERSION_TO_DOWNLOAD="..." # find this from Bintray. Looks like: 20180828_1449_b9d7759cf80f0b4a
$ curl --location --fail --output ubackup "https://dl.bintray.com/function61/dl/ubackup/$VERSION_TO_DOWNLOAD/ubackup_linux-amd64" && sudo chmod +x ubackup
```

Create encryption & decryption keys:

```
$ ./ubackup decryption-key-generate > backups.key
$ ./ubackup decryption-key-to-encryption-key < backups.key > backups.pub
```

For security you should not actually ever store the decryption key on the same machine
that takes the backups, but this is provided for demonstration purposes.

Create configuration file stub (and embed encryption key in the config):

```
$ ./ubackup print-default-config --pubkey-file backups.pub > config.json
```

Edit the configuration further (specify your S3 bucket details)

```
$ vim config.json
```

Install & start the service:

```
$ ./ubackup scheduler install-systemd-service-file
Wrote unit file to /etc/systemd/system/ubackup.service
Run to enable on boot & to start now:
        $ systemctl enable ubackup
        $ systemctl start ubackup
        $ systemctl status ubackup
```


Restoring from backup
---------------------

Remember to test your backup recovery! Nobody actually wants backups, but everybody wants
a restore. Without disaster recovery drills you don't know if your backups work.

Download the `.gz.aes` backup file from your S3 bucket. You can also do it from µbackup CLI
(just give the service ID the backup was stored under - in this example it's `varasto`):

```
$ ubackup storage ls varasto
varasto/2019-07-25 0830Z_joonas_10028.gz.aes
varasto/2019-07-26 0825Z_joonas_10028.gz.aes
```

The lines output are "backup ID"s, which is enough info to download the backup:

```
$ ubackup storage get 'varasto/2019-07-26 0825Z_joonas_10028.gz.aes' > '2019-07-26 0825Z_joonas_10028.gz.aes'
```

Now you have the encrypted and compressed file - you still have to decrypt it.

The `decrypt` verb of µbackup requires path to your decryption key, reads the encrypted
backup file from stdin and outputs the decrypted file to stdout.

```
./ubackup decrypt backups.key < '2019-07-26 0825Z_joonas_10028.gz.aes' > '2019-07-26 0825Z_joonas_10028'
```


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
allow you to recover these tampered files in this described scenario. This effectively
implements "ransomware protection".

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:PutObject",
                "s3:GetObject",
                "s3:PutObjectAcl"
            ],
            "Resource": "arn:aws:s3:::YOURBUCKET/*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "s3:ListBucket"
            ],
            "Resource": "arn:aws:s3:::YOURBUCKET*"
        }
    ]
}
```
