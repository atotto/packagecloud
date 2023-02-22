# packagecloud

simple packagecloud command line tool.

## Install

    go install github.com/atotto/packagecloud

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


## CircleCI

### Configure environment variables

Set the `$PACKAGECLOUD_TOKEN`

create an environment variable with the name `PACKAGECLOUD_TOKEN`, containing the value of a packagecloud API token.


### Use docker

```yaml
jobs:
  deploy:
    docker:
      - image: atotto/packagecloud
    steps:
      - attach_workspace:
          at: /workspace
      - run: |
          packagecloud push atotto/repository/any/any my-debian-package_*.deb
```

### Use debain package

```yaml
    steps:
      - run:
          name: setup atotto packagecloud
          command: |
            curl -s https://packagecloud.io/install/repositories/atotto/debian-utils/script.deb.sh | os=any dist=any bash
            apt install -y packagecloud

```
