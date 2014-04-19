build:
	( test -d dist || mkdir dist ) && cd dist && gox ../