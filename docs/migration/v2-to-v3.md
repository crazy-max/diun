# Diun v2 to v3

## File provider

`static` provider has been renamed `file`. This now allows the static configuration to be declared in one or more files to avoid overloading the current configuration file and also dynamic updating.

!!! example "v2"
    ```yaml
    providers:
      static:
        - name: docker.io/crazymax/diun
          watch_repo: true
          max_tags: 10
    ```

!!! example "v3"
    ```yaml
    providers:
      file:
        # Watch images from filename /path/to/config.yml
        filename: /path/to/config.yml
        # OR watch images from directory /path/to/config/folder
        directory: /path/to/config/folder
    ```
    ```yaml
    # /path/to/config.yml
    - name: docker.io/crazymax/diun
      watch_repo: true
      max_tags: 10
    ```

## Allow only one Docker and Swarm provider

Now you can declare only one Docker and/or Swarm provider. This is due to a limitation of the Docker engine.

!!! example "v2"
    ```yaml
    providers:
      docker:
        mydocker:
          watch_stopped: true
      swarm:
        myswarm:
          watch_by_default: true
    ```

!!! example "v3"
    ```yaml
    providers:
      docker:
        watch_stopped: true
      swarm:
        watch_by_default: true
    ```

## Remove `enable` setting for notifiers

The `enable` entry has been removed for notifiers. If you don't want a notifier to be enabled, you must now remove or comment its configuration.

!!! example "v2"
    ```yaml
    notif:
      amqp:
        enable: false
        host: localhost
        port: 5672
      gotify:
        enable: true
        endpoint: http://gotify.foo.com
        token: Token123456
        priority: 1
        timeout: 10
    ```

!!! example "v3"
    ```yaml
    notif:
      gotify:
        endpoint: http://gotify.foo.com
        token: Token123456
        priority: 1
        timeout: 10
    ```
