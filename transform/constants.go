package transform

const (
	// GOT_BUILD_DIR is the directory where the generated go files are saved.
	GOT_BUILD_DIR = "got/"

	// GOT_METHODS_DIR is the directory where the generated methods are saved.
	GOT_METHODS_DIR = "methods/"

	// GOT_DECORATORS_DIR is the directory where generated decorators are saved.
	GOT_DECORATORS_DIR = "decorators/"

	// GOT_EXTRACT_DIR is the directory where extracted functions are saved.
	GOT_EXTRACT_DIR = "extracted/"

	// GOT_BUILD_FILE is the name of the generated go file.
	GOT_PREFIX = "#"

	// GOT_PREFIX_LEN is the length of the prefix.
	GOT_PREFIX_LEN = len(GOT_PREFIX)

	// GO_FILE_EXTENSION is the extension of the generated go file.
	GO_FILE_EXTENSION = ".go"

	// SINGLE_COMMENT is the prefix for a single line comment.
	SINGLE_COMMENT = "//"

	// SINGLE_COMMENT_PREFIX_LEN is the length of the single line comment prefix
	COMMENT_PREFIX_LEN = len(SINGLE_COMMENT)

	// GO_BUILD_COMMENT is the prefix for a build constraint.
	GO_BUILD_COMMENT = "//go:build"

	// GO_BUILD_COMMENT_LEN is the length of the build constraint prefix.
	GO_BUILD_COMMENT_LEN = len(GO_BUILD_COMMENT)
)
