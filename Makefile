# Easily download all dependencies
deps:
	@cd cmd; go get -d -v; cd ..  

# Run the app locally, using memory storage exposing prometheus metrics
mem: deps
	@go run cmd/main.go --metrics=true --admin=true

# Build a new docker image
docker:
	@docker build -t marcoamador/go-payments-api:latest .

# Run all BDD scenarios
bdd:
	@cd test; godog; cd ..

# Run individual BDD scenarios
# This target looks for scenarios tagged @wip
bdd-wip:
	@cd test; godog --tags=wip; cd ..

# Start GoDoc's online documentation
# http://localhost:6060/pkg/github.com/mfamador/go-payments-api/
doc:
	@godoc -http=:6060
