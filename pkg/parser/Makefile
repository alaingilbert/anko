all : parser.go

# Install goyacc using:
# go install modernc.org/goyacc@latest
parser.go : parser.go.y
	goyacc -o $@ parser.go.y
	gofmt -s -w .

.PHONY: parser.go