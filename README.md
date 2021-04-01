# packagecloud

simple packagecloud command line tool.

## Install

    go get github.com/atotto/packagecloud

or

    curl -s https://packagecloud.io/install/repositories/atotto/debian-utils/script.deb.sh | os=any dist=any sudo -E bash
    sudo apt install packagecloud

## Usage

### Pushing a package

    packagecloud push example-user/example-repository/ubuntu/xenial /tmp/example.deb
    
### Deleting a package

    packagecloud rm example-user/example-repository/ubuntu/xenial example_1.0.1-1_amd64.deb

### Promoting packages between repositories

    packagecloud promote example-user/repo1/ubuntu/xenial example_1.0-1_amd64.deb example-user/repo2
