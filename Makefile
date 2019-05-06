build-deb:
	env GOOS=linux GOARCH=amd64 go build -o kintoun_linux_amd64 -v .

deb:
	fpm -f -s dir -t deb --version 0.1.0 --license MIT -m madebyais@gmail.com --url https://github.com/madebyais/kintoun --vendor madebyais.com --description "FTP, FTPS and SFTP to HTTP REST file syncer" -n kintoun ~/Workspace/Engineering/src/github.com/madebyais/kintoun/kintoun_linux_amd64=/opt/kintoun

rpm:
	fpm -f -s dir -t rpm --version 0.1.0 --license MIT -m madebyais@gmail.com --url https://github.com/madebyais/kintoun --vendor madebyais.com --description "FTP, FTPS and SFTP to HTTP REST file syncer" -n kintoun ~/Workspace/Engineering/src/github.com/madebyais/kintoun=/opt/kintoun
