# gzs

This is an attempt to extract my take on [soci](https://github.com/awslabs/soci-snapshotter) from my [dag.dev repo](https://github.com/jonjohnsonjr/dag.dev/tree/main/internal/soci).

This mostly exists to avoid shelling out or using cgo.
If you're looking for a tool that does this kind of thing, you are probably better off using [`gztool`](https://github.com/circulosmeos/gztool).

The interfaces and data formats in this are suboptimal and will change over time.

I've included a scrappy `gzs` tool under `cmd/gzs` that implements `index`, `ls`, and `cat` to verify that it kind of works.

## Demo

```console
$ # Grab a nice tar.gz from the internet.
$ crane blob ubuntu@sha256:01007420e9b005dc14a8c8b0f996a2ad8e0d4af6c3d01e62f123be14fe48eec7 > ubuntu.tar.gz

$ # Create an index of it.
$ gzs index ubuntu.tar.gz > ubuntu.gzs
2024/03/04 19:24:35 wrote 19 checkpoints for 3518 files

$ # Extract os-release in the usual way with tar.
$ time tar -Oxf ubuntu.tar.gz usr/lib/os-release
PRETTY_NAME="Ubuntu 22.04.3 LTS"
NAME="Ubuntu"
VERSION_ID="22.04"
VERSION="22.04.3 LTS (Jammy Jellyfish)"
VERSION_CODENAME=jammy
ID=ubuntu
ID_LIKE=debian
HOME_URL="https://www.ubuntu.com/"
SUPPORT_URL="https://help.ubuntu.com/"
BUG_REPORT_URL="https://bugs.launchpad.net/ubuntu/"
PRIVACY_POLICY_URL="https://www.ubuntu.com/legal/terms-and-policies/privacy-policy"
UBUNTU_CODENAME=jammy
tar -Oxf ubuntu.tar.gz usr/lib/os-release  0.14s user 0.01s system 95% cpu 0.161 total

$ # Extract os-release with the help of the index we created.
$ time gzs cat usr/lib/os-release ubuntu.tar.gz ubuntu.gzs
PRETTY_NAME="Ubuntu 22.04.3 LTS"
NAME="Ubuntu"
VERSION_ID="22.04"
VERSION="22.04.3 LTS (Jammy Jellyfish)"
VERSION_CODENAME=jammy
ID=ubuntu
ID_LIKE=debian
HOME_URL="https://www.ubuntu.com/"
SUPPORT_URL="https://help.ubuntu.com/"
BUG_REPORT_URL="https://bugs.launchpad.net/ubuntu/"
PRIVACY_POLICY_URL="https://www.ubuntu.com/legal/terms-and-policies/privacy-policy"
UBUNTU_CODENAME=jammy
gzs cat usr/lib/os-release ubuntu.tar.gz ubuntu.gzs  0.04s user 0.01s system 98% cpu 0.047 total
```

Notably, `gzs` can extract this file in 1/3 the time as `tar` by parsing the index and seeking forward to the nearest checkpoint.
For small local blobs, this doesn't make a huge difference, but for large blobs over the network with metered egress, this can save a lot of time and money.
