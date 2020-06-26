# Diun v0 to v1

Some fields in configuration file has been changed:

* `registries` renamed `regopts`
* `items` renamed `image`
* `items[].image` renamed `image[].name`
* `items[].registry_id` renamed `image[].regopts_id`
* `watch.os` and `watch.arch` moved to `image[].os` and `image[].arch`

!!! example "v0"
    ```yaml
    watch:
      os: linux
      arch: amd64
    
    registries:
      someregistryoptions:
        username: foo
        password: bar
        timeout: 20
    
    items:
      - image: docker.io/crazymax/nextcloud:latest
        registry_id: someregistryoptions
    ```

!!! example "v1"
    ```yaml
    regopts:
      someregistryoptions:
        username: foo
        password: bar
        timeout: 20
    
    image:
      - name: docker.io/crazymax/nextcloud:latest
        regopts_id: someregistryoptions
        os: linux
        arch: amd64
    ```
