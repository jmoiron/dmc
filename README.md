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
```

dmc runs everything in parallel but prints the full output for each system as
it becomes available, making it fast but also easy to read.
