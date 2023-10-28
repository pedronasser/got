package transform

const (
	GOT_BUILD_DIR        = "got/"
	GOT_METHODS_DIR      = "methods/"
	GOT_DECORATORS_DIR   = "decorators/"
	GOT_EXTRACT_DIR      = "extracted/"
	GOT_PREFIX           = "#"
	GOT_PREFIX_LEN       = len(GOT_PREFIX)
	GO_FILE_EXTENSION    = ".go"
	SINGLE_COMMENT       = "//"
	COMMENT_PREFIX_LEN   = len(SINGLE_COMMENT)
	GO_BUILD_COMMENT     = "//go:build"
	GO_BUILD_COMMENT_LEN = len(GO_BUILD_COMMENT)
)
