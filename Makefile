###
# Deploy on Prod
###

buildProd:
	sudo systemctl stop preebot.service
	cd ~/git/preebot
	rm /usr/local/bin/preebot
	rm preebot
	go build -o preebot
	sudo cp -p preebot /usr/local/bin/.
	sudo systemd-analyze verify preebot.service
	sudo systemctl daemon-reload
	sudo systemctl start preebot.service
	sudo journalctl -f -u preebot