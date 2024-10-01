app_name="dobby"

mkdir -p ~/.dobby/bin
mkdir -p ~/.dobby/scripts

GOOS=darwin GOARCH=amd64 go build -o ~/.dobby/bin/$app_name

cp ./scripts/register.sh ~/.dobby/scripts/register.sh
chmod +x ~/.dobby/scripts/register.sh

if ! grep -q "source ~/.dobby/scripts/register.sh" ~/.zshrc; then
  echo "source ~/.dobby/scripts/register.sh" >> ~/.zshrc
fi
