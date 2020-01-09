start:
	go run *.go
.ONESHELL:
deploy:
	echo "Message: ";
	read msg;
	go build *.go
	mv controller service
	git add .
	git commit -m "$$msg"
	git push heroku HEAD:master
logs:
	heroku logs --tail
restart:
	heroku restart
