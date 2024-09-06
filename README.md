# sipnetworkcapture-go

การรัน Go program เป็น service บน Linux มีหลายวิธี แต่วิธีที่แนะนำคือการใช้ systemd ซึ่งเป็นระบบ init และ service manager ที่ใช้กันอย่างแพร่หลายใน Linux distributions สมัยใหม่ ต่อไปนี้เป็นขั้นตอนการตั้งค่า:

สร้าง executable จาก Go program ของคุณ:
Copygo build -o myapp main.go

ย้าย executable ไปยังตำแหน่งที่เหมาะสม:
Copysudo mv myapp /usr/local/bin/

สร้างไฟล์ service unit สำหรับ systemd:
Copysudo nano /etc/systemd/system/myapp.service

เพิ่มเนื้อหาต่อไปนี้ในไฟล์ myapp.service:
Copy[Unit]
Description=My Go Application
After=network.target

[Service]
ExecStart=/usr/local/bin/myapp
Restart=always
User=nobody
Group=nogroup
Environment="PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

[Install]
WantedBy=multi-user.target
แก้ไขส่วนต่างๆ ตามความเหมาะสม เช่น Description, ExecStart, User, Group
บันทึกไฟล์และออกจาก editor
รีโหลด systemd เพื่อให้อ่านไฟล์ service unit ใหม่:
Copysudo systemctl daemon-reload

เริ่มต้น service:
Copysudo systemctl start myapp

ตั้งค่าให้ service เริ่มต้นอัตโนมัติเมื่อระบบบูต:
Copysudo systemctl enable myapp

ตรวจสอบสถานะของ service:
Copysudo systemctl status myapp


คำสั่งอื่นๆ ที่มีประโยชน์:

หยุด service: sudo systemctl stop myapp
รีสตาร์ท service: sudo systemctl restart myapp
ดู logs: sudo journalctl -u myapp