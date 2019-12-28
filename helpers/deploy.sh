read -r -p "Message: " msg
go build *.go
mv controller service
git add .
git commit -m "$msg"
git push heroku HEAD:master


