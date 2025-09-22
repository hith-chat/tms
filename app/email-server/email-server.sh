#!/bin/bash

echo "Testing Guerrilla Mail Server..."

# 1. Check if port is open
echo "1. Checking SMTP port..."
nc -zv tickets.hith.chat 25
nc -zv 91.99.186.50 25

# 2. Send test email
echo "2. Sending test email..."
python3 << EOF
import smtplib
from email.mime.text import MIMEText

msg = MIMEText("Test ticket content\n\nThis should create a ticket.")
msg['From'] = 'customer@example.com'
msg['To'] = 'tenant-mycompany@tickets.hith.chat'
msg['Subject'] = 'Urgent: Server is down'

try:
    with smtplib.SMTP('localhost', 25) as server:
        server.send_message(msg)
    print("✅ Email sent successfully")
except Exception as e:
    print(f"❌ Error: {e}")
EOF

# 3. Check logs
echo "3. Recent guerrilla-mail logs:"
docker-compose logs --tail=10 guerrilla-mail

echo "4. Recent backend logs:"
docker-compose logs --tail=10 backend