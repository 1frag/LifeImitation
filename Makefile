start:
	go run controller.go manager.go server.go storage.go useful.go
.ONESHELL:
deploy:
	echo "Message: ";
	read msg;
	go build *.go
	git add .
	git commit -m "$$msg"
	git push heroku HEAD:master
logs:
	heroku logs --tail
restart:
	heroku restart
