import smtplib
from email.mime.text import MIMEText

msg = MIMEText("Test ticket content\n\nThis should create a ticket.")
msg['From'] = 'customer@example.com'
msg['To'] = 'tenant-mycompany@yourmailserver.com'
msg['Subject'] = 'Urgent: Server is down'

try:
    with smtplib.SMTP('localhost', 25) as server:
        server.send_message(msg)
    print("✅ Email sent successfully")
except Exception as e:
    print(f"❌ Error: {e}")
