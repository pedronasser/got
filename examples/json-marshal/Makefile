test: 
	GOMEMLIMIT=2GiB GOGC=off ../../bin/got test -v -tags=generated -benchmem -bench=. . 
	GOMEMLIMIT=2GiB GOGC=off go test -v -bench=. -benchmem
