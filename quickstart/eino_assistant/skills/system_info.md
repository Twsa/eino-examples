---
name: system_info
description: 获取详细的系统版本、内核、磁盘和负载信息
---

echo "--- OS Release ---"
cat /etc/os-release
echo "--- Kernel Info ---"
uname -a
echo "--- Uptime ---"
uptime
echo "--- Disk Usage ---"
df -h /
