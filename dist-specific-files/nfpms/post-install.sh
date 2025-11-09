echo "...Changing to executable"
chmod +x /opt/godocs/godocs
echo "...Changing permissions"
chown -R godocs:godocs /opt/godocs
echo "...Enabling systemd service"
systemctl enable godocs.service
echo "...Starting godocs service"
systemctl start godocs.service
