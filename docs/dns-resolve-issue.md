# **Nomad Container DNS Failure due to `netclient` Port Conflict**

### **Overview**

Nomad containers suddenly lost outbound internet access after a network or DNS change on the host.
Root cause: the `netclient` service was binding to port **53** (DNS) on the host’s WireGuard interface, conflicting with `dnsmasq` — the main DNS forwarder used by Nomad and Docker bridges.

---

### **Symptoms**

* Containers could ping IPs (e.g., `8.8.8.8`) but failed to resolve hostnames (`curl google.com` hung or timed out).
* `/etc/resolv.conf` inside containers showed:

  ```
  nameserver 172.17.0.1
  ```
* `dnsmasq` appeared healthy but DNS queries still failed.

---

### **Diagnosis Steps**

1. **Check which service owns port 53**

   ```bash
   sudo ss -lntup | grep :53
   ```

   Example output:

   ```
   udp   UNCONN 0 0 10.10.85.1:53  0.0.0.0:*  users:(("dnsmasq",pid=1915740))
   udp   UNCONN 0 0 10.10.85.1:53  0.0.0.0:*  users:(("netclient",pid=2268894))
   ```

   This indicates a port conflict — both `dnsmasq` and `netclient` trying to listen on `10.10.85.1:53`.

2. **Verify `dnsmasq` is configured correctly**

   * Located at `/etc/dnsmasq.d/10-consul`
   * Confirms it listens on `127.0.0.1`, `172.17.0.1`, and `10.10.85.1`.

3. **Identify the conflicting service**

   ```bash
   systemctl status netclient
   ```

---

### **Resolution**

To permanently fix the conflict and restore DNS for Nomad containers:

```bash
sudo systemctl stop netclient
sudo systemctl disable netclient
sudo systemctl restart dnsmasq
```

Verify that only `dnsmasq` owns port 53:

```bash
sudo ss -lntup | grep :53
```

Expected output:

```
udp   UNCONN 0 0 127.0.0.1:53   0.0.0.0:*  users:(("dnsmasq",pid=xxxx))
udp   UNCONN 0 0 172.17.0.1:53  0.0.0.0:*  users:(("dnsmasq",pid=xxxx))
udp   UNCONN 0 0 10.10.85.1:53  0.0.0.0:*  users:(("dnsmasq",pid=xxxx))
```

Then test from inside a Nomad container:

```bash
nomad alloc exec -i -t <alloc_id> nslookup google.com
```

DNS resolution should now succeed instantly.

---

### **Optional (if netclient must stay enabled)**

If `netclient` is required for VPN routing, you can instead:

* Edit `/etc/netclient/config.yaml` (or its config file)

  ```yaml
  dns_listen: false
  ```

  or change its port:

  ```yaml
  dns_port: 5353
  ```
* Restart:

  ```bash
  sudo systemctl restart netclient dnsmasq
  ```

---

### **Summary**

| Problem                                  | Cause                                        | Fix                                                     |
| ---------------------------------------- | -------------------------------------------- | ------------------------------------------------------- |
| Nomad containers can’t resolve DNS       | `netclient` hijacked port 53 on WireGuard IP | `sudo systemctl disable netclient` or move its DNS port |
| dnsmasq errors: “Address already in use” | Port conflict on 10.10.85.1                  | Same as above                                           |
| Works after restart, breaks again        | `netclient` auto-starts                      | Disable permanently with `systemctl disable`            |
