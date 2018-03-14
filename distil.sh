MOON="moonshine"

env GOOS=windows GOARCH=386 go build
mv "$MOON".exe "$MOON"-win386.exe
env GOOS=windows GOARCH=amd64 go build
mv "$MOON".exe "$MOON"-win64.exe
env GOOS=linux GOARCH=amd64 go build
mv "$MOON" "$MOON"-linux64
env GOOS=darwin GOARCH=386 go build
mv "$MOON" "$MOON"-darwin386
env GOOS=darwin GOARCH=amd64 go build
mv "$MOON" "$MOON"-darwinAmd64
