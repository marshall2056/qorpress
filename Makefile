#---* Makefile *---#

# To do
# git rev
build:
	@docker build -t x0rzkov/qor-example .

run:
	@docker run -ti -p 8080:8080 x0rzkov/qor-example
