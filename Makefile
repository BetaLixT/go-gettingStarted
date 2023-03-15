run server:
	go run cmd/server/main.go
gen:
	go generate gen.go
setup:
	@prname='techunicorn.com/udc-core/gettingStarted';read -p "Enter Project Name ($$prname):" new; find . -type f  -not -path "./.git/*" -not -path "./pkg/app/server/static/swagger/*" -exec sed -i "s|$$prname|$$new|g" {} \;
	@prname='gettingStarted';read -p "Enter Service Name ($$prname):" new; find . -type f -not -path "./.git/*" -not -path "./pkg/app/server/static/swagger/*" -exec sed -i "s|$$prname|$$new|g" {} \;
