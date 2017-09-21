RELEASE=$(shell git describe --always --long --dirty)

clean:
	@rm -rf ./dist

build: clean
	@goxc -pv=$(version)-$(RELEASE)

version:
	@echo $(RELEASE)
