dmc runs programs on other machines using ssh

usage:

```sh
$ cat mysql_master_hosts
10.0.0.1
10.0.0.2
10.0.0.3

$ cat mysql_master_hosts | dmc "grep DESCRIPTION /etc/lsb-release"
[10.0.0.1]$ grep DESCRIPTION /etc/lsb-release
DISTRIB_DESCRIPTION="Ubuntu 12.04.5 LTS"
[10.0.0.2]$ grep DESCRIPTION /etc/lsb-release
DISTRIB_DESCRIPTION="Ubuntu 12.04.5 LTS"
[10.0.0.3]$ grep DESCRIPTION /etc/lsb-release
DISTRIB_DESCRIPTION="Ubuntu 12.04.5 LTS"

$ dig master.pg.service.consul |grep -v '^;' |grep A
master.pg.service.consul. 10	IN	A 10.0.0.1
master.pg.service.consul. 10	IN	A	10.0.0.2

$ dmc -d master.pg.service.consul "grep DESCRIPTION /etc/lsb-release"
[10.0.0.1]$ grep DESCRIPTION /etc/lsb-release
DISTRIB_DESCRIPTION="Ubuntu 12.04.5 LTS"
[10.0.0.2]$ grep DESCRIPTION /etc/lsb-release
DISTRIB_DESCRIPTION="Ubuntu 12.04.5 LTS"

$ dmc -h
Usage of dmc:
  -d="": dns name for multi-hosts
  -hosts="": list of hosts
  -p="": prefix for command echo
  -v=false: verbose output
```

dmc runs all commands in parallel but prints the full output for each system as
it becomes available, making it fast but also easy to read.
