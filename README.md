# mackerel-plugin-nginx-cache

```
$ ./mackerel-plugin-nginx-cache -path /dev/shm/proxy_cache -size 1024m -kname cache -ksize 128m
nginx-cache.disk-cache.usage       128     1471097471
nginx-cache.disk-cache.size        1024    1471097471
nginx-cache.keys-cache.zone_usage  15      1471097471
nginx-cache.keys-cache.zone_size   128     1471097471
```

The Unit is megabyte.

## Usage

```
Usage of mackerel-plugin-nginx-cache:
-kname string
        proxy_cache_path $keys_zone_name
-ksize string
        proxy_cache_path $keys_zone_size
-path string
        proxy_cache_path $path
-size string
        proxy_cache_path $max_size
-tempfile string
        temporary file path
```
