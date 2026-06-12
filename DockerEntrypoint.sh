#!/bin/sh

# Start fail2ban with the Nova_x-ui ip-limit jail
if [ "$XUI_ENABLE_FAIL2BAN" = "true" ]; then
    LOG_FOLDER="${XUI_LOG_FOLDER:-/var/log/Nova_x-ui}"
    mkdir -p "$LOG_FOLDER"
    touch "$LOG_FOLDER/Nova_x-ui-iplimit.log" "$LOG_FOLDER/Nova_x-ui-iplimit-banned.log"

    mkdir -p /etc/fail2ban/jail.d /etc/fail2ban/filter.d /etc/fail2ban/action.d

    cat > /etc/fail2ban/jail.d/Nova_x-ui-iplimit.conf << EOF
[Nova_x-ui-iplimit]
enabled=true
backend=auto
filter=Nova_x-ui-iplimit
action=Nova_x-ui-iplimit
logpath=$LOG_FOLDER/Nova_x-ui-iplimit.log
maxretry=1
findtime=32
bantime=30m
EOF

    cat > /etc/fail2ban/filter.d/Nova_x-ui-iplimit.conf << 'EOF'
[Definition]
datepattern = ^%%Y/%%m/%%d %%H:%%M:%%S
failregex   = \[LIMIT_IP\]\s*Email\s*=\s*<F-USER>.+</F-USER>\s*\|\|\s*Disconnecting OLD IP\s*=\s*<ADDR>\s*\|\|\s*Timestamp\s*=\s*\d+
ignoreregex =
EOF

    cat > /etc/fail2ban/action.d/Nova_x-ui-iplimit.conf << EOF
[INCLUDES]
before = iptables-allports.conf

[Definition]
actionstart = <iptables> -N f2b-<name>
              <iptables> -A f2b-<name> -j <returntype>
              <iptables> -I <chain> -p <protocol> -j f2b-<name>

actionstop = <iptables> -D <chain> -p <protocol> -j f2b-<name>
             <actionflush>
             <iptables> -X f2b-<name>

actioncheck = <iptables> -n -L <chain> | grep -q 'f2b-<name>[ \t]'

actionban = <iptables> -I f2b-<name> 1 -s <ip> -j <blocktype>
            echo "\$(date +"%%Y/%%m/%%d %%H:%%M:%%S")   BAN   [Email] = <F-USER> [IP] = <ip> banned for <bantime> seconds." >> $LOG_FOLDER/Nova_x-ui-iplimit-banned.log

actionunban = <iptables> -D f2b-<name> -s <ip> -j <blocktype>
              echo "\$(date +"%%Y/%%m/%%d %%H:%%M:%%S")   UNBAN   [Email] = <F-USER> [IP] = <ip> unbanned." >> $LOG_FOLDER/Nova_x-ui-iplimit-banned.log

[Init]
name = default
protocol = tcp
chain = INPUT
EOF

    fail2ban-client -x start
fi

# Run Nova_x-ui
exec /app/Nova_x-ui
