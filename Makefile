clean:
	@rm -rf ./dist

build: clean
	@goxc -pv=$(version)

version:
	@echo $(RELEASE)
