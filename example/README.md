## Build Jobs
```shell
# output: dist/*
datax-cli -j job -t template/mysql2mysql.template.json -o dist
```

## Run Jobs
```shell
# log: logs/*
datax-cli -j dist -e env/default.properties
```