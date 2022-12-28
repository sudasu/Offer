# Linux

## 对日志文件中出现的ip进行排序

cat log.txt | grep xxxx.xxx.xx.xx | uniq -c |sort -k 2 | head -n 10