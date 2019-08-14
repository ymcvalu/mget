# mget
a multi-thread file downloader

### Usage
```redshift
$ mget -n 10 -f xxx.tar.gz https://www.xxx.com/xxx/xxx/xxx.tar.gz
```
- -n: how many thread(goroutine actually) to use
- -f: specify the path to save