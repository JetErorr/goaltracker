ooga booga goes here: 

Docs: 
https://hub.docker.com/_/mongo/
<!-- https://upper.io/v4/adapter/mongo/ -->


Go packages: 
go get -u github.com/gorilla/mux
go get -u github.com/gin-gonic/gin
<!-- go get github.com/upper/db/v4/adapter/mongodb -->
go get go.mongodb.org/mongo-driver
go get go.mongodb.org/mongo-driver/bson
go get github.com/golang-jwt/jwt



To run: 

`sudo docker compose up -d` launches the Dockerized MongoDB and mongo-express webui. 
`go run .` runs the goaltracker.go file, launching the API. 