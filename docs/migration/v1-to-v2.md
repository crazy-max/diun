# Diun v1 to v2

`image` field has been moved to `providers.static` in configuration file:

!!! example "v1"
    ```yaml
    image:
      - name: docker.io/crazymax/diun
        watch_repo: true
        max_tags: 10
    ```

!!! example "v2"
    ```yaml
    providers:
      static:
        - name: docker.io/crazymax/diun
          watch_repo: true
          max_tags: 10
    ```
