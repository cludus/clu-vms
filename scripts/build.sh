SCRIPT_DIR=$(dirname "$(realpath "$0")")
PARENT_DIR=$(dirname "$SCRIPT_DIR")
cd "$PARENT_DIR" || exit

currentVersion=$(<VERSION)
echo "building current version: $currentVersion"
go build -ldflags "-X main.Version=$currentVersion" -o ./bin/cvms ./cmd/cvms